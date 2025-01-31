package helloasso

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"registro/config"
	"registro/sql/dons"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

const (
	dateLayoutHelloAsso = "2006-01-02T15:04:05-07:00" // format personnalisé imposé par HelloAsso
	maxTry              = 3
	retryDelayMax       = 10 // en seconds
	typeDon             = "DONATION"
	typeDonRecurrent    = "RECURRENT_DONATION"

	resultsPerPage = 1000000
)

// PingHelloAsso effectue une requête de test
// et renvoie l'éventuelle erreur.
func PingHelloAsso(creds config.Helloasso) error {
	_, err := getAccessToken(creds)
	return err
}

func parseDateHelloAsso(d string) (shared.Date, error) {
	date, err := time.Parse(dateLayoutHelloAsso, d)
	if err != nil {
		return shared.Date{}, fmt.Errorf("format HelloAsso incorrect: %s", err)
	}
	return shared.NewDateFrom(date), nil
}

type DonDonateur struct {
	Don      dons.Don
	Donateur pr.Etatcivil
}

func newDonFromPayment(payment paiementHelloAsso, affectation string) (DonDonateur, error) {
	date, err := parseDateHelloAsso(payment.Date)
	if err != nil {
		return DonDonateur{}, err
	}
	id := idV5ToV3(payment.Id)
	don := dons.Don{
		Valeur:       dossiers.Montant{Cent: payment.Amount, Currency: dossiers.Euros}, // TODO: currency
		Date:         date,
		ModePaiement: dossiers.Helloasso,
		Affectation:  affectation,
		Details:      id, // pourra être modifié

		IdPaiementHelloasso: id,
	}
	// HelloAsso utilise le code ISO à 3 lettres
	country, ok := paysCode3[payment.Payer.Country]
	if !ok {
		country = pr.Pays(payment.Payer.Country)
	}
	donateur := pr.Etatcivil{
		Prenom:     payment.Payer.FirstName,
		Nom:        payment.Payer.LastName,
		Adresse:    payment.Payer.Address,
		CodePostal: payment.Payer.ZipCode,
		Ville:      payment.Payer.City,
		Pays:       country,
		Mail:       payment.Payer.Email,
	}
	return DonDonateur{Don: don, Donateur: donateur}, nil
}

// TODO: use config
// les formulaire de dons sont identifiés par les slugs suivant
var formsHelloAsso = [...]formHelloAsso{
	{ // soutien: dons généraux, réguliers
		formSlug:    "1",
		formType:    "Donation",
		attribution: "",
	},
	{
		formSlug:    "soutien-a-l-accueil-d-enfants-refugies-a-montmeyran-2024",
		formType:    "CrowdFunding",
		attribution: "Montmeyran 2024",
	},
	{
		formSlug:    "soutien-chantier-humanitaire-en-bosnie-herzegovine-2024",
		formType:    "CrowdFunding",
		attribution: "Bosnie 2024",
	},
	{
		formSlug:    "soutien-au-spectacle-musical-de-menditte",
		formType:    "CrowdFunding",
		attribution: "Menditte 2024",
	},
}

type formHelloAsso struct {
	formType    string
	formSlug    string
	attribution string
}

// fetch and process the paiements for one HelloAsso form
func (form formHelloAsso) importDons(accesToken string, alreadyImported map[string]bool) ([]DonDonateur, error) {
	paiements, err := fetchAllFormPaiements(accesToken, form.formType, form.formSlug)
	if err != nil {
		return nil, err
	}

	var out []DonDonateur
	for _, paiement := range paiements {
		idV3 := idV5ToV3(paiement.Id)

		if alreadyImported[idV3] {
			continue
		}

		// only import proper paiements
		if paiement.State != "Authorized" {
			continue
		}

		item, err := newDonFromPayment(paiement, form.attribution)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// ImportDonsHelloasso cherche les nouveaux dons, utilisant [db]
// pour ignorer les dons déjà importés.
func ImportDonsHelloasso(creds config.Helloasso, db pr.DB) ([]DonDonateur, error) {
	accesToken, err := getAccessToken(creds)
	if err != nil {
		return nil, err
	}

	dons, err := dons.SelectAllDons(db)
	if err != nil {
		return nil, err
	}

	// on ne renvoie que les dons n'ayant pas encore été importés
	// TODO: utiliser une date pour éviter de tout chercher
	alreadyImported := map[string]bool{}
	for _, don := range dons {
		if don.ModePaiement == dossiers.Helloasso {
			alreadyImported[don.IdPaiementHelloasso] = true
		}
	}

	var out []DonDonateur
	for _, form := range formsHelloAsso {
		l, err := form.importDons(accesToken, alreadyImported)
		if err != nil {
			return nil, err
		}
		out = append(out, l...)
	}

	return out, nil
}

// -------------------------- API HelloAsso ----------------------------------------

type helloAssoError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func getAccessToken(pass config.Helloasso) (string, error) {
	params := url.Values{}
	params.Set("client_id", pass.ID)
	params.Set("client_secret", pass.Secret)
	params.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, "https://api.helloasso.com/oauth2/token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("La requête vers HelloAsso a échoué : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		var payload helloAssoError
		err = json.NewDecoder(resp.Body).Decode(&payload)
		if err != nil {
			return "", fmt.Errorf("La réponse d'HelloAsso est invalide : %s", err)
		}
		return "", fmt.Errorf("Erreur renvoyée par HelloAsso : %s", payload.ErrorDescription)
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("internal error : invalid HelloAsso response status %s", resp.Status)
	}

	var payload struct {
		AccessToken  string `json:"access_token"`  //	The JWT token to use in future requests
		RefreshToken string `json:"refresh_token"` //	Token used to refresh the token and get a new JWT token after expiration
		TokenType    string `json:"token_type"`    //	Token Type : always "bearer"
		ExpiresIn    int    `json:"expires_in"`    //	The lifetime in seconds of the access token
	}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return "", fmt.Errorf("La réponse d'HelloAsso est invalide : %s", err)
	}
	return payload.AccessToken, nil
}

func getJSON(url, accesToken string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("internal error in HelloAsso request : %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+accesToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("La requête vers HelloAsso a échoué : %s. Détail : %s", resp.Status, b)
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	return err
}

type pagination struct {
	TotalCount        int    `json:"totalCount"` // Total number of results available
	PageIndex         int    `json:"pageIndex"`  // Current page index
	TotalPages        int    `json:"totalPages"` // Total number of pages of results with current page size
	ContinuationToken string `json:"continuationToken"`
}

type paiementHelloAsso struct {
	Payer struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Address   string `json:"address"`
		ZipCode   string `json:"zipCode"`
		City      string `json:"city"`
		Country   string `json:"country"`
		Email     string `json:"email"`
	} `json:"payer"`
	Date   string `json:"date"`
	Id     int    `json:"id"`
	Amount int    `json:"amount"` // Montant en centimes
	Items  []struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
	} `json:"items"`
	State string `json:"state"`
}

func idV5ToV3(id int) string {
	return fmt.Sprintf("%011d3", id)
}

func fetchAllFormPaiements(accesToken string, formType, formSlug string) ([]paiementHelloAsso, error) {
	const resultsPerPage = 100 // maximum allowed

	var all []paiementHelloAsso
	var continuationToken string

	// start with empty continuation token
	for { // fetch all pages
		params := url.Values{}
		params.Set("pageSize", strconv.Itoa(resultsPerPage))
		if continuationToken != "" {
			params.Set("continuationToken", continuationToken)
		}

		url := fmt.Sprintf("https://api.helloasso.com/v5/organizations/acve/forms/%s/%s/payments?%s",
			formType, formSlug, params.Encode())
		var out struct {
			Data       []paiementHelloAsso `json:"data"`
			Pagination pagination          `json:"pagination"`
		}
		err := getJSON(url, accesToken, &out)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Data...)

		if out.Pagination.PageIndex == out.Pagination.TotalPages { // done
			break
		} else {
			continuationToken = out.Pagination.ContinuationToken
		}
	}

	return all, nil
}

// paysCode3 associe au code ISO à 3 lettres le pays (2 lettres).
var paysCode3 = map[string]pr.Pays{
	"BRB": "BB",
	"BEN": "BJ",
	"BRN": "BN",
	"CHN": "CN",
	"GTM": "GT",
	"JOR": "JO",
	"MNP": "MP",
	"PCN": "PN",
	"AUT": "AT",
	"JPN": "JP",
	"MAR": "MA",
	"NPL": "NP",
	"CCK": "CC",
	"DJI": "DJ",
	"NGA": "NG",
	"SMR": "SM",
	"MNE": "ME",
	"PER": "PE",
	"RUS": "RU",
	"BGD": "BD",
	"COD": "CD",
	"IRL": "IE",
	"LVA": "LV",
	"DMA": "DM",
	"OMN": "OM",
	"JAM": "JM",
	"PRT": "PT",
	"TON": "TO",
	"GNB": "GW",
	"GUY": "GY",
	"NRU": "NR",
	"SWE": "SE",
	"GGY": "GG",
	"ROU": "RO",
	"SJM": "SJ",
	"MRT": "MR",
	"HRV": "HR",
	"ETH": "ET",
	"GIB": "GI",
	"GRC": "GR",
	"COK": "CK",
	"QAT": "QA",
	"SEN": "SN",
	"SDN": "SD",
	"VIR": "VI",
	"BMU": "BM",
	"COL": "CO",
	"HUN": "HU",
	"MSR": "MS",
	"NCL": "NC",
	"VCT": "VC",
	"BOL": "BO",
	"BWA": "BW",
	"FJI": "FJ",
	"GRD": "GD",
	"NIC": "NI",
	"ASM": "AS",
	"BTN": "BT",
	"CHE": "CH",
	"WLF": "WF",
	"KHM": "KH",
	"GEO": "GE",
	"ZWE": "ZW",
	"FRA": "FR",
	"LIE": "LI",
	"NIU": "NU",
	"VGB": "VG",
	"REU": "RE",
	"SYR": "SY",
	"UMI": "UM",
	"AUS": "AU",
	"ISL": "IS",
	"IRN": "IR",
	"MKD": "MK",
	"KWT": "KW",
	"LKA": "LK",
	"AND": "AD",
	"ARM": "AM",
	"AZE": "AZ",
	"BDI": "BI",
	"ALA": "AX",
	"BGR": "BG",
	"HKG": "HK",
	"TKL": "TK",
	"MAF": "MF",
	"SVN": "SI",
	"SLB": "SB",
	"TUV": "TV",
	"AIA": "AI",
	"FIN": "FI",
	"MMR": "MM",
	"PSE": "PS",
	"KIR": "KI",
	"VUT": "VU",
	"CUB": "CU",
	"LSO": "LS",
	"PAN": "PA",
	"BHR": "BH",
	"LAO": "LA",
	"LBN": "LB",
	"LBY": "LY",
	"BLZ": "BZ",
	"BVT": "BV",
	"HMD": "HM",
	"HND": "HN",
	"DZA": "DZ",
	"DOM": "DO",
	"TWN": "TW",
	"CIV": "CI",
	"LTU": "LT",
	"NLD": "NL",
	"ATA": "AQ",
	"CPV": "CV",
	"CRI": "CR",
	"NOR": "NO",
	"FSM": "FM",
	"TLS": "TL",
	"BHS": "BS",
	"CUW": "CW",
	"GAB": "GA",
	"KGZ": "KG",
	"NAM": "NA",
	"ERI": "ER",
	"IDN": "ID",
	"MLT": "MT",
	"MUS": "MU",
	"GLP": "GP",
	"MAC": "MO",
	"MDG": "MG",
	"UZB": "UZ",
	"URY": "UY",
	"YEM": "YE",
	"ARG": "AR",
	"MYT": "YT",
	"STP": "ST",
	"GBR": "GB",
	"BES": "BQ",
	"GNQ": "GQ",
	"BRA": "BR",
	"SUR": "SR",
	"BIH": "BA",
	"ESH": "EH",
	"BLR": "BY",
	"CAN": "CA",
	"MOZ": "MZ",
	"SPM": "PM",
	"AGO": "AO",
	"MYS": "MY",
	"MCO": "MC",
	"ESP": "ES",
	"SVK": "SK",
	"TZA": "TZ",
	"TUR": "TR",
	"TCA": "TC",
	"ATF": "TF",
	"GUM": "GU",
	"MNG": "MN",
	"RWA": "RW",
	"MHL": "MH",
	"MDA": "MD",
	"UGA": "UG",
	"CMR": "CM",
	"VAT": "VA",
	"TUN": "TN",
	"ECU": "EC",
	"KEN": "KE",
	"SXM": "SX",
	"SGS": "GS",
	"ALB": "AL",
	"LUX": "LU",
	"KNA": "KN",
	"NER": "NE",
	"PAK": "PK",
	"SOM": "SO",
	"AFG": "AF",
	"TCD": "TD",
	"DEU": "DE",
	"MLI": "ML",
	"CHL": "CL",
	"PLW": "PW",
	"VEN": "VE",
	"SGP": "SG",
	"THA": "TH",
	"BFA": "BF",
	"COM": "KM",
	"IMN": "IM",
	"LCA": "LC",
	"TTO": "TT",
	"NFK": "NF",
	"PRY": "PY",
	"SHN": "SH",
	"SSD": "SS",
	"SLE": "SL",
	"CYP": "CY",
	"EGY": "EG",
	"MTQ": "MQ",
	"NZL": "NZ",
	"SYC": "SC",
	"TJK": "TJ",
	"PYF": "PF",
	"GMB": "GM",
	"IRQ": "IQ",
	"KAZ": "KZ",
	"IOT": "IO",
	"COG": "CG",
	"WSM": "WS",
	"UKR": "UA",
	"CYM": "KY",
	"CXR": "CX",
	"PNG": "PG",
	"BLM": "BL",
	"CZE": "CZ",
	"JEY": "JE",
	"ARE": "AE",
	"USA": "US",
	"TKM": "TM",
	"MWI": "MW",
	"PHL": "PH",
	"POL": "PL",
	"ZAF": "ZA",
	"SLV": "SV",
	"HTI": "HT",
	"KOR": "KR",
	"LBR": "LR",
	"ITA": "IT",
	"MEX": "MX",
	"ABW": "AW",
	"EST": "EE",
	"GIN": "GN",
	"SAU": "SA",
	"FRO": "FO",
	"GUF": "GF",
	"BEL": "BE",
	"PRK": "KP",
	"PRI": "PR",
	"CAF": "CF",
	"DNK": "DK",
	"FLK": "FK",
	"VNM": "VN",
	"ZMB": "ZM",
	"TGO": "TG",
	"SWZ": "SZ",
	"IND": "IN",
	"MDV": "MV",
	"SRB": "RS",
	"ATG": "AG",
	"GHA": "GH",
	"GRL": "GL",
	"ISR": "IL",
}

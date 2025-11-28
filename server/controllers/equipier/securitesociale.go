package equipier

import (
	"fmt"
	"registro/controllers/equipier/communes"
	pr "registro/sql/personnes"
	"strconv"
	"strings"
	"time"
)

type SecuriteSocialeCheck struct {
	Err                                    string // empty if valid
	DepartementNaissance, CommuneNaissance string // computed
}

func checkSecuriteSociale(sexe pr.Sexe, dateNaissance time.Time, securiteSociale string) SecuriteSocialeCheck {
	securiteSociale = strings.ToUpper(strings.ReplaceAll(securiteSociale, " ", ""))
	if len(securiteSociale) != 15 {
		return SecuriteSocialeCheck{Err: "Merci de renseigner les 15 chiffres."}
	}

	sexeC := securiteSociale[0]
	yearC := securiteSociale[1:3]
	monthC := securiteSociale[3:5]
	inseeC := securiteSociale[5:10]
	_ = securiteSociale[10:13] // numéro d'ordre
	controleC := securiteSociale[13:15]

	if (sexeC == '1' && sexe != pr.Man) || (sexeC == '2' && sexe != pr.Woman) {
		return SecuriteSocialeCheck{Err: "Le chiffre du sexe est invalide."}
	}
	year, err := strconv.Atoi(yearC)
	if err != nil {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("L'année de naissance est invalide (%s).", err)}
	}
	if dateNaissance.Year()%100 != year {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("L'année de naissance ne correspond pas (%d).", year)}
	}

	month, err := strconv.Atoi(monthC)
	if err != nil {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le mois de naissance est invalide (%s).", err)}
	}
	if (1 <= month && month <= 12) && int(dateNaissance.Month()) != month {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le mois de naissance ne correspond pas (%d).", month)}
	}

	departement, commune, err := communes.CommuneByCode(inseeC)
	if err != nil {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le code de la commune de naissance est invalide (%s).", err)}
	}

	// compute and check controle, handling Corse
	// https://fr.wikipedia.org/wiki/Numéro_de_sécurité_sociale_en_France#cite_note-F
	gotControle, err := strconv.Atoi(controleC)
	if err != nil {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le numéro de contrôle est invalide (%s).", err)}
	}
	bytes := []byte(securiteSociale[:13])
	is2A, is2B := communes.IsCorse(departement)
	if is2A {
		bytes[5], bytes[6] = '1', '9'
	} else if is2B {
		bytes[5], bytes[6] = '1', '8'
	}
	number, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le numéro est invalide (%s).", err)}
	}
	expControle := int(97 - number%97)
	if expControle != gotControle {
		return SecuriteSocialeCheck{Err: fmt.Sprintf("Le numéro de contrôle ne correspond pas (%d %d).", gotControle, expControle)}
	}

	return SecuriteSocialeCheck{DepartementNaissance: departement, CommuneNaissance: commune}
}

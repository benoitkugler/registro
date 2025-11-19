package config

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type Asso struct {
	ID string

	Title, Infos string // used in documents footer

	Icon                         string       // included as image src attribute
	ColorPrimary, ColorSecondary template.CSS // included as background-color

	// Used as default right header.
	// Also used in "Lettre aux parents" if UseCoordCentre is true, and as contact in espaceperso
	ContactNom, ContactTel, ContactMail string

	MailsSettings MailsSettings

	bankNames, bankIBANs []string       // displayed in espace perso
	ChequeSettings       ChequeSettings // nil for disabled

	RemisesHints RemisesHints

	ConfigInscription
}

func (a *Asso) BankAccounts() [][2]string {
	out := make([][2]string, len(a.bankNames))
	for i, name := range a.bankNames {
		iban := a.bankIBANs[i]
		out[i] = [2]string{name, iban}
	}
	return out
}

type ConfigInscription struct {
	SupportBonsCAF            bool // if true, displayed in inscription form and activates Aides on espaceperso
	SupportANCV               bool // if true, displayed in inscription form
	SupportPaiementEnLigne    bool // if true, displayed in inscription and activated on espaceperso
	EmailRetraitMedia         string
	ShowFondSoutien           bool // if true, displayed in inscription form
	ShowCharteConduite        bool // if true, displayed in inscription form
	AskNationnalite           bool // if true, displayed for participants in inscription form
	ShowInscriptionRapide     bool // if true, displays a bar in inscription form
	ShowAutorisationVehicules bool // if true, displays an autorisation checkbox in inscription form
	ShowAnnulationConditions  bool // if true, displays a warning in inscription form (step 3)
}

var acve = Asso{
	ID: "acve",

	Title: "ACVE",
	Infos: "Association loi 1901 - N° Siret: 781 875 851 00037 - Jeunesse et Sport : 026ORG0163",

	ContactNom:  "ACVE - Centre d'inscriptions",
	ContactTel:  "04 75 22 03 95",
	ContactMail: "inscriptions@acve.asso.fr",

	Icon:           "logo_acve.png",
	ColorPrimary:   "#b8dbf1", // rgb(184, 219, 241)
	ColorSecondary: "#feee00",

	MailsSettings: MailsSettings{
		AssoName:            "ACVE",
		Sauvegarde:          "acve@alwaysdata.net",
		Unsubscribe:         "contact@acve.asso.fr",
		ReplyTo:             "inscriptions@acve.asso.fr",
		SignatureMailCentre: "Pour le centre d'inscriptions, <br /> Marie-Pierre BUFFET",
	},

	bankNames: []string{"Crédit Mutuel (FR)", "Crédit mutual (CHF)"},

	ChequeSettings: ChequeSettings{
		IsValid: true,
		Ordre:   "ACVE",
		Adresse: [2]string{"Centre d'inscriptions - Marie-Pierre Buffet", "27, impasse Vignon - 26150 Chamaloc"},
	},

	RemisesHints: RemisesHints{
		ParentEquipier: 30,
		AutreInscrit:   15,
	},

	ConfigInscription: ConfigInscription{
		SupportBonsCAF: true, SupportANCV: true,
		SupportPaiementEnLigne:    true,
		EmailRetraitMedia:         "contact@acve.asso.fr",
		ShowFondSoutien:           false,
		ShowCharteConduite:        false,
		AskNationnalite:           false,
		ShowInscriptionRapide:     true,
		ShowAutorisationVehicules: true,
		ShowAnnulationConditions:  false,
	},
}

var repere = Asso{
	ID: "repere",

	Title: "Repère",
	Infos: "2025",

	ContactNom:  "Repère - Centre d'inscriptions",
	ContactTel:  "",
	ContactMail: "webmaster@lerepere.ch",

	Icon:           "logo_repere.png",
	ColorPrimary:   "#2b678a",
	ColorSecondary: "#2ebfdc",

	MailsSettings: MailsSettings{
		AssoName:            "Repère",
		Sauvegarde:          "",
		Unsubscribe:         "info@lerepere.ch",
		ReplyTo:             "info@lerepere.ch",
		SignatureMailCentre: "L'équipe Repère",
	},

	bankNames: []string{"Crédit Mutuel (EUR)", "Crédit mutual (CHF)"},

	ChequeSettings: ChequeSettings{}, // disabled

	RemisesHints: RemisesHints{
		ParentEquipier: 50,
		AutreInscrit:   10,
	},

	ConfigInscription: ConfigInscription{
		SupportBonsCAF: false, SupportANCV: false,
		SupportPaiementEnLigne:    false,
		EmailRetraitMedia:         "webmaster@lerepere.ch",
		ShowFondSoutien:           true,
		ShowCharteConduite:        true,
		AskNationnalite:           true,
		ShowInscriptionRapide:     false, // pour la première année
		ShowAutorisationVehicules: false,
		ShowAnnulationConditions:  true,
	},
}

type ChequeSettings struct {
	IsValid bool
	Ordre   string
	Adresse [2]string // nom - adresse
}

type RemisesHints struct {
	ParentEquipier int // in %
	AutreInscrit   int // in %
}

type MailsSettings struct {
	AssoName            string // used in adress and as object prefix
	Sauvegarde          string // send copies to this adress
	Unsubscribe         string // used in 'List-Unsubscribe' header
	ReplyTo             string // Adresse à laquelle répondre
	SignatureMailCentre template.HTML
}

// NewAsso read the ASSO env variable and returns the associated configuration.
// The following configs are supported:
//   - acve
//   - repere
//
// The ASSO_BANK_IBAN env is also read
func NewAsso() (Asso, error) {
	var out Asso

	switch asso := os.Getenv("ASSO"); asso {
	case "acve":
		out = acve
	case "repere":
		out = repere
	default:
		return Asso{}, fmt.Errorf("missing or unsupported ASSO env. (%s)", asso)
	}

	ibans := os.Getenv("ASSO_BANK_IBAN")
	if ibans == "" {
		return Asso{}, errors.New("missing ASSO_BANK_IBAN env. variable")
	}
	out.bankIBANs = strings.Split(ibans, ",")
	if len(out.bankIBANs) != len(out.bankNames) {
		return Asso{}, errors.New("inconsistent length in ASSO_BANK_IBAN")
	}

	return out, nil
}

package config

import (
	"errors"
	"fmt"
	"html/template"
	"os"
)

type Asso struct {
	Title, Infos string // used in documents footer

	Icon               string       // included as image src attribute
	Color, ColorAccent template.CSS // included as background-color

	ContactNom, ContactTel, ContactMail string

	MailsSettings MailsSettings

	BankName, BankIBAN string // displayed in espace perso
}

var acve = Asso{
	Title: "ACVE",
	Infos: "Association loi 1901 - N° Siret: 781 875 851 00037 - Jeunesse et Sport : 026ORG0163",

	ContactNom:  "ACVE - Centre d'inscriptions",
	ContactTel:  "04 75 22 03 95",
	ContactMail: "inscriptions@acve.asso.fr",

	Icon:        "assets/logo_acve.png",
	Color:       "#b8dbf1", // rgb(184, 219, 241)
	ColorAccent: "#feee00",

	MailsSettings: MailsSettings{
		AssoName:            "ACVE",
		Sauvegarde:          "acve@alwaysdata.net",
		Unsubscribe:         "contact@acve.asso.fr",
		ReplyTo:             "inscriptions@acve.asso.fr",
		SignatureMailDefaut: "Pour le centre d'inscriptions, <br /> Marie-Pierre BUFFET",
	},

	BankName: "Crédit Mutuel",
}

// TODO:
var repere = Asso{}

type MailsSettings struct {
	AssoName            string // used in adress
	Sauvegarde          string // send copies to this adress
	Unsubscribe         string // used in 'List-Unsubscribe' header
	ReplyTo             string // Adresse à laquelle répondre
	SignatureMailDefaut template.HTML
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

	iban := os.Getenv("ASSO_BANK_IBAN")
	if iban == "" {
		return Asso{}, errors.New("missing ASSO_BANK_IBAN env. variable")
	}
	out.BankIBAN = iban
	return out, nil
}

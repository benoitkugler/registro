package config

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
)

type Asso struct {
	Title, Infos string // used in documents footer

	Icon               string       // included as image src attribute
	Color, ColorAccent template.CSS // included as background-color

	// TODO: préciser l'utilisation
	ContactNom, ContactTel, ContactMail string

	MailsSettings MailsSettings

	bankNames, bankIBANs []string // displayed in espace perso

	SupportBonsCAF, SupportANCV bool // if true, displayed in inscription form
	EmailRetraitMedia           string
	ShowFondSoutien             bool // if true, displayed in inscription form
	ShowCharteConduite          bool // if true, displayed in inscription form
}

func (a *Asso) BankAccounts() [][2]string {
	out := make([][2]string, len(a.bankNames))
	for i, name := range a.bankNames {
		iban := a.bankIBANs[i]
		out[i] = [2]string{name, iban}
	}
	return out
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
		SignatureMailCentre: "Pour le centre d'inscriptions, <br /> Marie-Pierre BUFFET",
	},

	bankNames: []string{"Crédit Mutuel (FR)", "Crédit mutual (CHF)"},

	SupportBonsCAF: true, SupportANCV: true,
	EmailRetraitMedia:  "contact@acve.asso.fr",
	ShowFondSoutien:    false,
	ShowCharteConduite: false,
}

var repere = Asso{
	Title: "Repère",
	Infos: "Association Repère 2022"

	ContactNom:  "Repère - Centre d'inscriptions",
	ContactTel:  "",
	ContactMail: "info@lerepere.ch",

	Icon: "assets/logo_repere.png"
	Color:  "#2b678a",
	ColorAccent: "#2eaadc",

	MailsSettings: MailsSettings{
		AssoName:            "Repère",
		Sauvegarde:          "",
		Unsubscribe:         "info@lerepere.ch",
		ReplyTo:             "info@lerepere.ch",
		SignatureMailCentre: "L'équipe Repère",
	},

	bankNames: []string{"Crédit Mutuel (FR)", "Crédit mutual (CHF)"},

	SupportBonsCAF: false, SupportANCV: false,
	EmailRetraitMedia:  "webmaster@lerepere.ch",
	ShowFondSoutien:    true,
	ShowCharteConduite: true,
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

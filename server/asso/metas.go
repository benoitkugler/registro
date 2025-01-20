package asso

import "html/template"

type AssoMeta struct {
	Title, Infos template.HTML

	Icon  string       // included as image src attribute
	Color template.CSS // included as background-color

	ContactNom, ContactTel, ContactMail string
}

// TODO:
var Asso = AssoMeta{
	Title: "<b>ACVE</b>",
	Infos: "Association loi 1901 - N° Siret: 781 875 851 00037 - Jeunesse et Sport : 026ORG0163",

	ContactNom:  "ACVE - Centre d'inscriptions",
	ContactTel:  "04 75 22 03 95",
	ContactMail: "inscriptions@acve.asso.fr",

	Icon:  "assets/logo_acve.png",
	Color: "rgb(184, 219, 241)",
}

package recufiscal

// dateAsso est une date découpée en champs texte pour le formulaire Cerfa.
type dateAsso struct {
	Jour  string
	Mois  string
	Annee string
}

// association regroupe les informations de l'association émettrice
// nécessaires au remplissage du reçu fiscal.
type association struct {
	Nom                     string
	NumeroSIREN             string
	Adresse                 string
	CodePostal              string
	Ville                   string
	Pays                    string
	ObjectifLigne1          string
	ObjectifLigne2          string
	ReconnueUtilitePublique bool
	Decret                  dateAsso
	JournalOfficiel         dateAsso
}

// assoACVE contient les valeurs propres à l'ACVE.
// NumeroSIREN est le numéro SIREN ou RNA de l'association.
// TODO: À renseigner avec la valeur officielle.
var assoACVE = association{
	Nom:            "ACVE",
	NumeroSIREN:    "",
	Adresse:        "La Maison du Rocher",
	CodePostal:     "26160",
	Ville:          "CHAMALOC",
	Pays:           "France",
	ObjectifLigne1: "Créer et gérer des séjours pour enfants, adolescents et adultes.",
	ObjectifLigne2: "Faire connaître, à travers des animations adaptées à l’âge des participants, les valeurs chrétiennes.",

	ReconnueUtilitePublique: true,

	Decret: dateAsso{
		Jour:  "5",
		Mois:  "1",
		Annee: "1957",
	},
	JournalOfficiel: dateAsso{
		Jour:  "29",
		Mois:  "1",
		Annee: "1957",
	},
}

package inscriptions

import (
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type ResponsableLegal struct {
	Nom           string
	Prenom        string
	DateNaissance shared.Date
	Sexe          pr.Sexe
	Mail          string
	Tels          pr.Tels

	Adresse    string
	CodePostal string
	Ville      string
	Pays       pr.Pays
}

type OptIdPersonne shared.OptID[pr.IdPersonne]

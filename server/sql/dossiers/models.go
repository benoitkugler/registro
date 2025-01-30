package dossiers

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type (
	IdDossier  int64
	IdPaiement int64
)

// Dossier représente un dossier d'inscription validé,
// et permet le suivi de l'inscription.
// En particulier, il y a un espace personnel par dossier.
type Dossier struct {
	Id            IdDossier
	IdResponsable pr.IdPersonne // responsable légal en charge du dossier

	// CopiesMails est une liste d'adresse en copies des mails envoyés,
	// donnant entre autre accès à l'espace personnel
	CopiesMails pr.Mails

	LastConnection time.Time // connection sur l'espace personnel

	// IsValidated devient 'true' lorsque l'inscription
	// est validée manuellement par le centre ou un directeur.
	IsValidated bool

	// Autorisation de partage des adresses aux autres participants
	PartageAdressesOK bool
}

type Paiement struct {
	Id        IdPaiement
	IdDossier IdDossier `gomacro-sql-on-delete:"CASCADE"`

	IsAcompte       bool
	IsRemboursement bool

	Montant Montant
	Payeur  string
	Mode    ModePaiement
	Date    shared.Date
	// Label peut stocker le numéro du chèque ou la banque
	Label string
	// Details peut stocker un motif ou la date d'encaissement d'un chèque
	Details string
}

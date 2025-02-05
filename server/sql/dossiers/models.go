package dossiers

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type (
	IdTaux     int64
	IdDossier  int64
	IdPaiement int64
)

// Taux définit le taux de convertion de chaque
// monnaie.
//
// Une table de conversion est associée à chaque camp,
// et tous les camps d'un même dossier doivent être liés à la
// même table.
// Un camp sans [Taux] équivaut à la table {1000, 0}, c'est à dire
// avec un support pour les Euros seulement
//
// gomacro:SQL ADD UNIQUE(Label)
type Taux struct {
	Id IdTaux

	Label string

	// 1[Monnaie]  = [Field] / 1000 €

	Euros        int
	FrancsSuisse int
}

// Dossier représente un dossier d'inscription validé,
// et permet le suivi de l'inscription.
// En particulier, il y a un espace personnel par dossier.
//
// Requise par la contrainte Participant
// gomacro:SQL ADD UNIQUE(Id, IdTaux)
type Dossier struct {
	Id            IdDossier
	IdResponsable pr.IdPersonne // responsable légal en charge du dossier
	// IdTaux is used for consistency
	IdTaux IdTaux

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

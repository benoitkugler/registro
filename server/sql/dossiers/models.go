package dossiers

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	"registro/sql/camps"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type (
	IdParticipant int64
	IdDossier     int64
	IdSondage     int64
	IdPaiement    int64
)

type Participant struct {
	Id         IdParticipant
	IdCamp     camps.IdCamp  `gomacro-sql-on-delete:"CASCADE"`
	IdPersonne pr.IdPersonne `gomacro-sql-on-delete:"CASCADE"`
	IdDossier  IdDossier     `gomacro-sql-on-delete:"CASCADE"`

	ListeAttente     ListeAttente
	Remises          Remises
	QuotientFamilial int

	// TODO:
	// OptionPrix   OptionPrixParticipant `json:"option_prix"`

	Details string // rempli sur l'espace de suivi
	Bus     Bus    // rempli sur l'espace de suivi
}

// Dossier représente un dossier d'inscription validé,
// et permet le suivi de l'inscription.
// En particulier, il y a un espace personnel par dossier.
type Dossier struct {
	Id         IdDossier
	IdPersonne pr.IdPersonne // responsable légal en charge du dossier

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

	Montant camps.Montant
	Payeur  string
	Mode    ModePaiement
	Date    shared.Date
	// Label peut stocker le numéro du chèque ou la banque
	Label string
	// Details peut stocker un motif ou la date d'encaissement d'un chèque
	Details string
}

// Sondage enregistre les retours sur un séjour
//
// gomacro:SQL ADD UNIQUE(IdCamp, IdDossier)
type Sondage struct {
	IdSondage int64
	IdCamp    int64 `gomacro-sql-on-delete:"CASCADE"`
	IdDossier int64 `gomacro-sql-on-delete:"CASCADE"`
	Modified  time.Time

	ReponseSondage
}

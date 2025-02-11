package dossiers

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"fmt"
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
// Un taux par défaut est défini par {1000, 0}, c'est à dire
// avec un support pour les Euros seulement
//
// gomacro:SQL ADD UNIQUE(Label)
// gomacro:SQL ADD CHECK(Euros = 1000)
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
	IdTaux        IdTaux
	IdResponsable pr.IdPersonne // responsable légal en charge du dossier
	// IdTaux is used for consistency

	// CopiesMails est une liste d'adresse en copies des mails envoyés,
	// donnant entre autre accès à l'espace personnel
	CopiesMails pr.Mails
	// Autorisation de partage des adresses aux autres participants
	PartageAdressesOK bool

	// IsValidated devient 'true' lorsque l'inscription
	// est validée manuellement par le centre ou un directeur.
	IsValidated bool

	LastConnection time.Time // connection sur l'espace personnel

	KeyV1 string // Deprecated: for backward compatibility only
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

// DescriptionHTML renvoie une description et le montant, au format HTML
func (r Paiement) DescriptionHTML(taux Taux) (string, string) {
	var payeur string
	if r.IsAcompte {
		payeur = fmt.Sprintf("Acompte de <i>%s</i> au %s", r.Payeur, r.Date)
	} else if r.IsRemboursement {
		payeur = fmt.Sprintf("Remboursement au %s", r.Date)
	} else {
		payeur = fmt.Sprintf("Paiement de <i>%s</i> au %s", r.Payeur, r.Date)
	}
	m := r.Montant
	if r.IsRemboursement {
		m.Cent *= -1
	}
	montant := fmt.Sprintf("<i>%s</i>", taux.Convert(r.Montant).String())
	return payeur, montant
}

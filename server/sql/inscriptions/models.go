package inscriptions

import (
	"database/sql"
	"time"

	"registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type IdInscription int64

// Inscription enregistre l'inscription faite via le formulaire publique.
//
// L'inscription publique est transformée en [Dossier] dès réception,
// cette table ne sert donc qu'à garder une trace en cas de problème.
//
// gomacro:SQL ADD UNIQUE(Id, IdTaux)
type Inscription struct {
	Id IdInscription

	IdTaux dossiers.IdTaux // for consistency

	Responsable         ResponsableLegal
	ResponsablePreIdent OptIdPersonne `gomacro-sql-foreign:"Personne" gomacro-sql-on-delete:"SET NULL"`

	Message            string
	CopiesMails        pr.Mails
	PartageAdressesOK  bool
	DemandeFondSoutien bool

	DateHeure time.Time

	// IsConfirmed is set to 'true' when the
	// mail has been confirmed and the [Dossier] has been created.
	// Note that the [Dossier.IsValidated] is still false,
	// since some personnes may still require to be identified by humans
	IsConfirmed bool
}

// InscriptionParticipant
// gomacro:SQL ADD FOREIGN KEY (IdCamp, IdTaux) REFERENCES Camp (Id,IdTaux) ON DELETE CASCADE
// gomacro:SQL ADD FOREIGN KEY (IdInscription, IdTaux) REFERENCES Inscription (Id,IdTaux) ON DELETE CASCADE
type InscriptionParticipant struct {
	IdInscription IdInscription `gomacro-sql-on-delete:"CASCADE"`

	IdCamp camps.IdCamp `gomacro-sql-on-delete:"CASCADE"`

	IdTaux dossiers.IdTaux // for consistency

	// Optionel
	PreIdent OptIdPersonne `gomacro-sql-foreign:"Personne" gomacro-sql-on-delete:"SET NULL"`

	Nom           string
	Prenom        string
	DateNaissance shared.Date
	Sexe          pr.Sexe
	Nationnalite  pr.Nationnalite
}

// Create insert [insc], then update [participants] id's and insert them
func Create(tx *sql.Tx, insc Inscription, participants InscriptionParticipants) (Inscription, error) {
	insc, err := insc.Insert(tx)
	if err != nil {
		return insc, err
	}
	for i := range participants {
		participants[i].IdInscription = insc.Id
	}
	err = InsertManyInscriptionParticipants(tx, participants...)
	if err != nil {
		return insc, err
	}
	return insc, nil
}

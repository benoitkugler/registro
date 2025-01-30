package camps

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"database/sql"

	pr "registro/sql/personnes"
	"registro/sql/shared"
)

type (
	IdCamp        int64
	IdImagelettre int64
	IdEquipier    int64
)

type Camp struct {
	Id        IdCamp
	Nom       string
	DateDebut shared.Date
	Duree     int // nombre de jours date et fin inclus
	Agrement  string

	Prix Montant
}

func (cp *Camp) DateFin() shared.Date {
	return shared.Plage{From: cp.DateDebut, Duree: cp.Duree}.To()
}

// Lettredirecteur conserve le html utilisé pour générer la lettre.
// En revanche, c'est bien le document PDF généré et enregistré dans la
// table documents qui est envoyé aux parents.
//
// gomacro:SQL ADD UNIQUE(IdCamp)
type Lettredirecteur struct {
	IdCamp             int64 `gomacro-sql-on-delete:"CASCADE"`
	Html               string
	UseCoordCentre     bool
	ShowAdressePostale bool
	ColorCoord         string
}

// Imagelettre stockes les images contenues dans les lettres aux parents,
// accessibles via un lien crypté
type Imagelettre struct {
	Id       IdImagelettre
	IdCamp   int64  `gomacro-sql-on-delete:"CASCADE"`
	Filename string // as uploaded
	Content  []byte
}

// Equipier représente un participant dans l'équipe d'un séjour
//
// gomacro:SQL ADD UNIQUE(IdCamp, IdPersonne)
// gomacro:SQL CREATE UNIQUE INDEX ON Equipiers(IdCamp) WHERE #[Role.Direction] = ANY(Roles)
type Equipier struct {
	Id         IdEquipier
	IdCamp     IdCamp
	IdPersonne pr.IdPersonne `gomacro-sql-on-delete:"CASCADE"`

	Roles    Roles
	Presence OptionnalPlage

	Invitation InvitationEquipier
	// validation de la charte ACVE
	AccepteCharte sql.NullBool
}

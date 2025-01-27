package camps

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	pr "registro/sql/personnes"
)

type (
	IdCamp        int64
	IdImagelettre int64
)

type Camp struct {
	Id        IdCamp
	Nom       string
	DateDebut pr.Date
	Duree     int // nombre de jours date et fin inclus
	Agrement  string
}

func (cp *Camp) DateFin() pr.Date {
	out := cp.DateDebut.Time()
	return pr.NewDateFrom(out.Add(time.Hour * 24 * time.Duration(cp.Duree-1)))
}

// Lettredirecteur conserve le html utilisé pour générer la lettre.
// En revanche, c'est bien le document PDF généré et enregistré dans la 
// table documents qui est envoyé aux parents.
//
// gomacro:SQL ADD UNIQUE(IdCamp)
type Lettredirecteur struct {
	IdCamp             int64  `gomacro-sql-on-delete:"CASCADE"`
	Html               string 
	UseCoordCentre     bool   
	ShowAdressePostale bool   
	ColorCoord         string 
}

// Imagelettre stockes les images contenues dans les lettres aux parents,
// accessibles via un lien crypté
type Imagelettre struct {
	Id IdImagelettre
	IdCamp   int64  `gomacro-sql-on-delete:"CASCADE"`
	Filename string // as uploaded
	Content  []byte 
}

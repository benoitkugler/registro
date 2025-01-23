package camps

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"time"

	pr "registro/sql/personnes"
)

type (
	IdCamp        int64
	IdParticipant int64
)

type Camp struct {
	Id        IdCamp
	Nom       string
	DateDebut pr.Date
	Duree     int // nombre de jours date et fin inclus
	Agrement  string
}

func (cp *Camp) DateFin() pr.Date {
	out := time.Time(cp.DateDebut)
	return pr.Date(out.Add(time.Hour * 24 * time.Duration(cp.Duree-1)))
}

type Participant struct {
	Id         IdParticipant
	IdCamp     IdCamp
	IdPersonne pr.IdPersonne
}

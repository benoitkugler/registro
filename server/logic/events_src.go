package logic

import (
	"time"

	ds "registro/sql/dossiers"
	evs "registro/sql/events"
)

//go:generate ../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro events_src.go go/unions:events_gen.go

// Event exposes on event on the dossier track
type Event struct {
	Id        evs.IdEvent
	idDossier ds.IdDossier
	Created   time.Time
	Content   EventContent
}

// Events stores the [Event] for one [Dossier]
type Events []Event

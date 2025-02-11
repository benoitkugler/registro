package events

import (
	"testing"
	"time"

	"registro/sql/dossiers"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestEvents(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	_, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = dossiers.Dossier{IdTaux: 1, IdResponsable: 1}.Insert(db)

	event, err := Event{IdDossier: 1, Kind: Message, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventMessage{IdEvent: event.Id, Guard: Message}.Insert(db)
	tu.AssertNoErr(t, err)
}

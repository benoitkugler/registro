package events

import (
	"testing"
	"time"

	"registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestEvents(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	_, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ds.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ds.Dossier{IdTaux: 1, IdResponsable: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := camps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp2, err := camps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	event, err := Event{IdDossier: 1, Kind: Message, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventPlaceLiberee{IdEvent: event.Id}.Insert(db) // wrong type
	tu.AssertErr(t, err)
	err = EventMessage{IdEvent: event.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	event, err = Event{IdDossier: 1, Kind: Message, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventMessage{IdEvent: event.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = EventMessageVu{IdEvent: event.Id, IdCamp: camp1.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventMessageVu{IdEvent: event.Id, IdCamp: camp1.Id}.Insert(db)
	tu.AssertErr(t, err) // unique
	err = EventMessageVu{IdEvent: event.Id, IdCamp: camp2.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = camps.DeleteCampById(db, camp1.Id)
	tu.AssertNoErr(t, err) // cascade

	event, err = Event{IdDossier: 1, Kind: Attestation, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventAttestation{IdEvent: event.Id}.Insert(db)
	tu.AssertNoErr(t, err)
}

func TestSwitchDossier(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	_, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = ds.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	d1, err := ds.Dossier{IdTaux: 1, IdResponsable: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	d2, err := ds.Dossier{IdTaux: 1, IdResponsable: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := camps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	event, err := Event{IdDossier: d1.Id, Kind: Validation, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventValidation{IdEvent: event.Id, IdCamp: camp1.Id.Opt()}.Insert(db)
	tu.AssertNoErr(t, err)

	event, err = Event{IdDossier: d1.Id, Kind: Validation, Created: time.Now()}.Insert(db)
	tu.AssertNoErr(t, err)
	err = EventValidation{IdEvent: event.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = SwitchValidationAndMessageDossier(db, d1.Id, d2.Id)
	tu.AssertNoErr(t, err)
}

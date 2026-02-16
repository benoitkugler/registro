package backoffice

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

func createMessage(db events.DB, idDossier ds.IdDossier, origine events.Acteur, origineCamp events.OptIdCamp) error {
	_, _, err := events.CreateMessage(db, idDossier, time.Now(), events.EventMessage{Contenu: utils.RandString(30, true), Origine: origine, OrigineCamp: origineCamp})
	return err
}

func Test_events(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	pe1, err := pr.Personne{Identite: pr.Identite{DateNaissance: shared.Date(time.Now()), Prenom: "Benoit"}}.Insert(db)
	tu.AssertNoErr(t, err)
	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB, smtp: smtp, asso: asso}

	d1, err := ct.createDossier(pe1.Id)
	tu.AssertNoErr(t, err)

	t.Run("message CRUD", func(t *testing.T) {
		event, err := ct.sendMessage("localhost", EventsSendMessageIn{IdDossier: d1.Id, Contenu: `
		Merci pur l'inscriptio !
		
		Question : donnez moi vos bons cafs
		
		Merci
		Marie-Pierre
		`}, false)
		tu.AssertNoErr(t, err)
		tu.Assert(t, event.Kind == events.Message)

		err = ct.deleteEvent(event.Id)
		tu.AssertNoErr(t, err)

		_, err = ct.sendMessage("localhost", EventsSendMessageIn{IdDossier: d1.Id, Contenu: `.`}, true)
		tu.AssertNoErr(t, err)
	})

	t.Run("mark seen", func(t *testing.T) {
		err = createMessage(ct.db, d1.Id, events.Espaceperso, events.OptIdCamp{})
		tu.AssertNoErr(t, err)
		err = createMessage(ct.db, d1.Id, events.Backoffice, events.OptIdCamp{})
		tu.AssertNoErr(t, err)
		err = createMessage(ct.db, d1.Id, events.Directeur, camp1.Id.Opt())
		tu.AssertNoErr(t, err)

		err = ct.markMessagesSeen(d1.Id, false)
		tu.AssertNoErr(t, err)

		err = ct.markMessagesSeen(d1.Id, true)
		tu.AssertNoErr(t, err)
	})

	t.Run("facture", func(t *testing.T) {
		err = ct.sendFacture("", d1.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("documents & sondages", func(t *testing.T) {
		const toSend = 5
		var ids []ds.IdDossier
		for i := range [toSend + 3]int{} {
			pe, err := pr.Personne{}.Insert(ct.db)
			tu.AssertNoErr(t, err)
			dossier, err := ct.createDossier(pe.Id)
			tu.AssertNoErr(t, err)
			pa, err := ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier.Id, IdCamp: camp1.Id, IdPersonne: pe.Id})
			tu.AssertNoErr(t, err)

			if i < toSend {
				pa.Participant.Statut = cps.Inscrit
				err = ct.updateParticipant(pa.Participant)
				tu.AssertNoErr(t, err)

				ids = append(ids, dossier.Id)
			}

		}

		preview, err := ct.previewSendDocumentsCamp(camp1.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(preview.Dossiers) == toSend)

		it, err := ct.sendDocumentsCamp("", SendDocumentsCampIn{IdCamp: camp1.Id, IdDossiers: ids})
		tu.AssertNoErr(t, err)
		err = utils.StreamJSON(httptest.NewRecorder(), it)
		tu.AssertNoErr(t, err)

		it, err = ct.sendSondages("", camp1.Id)
		tu.AssertNoErr(t, err)
		err = utils.StreamJSON(httptest.NewRecorder(), it)
		tu.AssertNoErr(t, err)
	})

	t.Run("relance paiement", func(t *testing.T) {
		const toSend = 5
		var ids []ds.IdDossier
		for range [toSend]int{} {
			pe, err := pr.Personne{}.Insert(ct.db)
			tu.AssertNoErr(t, err)
			dossier, err := ct.createDossier(pe.Id)
			tu.AssertNoErr(t, err)
			_, err = ct.createParticipant(ParticipantsCreateIn{IdDossier: dossier.Id, IdCamp: camp1.Id, IdPersonne: pe.Id})
			tu.AssertNoErr(t, err)
			ids = append(ids, dossier.Id)
		}

		preview, err := ct.previewRelancePaiement(camp1.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(preview) == 0)

		out := httptest.NewRecorder()
		it, err := ct.sendRelancePaiement("", RelancePaiementIn{ids})
		tu.AssertNoErr(t, err)
		err = utils.StreamJSON(out, it)
		tu.AssertNoErr(t, err)
		tu.Assert(t, strings.Count(out.Body.String(), "\n") == 5)
	})
}

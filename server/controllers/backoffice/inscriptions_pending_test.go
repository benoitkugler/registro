package backoffice

import (
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	inAPI "registro/controllers/inscriptions"
	cps "registro/sql/camps"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestPending(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)

	ct := Controller{db: db.DB, asso: asso, smtp: smtp}
	l, err := ct.getPendingInscriptions()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l.Inscriptions) == 0)

	camp, err := cps.Camp{IdTaux: 1, Statut: cps.Ouvert, DateDebut: shared.NewDateFrom(time.Now()), Duree: 10}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	ctInsc := inAPI.NewController(ct.db, ct.key, ct.smtp, ct.asso)
	insc, parts, err := ctInsc.BuildInscription(inAPI.Inscription{
		Responsable: in.ResponsableLegal{
			Nom:           utils.RandString(10, false),
			Prenom:        utils.RandString(10, false),
			DateNaissance: shared.NewDate(1950, 1, 1),
		},
		Participants: []inAPI.Participant{{
			IdCamp:        camp.Id,
			Sexe:          pr.Man,
			DateNaissance: shared.NewDate(2000, 1, 1),
		}},
	})
	tu.AssertNoErr(t, err)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		insc, err = in.Create(tx, insc, parts)
		return err
	})
	tu.AssertNoErr(t, err)

	l, err = ct.getPendingInscriptions()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l.Inscriptions) == 1)

	err = ct.updatePendingInscription(UpdatePendingInscriptionIn{Id: insc.Id, Mail: "mdlsmd@free.fr"})
	tu.AssertNoErr(t, err)

	it, err := ct.relancePendingInscriptions("localhost", RelancePendingInscriptionsIn{Ids: []in.IdInscription{insc.Id}})
	tu.AssertNoErr(t, err)
	err = utils.StreamJSON(httptest.NewRecorder(), it)
	tu.AssertNoErr(t, err)

	err = ct.deletePendingInscription(insc.Id)
	tu.AssertNoErr(t, err)
}

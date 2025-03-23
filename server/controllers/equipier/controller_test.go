package equipier

import (
	"testing"

	cps "registro/sql/camps"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestUpdate(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1, Duree: 10}.Insert(db)
	tu.AssertNoErr(t, err)
	personne, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	eq, err := cps.Equipier{IdCamp: camp.Id, IdPersonne: personne.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	ct := Controller{db: db.DB}

	err = ct.update(eq.Id, pr.Etatcivil{}, cps.PresenceOffsets{})
	tu.AssertNoErr(t, err)
	err = ct.update(eq.Id, pr.Etatcivil{}, cps.PresenceOffsets{Debut: 4, Fin: -5})
	tu.AssertNoErr(t, err)
	err = ct.update(eq.Id, pr.Etatcivil{}, cps.PresenceOffsets{Debut: 5, Fin: -5})
	tu.AssertErr(t, err) // invalid with respect to duree
}

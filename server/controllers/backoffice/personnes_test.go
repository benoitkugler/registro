package backoffice

import (
	"testing"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestFichesanitaire(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	personne, err := pr.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	err = pr.Fichesanitaire{IdPersonne: personne.Id}.Insert(db)
	tu.AssertNoErr(t, err)
	ct := Controller{db: db.DB}

	err = ct.updateFichesanitaireAccess(UpdateFichesanitaireAccessIn{personne.Id, pr.Mails{}})
	tu.AssertErr(t, err)
	err = ct.updateFichesanitaireAccess(UpdateFichesanitaireAccessIn{personne.Id, pr.Mails{"test@free.fr", "autre@free.fr"}})
	tu.AssertNoErr(t, err)
	err = ct.updateFichesanitaireAccess(UpdateFichesanitaireAccessIn{personne.Id + 1, pr.Mails{"test@free.fr"}})
	tu.AssertErr(t, err)
}

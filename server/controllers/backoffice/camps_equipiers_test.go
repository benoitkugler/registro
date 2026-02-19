package backoffice

import (
	"testing"
	"time"

	"registro/config"
	"registro/crypto"
	cp "registro/sql/camps"
	cps "registro/sql/camps"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func TestEquipiers(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	pe1, err := pr.Personne{Identite: pr.Identite{Nom: "Banaza", Mail: "sdlsmd@dummy.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := pr.Personne{Identite: pr.Identite{Nom: "Kfdmlg"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe3, err := pr.Personne{Identite: pr.Identite{Nom: "Uruse", DateNaissance: shared.NewDateFrom(time.Now())}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, "", "", files.FileSystem{}, config.SMTP{}, config.Asso{}, config.Immich{}, config.Helloasso{})
	tu.AssertNoErr(t, err)

	camp, err := ct.createCamp("localhost")
	tu.AssertNoErr(t, err)

	_, err = ct.createEquipier(CreateEquipierIn{pe1.Id, camp.Camp.Camp.Id, cps.Roles{cp.Direction}})
	tu.AssertNoErr(t, err)

	_, err = ct.createEquipier(CreateEquipierIn{pe2.Id, camp.Camp.Camp.Id, cps.Roles{cp.Adjoint, cp.Menage}})
	tu.AssertNoErr(t, err)

	_, err = ct.createEquipier(CreateEquipierIn{pe3.Id, camp.Camp.Camp.Id, cps.Roles{cp.Chauffeur, cp.Cuisine}})
	tu.AssertNoErr(t, err)

	b, name, err := ct.exportListeEquipiers(camp.Camp.Camp.Id)
	tu.AssertNoErr(t, err)
	tu.Write(t, name, b)
}

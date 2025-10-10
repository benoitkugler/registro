package backoffice

import (
	"testing"

	"registro/config"
	"registro/crypto"
	"registro/joomeo"
	cps "registro/sql/camps"
	"registro/sql/files"
	"registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestJoomeo(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	joomeoCfg, err := config.NewJoomeo()
	tu.AssertNoErr(t, err)

	pe1, err := personnes.Personne{Etatcivil: personnes.Etatcivil{Mail: "bench26@gmail.com"}}.Insert(db)
	tu.AssertNoErr(t, err)
	pe2, err := personnes.Personne{Etatcivil: personnes.Etatcivil{Mail: "x.ben.x@free.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, "", "", files.FileSystem{}, config.SMTP{}, config.Asso{}, joomeoCfg, config.Helloasso{})
	tu.AssertNoErr(t, err)

	albums, err := ct.loadAlbums()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albums) == 0)

	camp1, err := ct.createCamp()
	tu.AssertNoErr(t, err)
	camp2, err := ct.createCamp()
	tu.AssertNoErr(t, err)
	_, err = ct.createCamp()
	tu.AssertNoErr(t, err)
	idCamp1, idCamp2 := camp1.Camp.Camp.Id, camp2.Camp.Camp.Id

	// create directeurs
	_, err = ct.createEquipier(CreateEquipierIn{IdPersonne: pe1.Id, IdCamp: idCamp1, Roles: cps.Roles{cps.Direction}})
	tu.AssertNoErr(t, err)
	_, err = ct.createEquipier(CreateEquipierIn{IdPersonne: pe2.Id, IdCamp: idCamp2, Roles: cps.Roles{cps.Direction}})
	tu.AssertNoErr(t, err)

	m, err := ct.createAlbums(CreateAlbumsIn{IdCamps: []cps.IdCamp{idCamp1, idCamp2}})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(m) == 2)

	albums, err = ct.loadAlbums()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albums) == 3)
	tu.Assert(t, albums[0].Album.Id != "")
	tu.Assert(t, albums[1].Album.Id != "")

	contacts, err := ct.addDirecteursToAlbums(AddDirecteursToAlbumsIn{[]cps.IdCamp{idCamp1, idCamp2}, false})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(contacts) == 2)
	perm1, perm2 := contacts[idCamp1], contacts[idCamp2]
	tu.Assert(t, perm1.AccesRules.AllowEditFileCaption && perm1.AccesRules.AllowDeleteFile)
	tu.Assert(t, perm2.AccesRules.AllowEditFileCaption && perm2.AccesRules.AllowDeleteFile)

	// cleanup
	api, err := joomeo.NewApi(joomeoCfg)
	tu.AssertNoErr(t, err)
	defer api.Close()
	err = api.DeleteAlbum(albums[0].Album.Id)
	tu.AssertNoErr(t, err)
	err = api.DeleteAlbum(albums[1].Album.Id)
	tu.AssertNoErr(t, err)
}

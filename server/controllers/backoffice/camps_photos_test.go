package backoffice

import (
	"testing"

	"registro/config"
	"registro/crypto"
	"registro/immich"
	cps "registro/sql/camps"
	"registro/sql/files"
	tu "registro/utils/testutils"
)

func TestPhotos(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	cfg, err := config.NewImmich()
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, crypto.Encrypter{}, "", "", files.FileSystem{}, config.SMTP{}, config.Asso{}, cfg, config.Helloasso{})
	tu.AssertNoErr(t, err)

	albums, err := ct.loadAlbums()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albums) == 0)

	camp1, err := ct.createCamp("localhost")
	tu.AssertNoErr(t, err)
	camp2, err := ct.createCamp("localhost")
	tu.AssertNoErr(t, err)
	_, err = ct.createCamp("localhost")
	tu.AssertNoErr(t, err)
	idCamp1, idCamp2 := camp1.Camp.Camp.Id, camp2.Camp.Camp.Id

	m, err := ct.createAlbums(CreateAlbumsIn{IdCamps: []cps.IdCamp{idCamp1, idCamp2}})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(m) == 2)

	albums, err = ct.loadAlbums()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albums) == 3)
	tu.Assert(t, albums[0].Album.Id != "")
	tu.Assert(t, albums[1].Album.Id != "")

	// cleanup
	api := immich.NewApi(cfg)
	err = api.DeleteAlbum(albums[0].Album.Id)
	tu.AssertNoErr(t, err)
	err = api.DeleteAlbum(albums[1].Album.Id)
	tu.AssertNoErr(t, err)
}

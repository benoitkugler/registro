package directeurs

import (
	"testing"

	"registro/config"
	"registro/immich"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	tu "registro/utils/testutils"
)

func TestPhotos(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)
	photos, err := config.NewImmich()
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, "", "", fs.FileSystem{}, smtp, asso, photos)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	data, err := ct.loadPhotos(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, !data.HasAlbum)

	api := immich.NewApi(ct.immich)

	album, err := api.CreateAlbum("__TEST")
	tu.AssertNoErr(t, err)
	defer api.DeleteAlbum(album.Id)

	camp.AlbumID = string(album.Id)
	_, err = camp.Update(ct.db)
	tu.AssertNoErr(t, err)

	data, err = ct.loadPhotos(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.HasAlbum)
	tu.Assert(t, data.Album.AlbumName == "__TEST")
	tu.Assert(t, data.Album.AssetCount == 0)

	_, err = ct.createEquipier("", EquipiersCreateIn{CreatePersonne: true, Mail: "dummy.fr"}, camp.Id)
	tu.AssertNoErr(t, err)

	ite, err := ct.sendMailInvitePhotos(camp.Id)
	tu.AssertNoErr(t, err)
	for _, err := range ite {
		tu.AssertNoErr(t, err)
	}
}

package directeurs

import (
	"testing"

	"registro/config"
	"registro/joomeo"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	tu "registro/utils/testutils"
)

func TestJoomeo(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	joomeoConfig, err := config.NewJoomeo()
	tu.AssertNoErr(t, err)

	ct, err := NewController(db.DB, "", "", fs.FileSystem{}, config.SMTP{}, config.Asso{}, joomeoConfig)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	data, err := ct.loadJoomeo(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.SpaceURL == "")

	api, err := joomeo.NewApi(ct.joomeo)
	tu.AssertNoErr(t, err)
	defer api.Close()

	sejoursFolder, err := api.GetSejoursFolder()
	tu.AssertNoErr(t, err)

	album, err := api.CreateAlbum(sejoursFolder, "__TEST")
	tu.AssertNoErr(t, err)
	camp.AlbumID = album.Id
	_, err = camp.Update(ct.db)
	tu.AssertNoErr(t, err)

	data, err = ct.loadJoomeo(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, data.SpaceURL != "")
	tu.Assert(t, data.Album.Label == "__TEST")
	tu.Assert(t, data.Album.FilesCount == 0)

	contacts, err := ct.invite(camp.Id, JoomeoInviteIn{
		Mails:    []string{"x.ben.x@free.fr", "bench26@gmail.com"},
		SendMail: false,
	})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(contacts) == 2)

	err = ct.removeContact(camp.Id, contacts[0].Id)
	tu.AssertNoErr(t, err)

	err = api.DeleteAlbum(album.Id)
	tu.AssertNoErr(t, err)
}

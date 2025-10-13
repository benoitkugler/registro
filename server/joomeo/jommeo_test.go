package joomeo

import (
	"testing"
	"time"

	"registro/config"
	"registro/utils"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Joomeo {
	tu.LoadEnv(t, "../env.sh")
	out, err := config.NewJoomeo()
	tu.AssertNoErr(t, err)
	return out
}

func TestConnexion(t *testing.T) {
	api, err := NewApi(devCreds(t))
	defer api.Close()
	tu.AssertNoErr(t, err)

	id, err := api.GetSejoursFolder()
	tu.AssertNoErr(t, err)
	tu.Assert(t, id != "")
}

func TestSejoursAlbums(t *testing.T) {
	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Close()

	root, err := api.GetSejoursFolder()
	tu.AssertNoErr(t, err)
	tu.Assert(t, root != "")

	a1, err := api.CreateAlbum(root, "__TEST  1")
	tu.AssertNoErr(t, err)
	tu.Assert(t, a1.Label == "__TEST  1")
	tu.Assert(t, a1.Date.Equal(time.Now().Truncate(time.Second)))
	tu.Assert(t, a1.FilesCount == 0)
	tu.Assert(t, len(a1.Contacts) == 0)

	a2, err := api.CreateAlbum(root, "__TEST 2")
	tu.AssertNoErr(t, err)

	ct, err := api.AddDirecteur(a1.Id, "Dummy@free.fr", false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, ct.Type == 1 && ct.AccesRules.AllowDeleteFile && ct.AccesRules.AllowEditFileCaption &&
		!ct.AccesRules.AllowCreateAlbum && !ct.AccesRules.AllowDeleteAlbum && !ct.AccesRules.AllowUpdateAlbum)

	// make sure the permissions are properly returned even if the contact already exists
	ct, err = api.AddDirecteur(a1.Id, "Dummy@free.fr", false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, ct.Type == 1 && ct.AccesRules.AllowDeleteFile && ct.AccesRules.AllowEditFileCaption &&
		!ct.AccesRules.AllowCreateAlbum && !ct.AccesRules.AllowDeleteAlbum && !ct.AccesRules.AllowUpdateAlbum)

	// This email is a valid "trash" adress
	ct2, err := api.AddDirecteur(a2.Id, "bench26@gmail.com", true)
	tu.Assert(t, ct2.Type == 1 && ct2.AccesRules.AllowDeleteFile && ct2.AccesRules.AllowEditFileCaption &&
		!ct2.AccesRules.AllowCreateAlbum && !ct2.AccesRules.AllowDeleteAlbum && !ct2.AccesRules.AllowUpdateAlbum)
	tu.AssertNoErr(t, err)

	l, err := api.LoadAlbums(utils.NewSet(a1.Id, a2.Id))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) >= 2)

	a1 = l[a1.Id]
	dir, ok := a1.FindContact("Dummy@free.fr")
	tu.Assert(t, ok)
	tu.Assert(t, dir.Contact == ct.Contact)

	otherA1, err := api.LoadAlbum(a1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, otherA1.Label == a1.Label)

	err = api.deleteContact(dir.Id)
	tu.AssertNoErr(t, err)

	err = api.DeleteAlbum(a1.Id)
	tu.AssertNoErr(t, err)
	err = api.DeleteAlbum(a2.Id)
	tu.AssertNoErr(t, err)
}

func TestGetLoginFromMail(t *testing.T) {
	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Close()

	root, err := api.GetSejoursFolder()
	tu.AssertNoErr(t, err)

	a1, err := api.CreateAlbum(root, "__TEST  1")
	tu.AssertNoErr(t, err)

	a2, err := api.CreateAlbum(root, "__TEST 2")
	tu.AssertNoErr(t, err)

	ct, err := api.AddDirecteur(a1.Id, "Dummy@free.fr", false)
	tu.AssertNoErr(t, err)
	_, err = api.AddDirecteur(a2.Id, "Dummy@free.fr", false)
	tu.AssertNoErr(t, err)

	contact, albums, err := api.GetContactByMail("Dummy@free.fr")
	tu.AssertNoErr(t, err)
	tu.Assert(t, contact.Id == ct.Id)
	tu.Assert(t, len(albums) == 2)

	err = api.deleteContact(ct.Id)
	tu.AssertNoErr(t, err)
	err = api.DeleteAlbum(a1.Id)
	tu.AssertNoErr(t, err)
	err = api.DeleteAlbum(a2.Id)
	tu.AssertNoErr(t, err)
}

func TestAddContacts(t *testing.T) {
	api, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Close()

	root, err := api.GetSejoursFolder()
	tu.AssertNoErr(t, err)

	a1, err := api.CreateAlbum(root, "__TEST  1")
	tu.AssertNoErr(t, err)

	err = api.AddContacts("C2 2019", a1.Id, []string{"x.ben.x@free.fr", "bench26@gmail.com"}, false)
	tu.AssertNoErr(t, err)

	err = api.AddContacts("C2 2019", a1.Id, []string{"x.ben.x@free.fr", "bench26@gmail.com"}, true)
	tu.AssertNoErr(t, err)

	l, err := api.loadContactsFor(a1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 2)

	perms, err := api.SetContactUploader(a1.Id, l[0].Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, perms.AccesRules.AllowDeleteFile && perms.AlbumPermissions.AllowUpload)

	err = api.UnlinkContact(a1.Id, l[0].Id)
	tu.AssertNoErr(t, err)

	l, err = api.loadContactsFor(a1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l) == 1)

	err = api.DeleteAlbum(a1.Id)
	tu.AssertNoErr(t, err)
}

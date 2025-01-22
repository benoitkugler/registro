package joomeo

import (
	"fmt"
	"testing"
	"time"

	"registro/config"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Joomeo {
	tu.LoadEnv(t, "../dev.env")
	out, err := config.NewJoomeo()
	tu.AssertNoErr(t, err)
	return out
}

// valid in dev mode
const albumidTest = "QVlJS3ZXTjE-MhfjH29qFA"

func TestConnexion(t *testing.T) {
	api, err := InitApi(devCreds(t))
	defer api.Kill()
	tu.AssertNoErr(t, err)
}

func TestContacts(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	l, err := api.getContacts()
	tu.AssertNoErr(t, err)
	fmt.Println(l)
}

func TestFolders(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	l, err := api.getFolders("")
	tu.AssertNoErr(t, err)
	fmt.Println(l)

	lchildren, err := api.getFolders(l[0].FolderId)
	tu.AssertNoErr(t, err)
	fmt.Println(lchildren)

	l2, err := api.getAlbums("")
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l2) != 0)

	fmt.Println(l2[0].Date.date())
	fmt.Println(l2[0].FolderId)
	fmt.Println(l2)
}

func TestGetAlbumsContacts(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	m1, m2, m3, err := api.GetAllAlbumsContacts()
	tu.AssertNoErr(t, err)
	fmt.Println(m1)
	fmt.Println(m2)
	fmt.Println("Nombre de contacts", len(m3))
}

func TestAjouteDirecteur(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	c, err := api.AjouteDirecteur(albumidTest, "x.ben.x@free.fr", true)
	tu.AssertNoErr(t, err)
	fmt.Println(c)
	c, err = api.AjouteDirecteur(albumidTest, "bench26@gmail.com", false)
	tu.AssertNoErr(t, err)
	fmt.Println(c)
}

func TestAjouteContacts(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	c, err := api.AjouteContacts("C2", 2019, albumidTest, []string{"x.ben.x@free.fr", "benoit.kugler@inria.fr"}, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(c) == 0)

	c, err = api.AjouteContacts("C2", 2019, albumidTest, []string{"x.ben.x@free.fr", "bench26@gmail.com"}, true)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(c) == 0)

	l, err := api.GetContactsFor(albumidTest)
	tu.AssertNoErr(t, err)
	L1 := len(l)
	tu.Assert(t, L1 >= 3)

	err = api.RemoveContact(albumidTest, l[0].ContactId)
	tu.AssertNoErr(t, err)
	l, err = api.GetContactsFor(albumidTest)
	tu.AssertNoErr(t, err)
	L2 := len(l)
	tu.Assert(t, L2 == L1-1)
}

func TestSetUploader(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	_, err = api.AjouteContacts("C2", 2019, albumidTest, []string{"x.ben.x@free.fr"}, false)
	tu.AssertNoErr(t, err)
	l, err := api.GetContactsFor(albumidTest)
	tu.AssertNoErr(t, err)

	err = api.SetContactUploader(albumidTest, l[0].ContactId)
	tu.AssertNoErr(t, err)
}

func TestGetFromMail(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	contact, albums, err := api.GetLoginFromMail("x.ben.x@free.fr")
	tu.AssertNoErr(t, err)
	tu.Assert(t, contact.ContactId != "")

	fmt.Println(contact)
	fmt.Println(albums)
}

func TestGetMetadatas(t *testing.T) {
	api, err := InitApi(devCreds(t))
	tu.AssertNoErr(t, err)
	defer api.Kill()

	alb, err := api.GetAlbumMetadatas(albumidTest)
	tu.AssertNoErr(t, err)
	tu.Assert(t, alb.NbFiles == 5)
	tu.Assert(t, alb.Date.date().Year() == 2019)
	tu.Assert(t, alb.Date.date().Month() == time.August)
	tu.Assert(t, alb.Date.date().Day() == 11)
}

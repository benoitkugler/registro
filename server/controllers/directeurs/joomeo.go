package directeurs

import (
	"errors"
	"slices"

	"registro/joomeo"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Joomeo struct {
	SpaceURL string // empty if no album is link to the camp
	Album    joomeo.Album

	MailsResponsables []string
	MailsInscrits     []string
	MailsEquipiers    []string
}

func (ct *Controller) JoomeoLoad(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.loadJoomeo(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func mailsFor(m pr.Personnes) []string {
	tmp := make(utils.Set[string])
	for _, p := range m {
		tmp.Add(p.Mail)
	}
	out := tmp.Keys()
	slices.Sort(out)
	return out
}

func (ct *Controller) loadJoomeo(id cps.IdCamp) (data Joomeo, err error) {
	camp, err := cps.LoadCampPersonnes(ct.db, id)
	if err != nil {
		return Joomeo{}, err
	}

	if camp.Camp.JoomeoID == "" {
		return Joomeo{}, nil
	}

	// load mails
	inscrits := camp.Personnes(true)

	dossiers, err := dossiers.SelectDossiers(ct.db, camp.IdDossiers()...)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}
	responsables, err := pr.SelectPersonnes(ct.db, dossiers.IdResponsables()...)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}

	equipiersL, err := cps.SelectEquipiersByIdCamps(ct.db, id)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}
	equipiers, err := pr.SelectPersonnes(ct.db, equipiersL.IdPersonnes()...)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}

	camp.IdDossiers()

	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return Joomeo{}, err
	}
	defer api.Close()

	album, err := api.LoadAlbum(camp.Camp.JoomeoID)
	if err != nil {
		return Joomeo{}, err
	}

	return Joomeo{api.SpaceURL(), album, mailsFor(responsables), mailsFor(inscrits), mailsFor(equipiers)}, nil
}

type JoomeoInviteIn struct {
	Mails    []string
	SendMail bool
}

func (ct *Controller) JoomeoInvite(c echo.Context) error {
	user := JWTUser(c)
	var args JoomeoInviteIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.invite(user, args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) invite(idCamp cps.IdCamp, args JoomeoInviteIn) ([]joomeo.ContactPermission, error) {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	if camp.JoomeoID == "" {
		return nil, errors.New("internal error: no Joomeo album")
	}
	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return nil, err
	}
	defer api.Close()

	err = api.AddContacts(camp.Label(), camp.JoomeoID, args.Mails, args.SendMail)
	if err != nil {
		return nil, err
	}

	album, err := api.LoadAlbum(camp.JoomeoID)
	if err != nil {
		return nil, err
	}

	return album.Contacts, nil
}

func (ct *Controller) JoomeoUnlinkContact(c echo.Context) error {
	user := JWTUser(c)
	joomeoId := c.QueryParam("joomeoId")

	err := ct.removeContact(user, joomeoId)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) removeContact(idCamp cps.IdCamp, joomeoId string) error {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return utils.SQLError(err)
	}

	if camp.JoomeoID == "" {
		return errors.New("internal error: no Joomeo album")
	}
	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return err
	}
	defer api.Close()

	err = api.UnlinkContact(camp.JoomeoID, joomeoId)
	return err
}

func (ct *Controller) JoomeoSetUploader(c echo.Context) error {
	user := JWTUser(c)
	joomeoId := c.QueryParam("joomeoId")
	out, err := ct.setUploader(user, joomeoId)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) setUploader(idCamp cps.IdCamp, contactId string) (joomeo.ContactPermission, error) {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return joomeo.ContactPermission{}, utils.SQLError(err)
	}

	if camp.JoomeoID == "" {
		return joomeo.ContactPermission{}, errors.New("internal error: no Joomeo album")
	}
	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return joomeo.ContactPermission{}, err
	}
	defer api.Close()

	return api.SetContactUploader(camp.JoomeoID, contactId)
}

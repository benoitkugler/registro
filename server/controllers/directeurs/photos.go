package directeurs

import (
	"errors"
	"iter"
	"slices"

	"registro/controllers/backoffice"
	"registro/immich"
	"registro/mails"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Photos struct {
	HasAlbum bool // false if no album is link to the camp
	Album    immich.AlbumAndLinks
}

func (ct *Controller) PhotosLoad(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.loadPhotos(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadPhotos(id cps.IdCamp) (data Photos, err error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return Photos{}, err
	}

	if camp.AlbumID == "" {
		return Photos{}, nil
	}

	api := immich.NewApi(ct.immich)

	album, err := api.LoadAlbum(immich.AlbumID(camp.AlbumID))
	if err != nil {
		return Photos{}, err
	}

	return Photos{HasAlbum: true, Album: album}, nil
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

// PhotosInvite send mails to the responsables (with read only link)
// and to the  equipiers (with upload link).
func (ct *Controller) PhotosInvite(c echo.Context) error {
	user := JWTUser(c)

	it, err := ct.sendMailInvitePhotos(user)
	if err != nil {
		return err
	}

	return utils.StreamJSON(c.Response(), it)
}

func (ct *Controller) sendMailInvitePhotos(idCamp cps.IdCamp) (iter.Seq2[backoffice.SendProgress, error], error) {
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	if camp.Camp.AlbumID == "" {
		return nil, errors.New("internal error: no Photos album")
	}

	// load mails
	dossiers, err := dossiers.SelectDossiers(ct.db, camp.IdDossiers()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	responsables, err := pr.SelectPersonnes(ct.db, dossiers.IdResponsables()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	_, equipiers, err := cps.LoadEquipiersByCamps(ct.db, idCamp)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	api := immich.NewApi(ct.immich)

	album, err := api.LoadAlbum(immich.AlbumID(camp.Camp.AlbumID))
	if err != nil {
		return nil, err
	}

	campLabel := camp.Camp.Label()

	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return nil, err
	}

	total := len(responsables) + len(equipiers)
	return func(yield func(backoffice.SendProgress, error) bool) {
		defer pool.Close()

		current := 1

		for _, responsable := range responsables {
			html, err := mails.SendPhotosLinkInscrits(ct.asso, mails.NewContact(&responsable), campLabel, album.InscritsURL)
			if err != nil {
				yield(backoffice.SendProgress{}, err)
				return
			}

			err = pool.SendMail(responsable.Mail, "Album photos", html, nil, nil)
			if !yield(backoffice.SendProgress{Current: current, Total: total}, err) {
				return
			}
			current++
		}

		for _, equipier := range equipiers {
			html, err := mails.SendPhotosLinkEquipiers(ct.asso, mails.NewContact(&equipier), campLabel, album.EquipiersURL)
			if err != nil {
				yield(backoffice.SendProgress{}, err)
				return
			}

			err = pool.SendMail(equipier.Mail, "Album photos", html, nil, nil)
			if !yield(backoffice.SendProgress{Current: current, Total: total}, err) {
				return
			}
			current++
		}
	}, nil
}

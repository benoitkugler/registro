package backoffice

import (
	"fmt"
	"slices"

	"registro/joomeo"
	cps "registro/sql/camps"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func loadDirecteurs(db cps.DB, camps []cps.IdCamp) (map[cps.IdCamp]pr.Personne, error) {
	tmp, personnes, err := cps.LoadEquipiersByCamps(db, camps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	equipiersByCamp := tmp.ByIdCamp()

	out := map[cps.IdCamp]pr.Personne{}
	for _, camp := range camps {
		dir, ok := equipiersByCamp[camp].Directeur()
		if ok {
			out[camp] = personnes[dir.IdPersonne]
		}
	}
	return out, nil
}

// CampsLoadAlbums charge la liste des albums Joomeo.
func (ct *Controller) CampsLoadAlbums(c echo.Context) error {
	out, err := ct.loadAlbums()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type CampJoomeo struct {
	Camp cps.Camp

	HasDirecteur bool // true if the camp has an [Equpier] with role [Direction]

	Album               joomeo.Album // may be empty
	DirecteurPermission joomeo.ContactPermission
}

func (ct *Controller) loadAlbums() ([]CampJoomeo, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// restrict to not terminated
	ids := utils.NewSet[joomeo.AlbumId]()
	for id, camp := range camps {
		if camp.Ext().IsTerminated {
			delete(camps, id)
			continue
		}
		if camp.JoomeoID != "" {
			ids.Add(camp.JoomeoID)
		}
	}

	directeurs, err := loadDirecteurs(ct.db, camps.IDs())
	if err != nil {
		return nil, err
	}

	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return nil, err
	}
	defer api.Close()

	albums, err := api.LoadAlbums(ids)
	if err != nil {
		return nil, err
	}

	out := make([]CampJoomeo, 0, len(camps))
	for _, camp := range camps {
		album := albums[camp.JoomeoID]

		var perm joomeo.ContactPermission

		dir, hasDirecteur := directeurs[camp.Id]
		if hasDirecteur {
			// joomeo contact are identified by Mail
			mail := dir.Mail
			perm, _ = album.FindContact(mail)
		}

		out = append(out, CampJoomeo{camp, hasDirecteur, album, perm})
	}

	slices.SortFunc(out, func(a, b CampJoomeo) int { return int(a.Camp.Id - b.Camp.Id) })
	slices.SortStableFunc(out, func(a, b CampJoomeo) int { return a.Camp.DateDebut.Time().Compare(b.Camp.DateDebut.Time()) })

	return out, nil
}

type CreateAlbumsIn struct {
	IdCamps []cps.IdCamp
}

func (ct *Controller) CampsCreateAlbums(c echo.Context) error {
	var args CreateAlbumsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createAlbums(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createAlbums(args CreateAlbumsIn) (map[cps.IdCamp]joomeo.Album, error) {
	camps, err := cps.SelectCamps(ct.db, args.IdCamps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// safety check : disallow duplicate album
	for _, camp := range camps {
		if camp.JoomeoID != "" {
			return nil, fmt.Errorf("internal error: %s already has a Joomeo album", camp.Label())
		}
	}

	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return nil, err
	}
	defer api.Close()

	sejoursFolder, err := api.GetSejoursFolder()
	if err != nil {
		return nil, err
	}

	out := make(map[cps.IdCamp]joomeo.Album)
	for _, camp := range camps {
		album, err := api.CreateAlbum(sejoursFolder, camp.Label())
		if err != nil {
			return nil, err
		}
		out[camp.Id] = album

		// register the ID
		camp.JoomeoID = album.Id
		_, err = camp.Update(ct.db)
		if err != nil {
			return nil, utils.SQLError(err)
		}
	}

	return out, nil
}

type AddDirecteursToAlbumsIn struct {
	IdCamps  []cps.IdCamp
	SendMail bool
}

func (ct *Controller) CampsAddDirecteursToAlbums(c echo.Context) error {
	var args AddDirecteursToAlbumsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.addDirecteursToAlbums(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) addDirecteursToAlbums(args AddDirecteursToAlbumsIn) (map[cps.IdCamp]joomeo.ContactPermission, error) {
	camps, err := cps.SelectCamps(ct.db, args.IdCamps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	directeurs, err := loadDirecteurs(ct.db, camps.IDs())
	if err != nil {
		return nil, err
	}

	// safety check first
	for _, camp := range camps {
		if camp.JoomeoID == "" {
			return nil, fmt.Errorf("internal error: %s does not have a Joomeo album", camp.Label())
		}
		if _, hasDirecteur := directeurs[camp.Id]; !hasDirecteur {
			return nil, fmt.Errorf("Le séjour %s n'a pas de directeur.", camp.Label())
		}
	}

	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return nil, err
	}
	defer api.Close()

	out := make(map[cps.IdCamp]joomeo.ContactPermission)
	for _, camp := range camps {
		directeur := directeurs[camp.Id]
		contact, err := api.AddDirecteur(camp.JoomeoID, directeur.Mail, args.SendMail)
		if err != nil {
			return nil, err
		}
		out[camp.Id] = contact
	}

	return out, nil
}

package backoffice

import (
	"fmt"
	"slices"

	"registro/immich"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// CampsLoadAlbums charge la liste des albums.
func (ct *Controller) CampsLoadAlbums(c echo.Context) error {
	out, err := ct.loadAlbums()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type CampAlbum struct {
	Camp cps.Camp

	Album immich.AlbumAndLinks // may be empty
}

func (ct *Controller) loadAlbums() ([]CampAlbum, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// restrict to not terminated
	ids := utils.NewSet[immich.AlbumID]()
	for id, camp := range camps {
		if camp.Ext().IsTerminated {
			delete(camps, id)
			continue
		}

		if camp.AlbumID != "" {
			ids.Add(immich.AlbumID(camp.AlbumID))
		}
	}

	api := immich.NewApi(ct.immich)

	albums, err := api.LoadAlbums(ids.Keys())
	if err != nil {
		return nil, err
	}

	out := make([]CampAlbum, 0, len(camps))
	for _, camp := range camps {
		album := albums[immich.AlbumID(camp.AlbumID)]
		out = append(out, CampAlbum{camp, album})
	}

	slices.SortFunc(out, func(a, b CampAlbum) int { return int(a.Camp.Id - b.Camp.Id) })
	slices.SortStableFunc(out, func(a, b CampAlbum) int { return a.Camp.DateDebut.Time().Compare(b.Camp.DateDebut.Time()) })

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

func (ct *Controller) createAlbums(args CreateAlbumsIn) (map[cps.IdCamp]immich.AlbumAndLinks, error) {
	camps, err := cps.SelectCamps(ct.db, args.IdCamps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// safety check : disallow duplicate album
	for _, camp := range camps {
		if camp.AlbumID != "" {
			return nil, fmt.Errorf("internal error: %s already has a photo album", camp.Label())
		}
	}

	api := immich.NewApi(ct.immich)

	out := make(map[cps.IdCamp]immich.AlbumAndLinks)
	for _, camp := range camps {
		album, err := api.CreateAlbum(camp.Label())
		if err != nil {
			return nil, err
		}
		out[camp.Id] = album

		// register the ID
		camp.AlbumID = string(album.Id)
		_, err = camp.Update(ct.db)
		if err != nil {
			return nil, utils.SQLError(err)
		}
	}

	return out, nil
}

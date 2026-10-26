package backoffice

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"

	"registro/immich"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// CampsLoadAlbums charge la liste des albums.
func (ct *Controller) CampsLoadAlbums(c echo.Context) error {
	albums, err := ct.loadAlbums()
	if err != nil {
		return err
	}
	out := CampPhotos{
		HostURL: ct.immich.BaseURL.String(),
		Camps:   albums,
	}
	return c.JSON(200, out)
}

type CampPhotos struct {
	HostURL string
	Camps   []CampAlbum
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
	out, thumbnails, err := ct.createAlbums(args)
	if err != nil {
		return err
	}

	// Best effort to setup the thumbnail, in the background
	go func() {
		err := ct.uploadAlbumThumbnails(thumbnails)
		if err != nil {
			log.Println("backoffice.CampsCreateAlbums:", err)
		}
	}()

	return c.JSON(200, out)
}

type imageForThumbnail struct {
	imageURL string
	albumID  immich.AlbumID
}

func (ct *Controller) createAlbums(args CreateAlbumsIn) (map[cps.IdCamp]immich.AlbumAndLinks, []imageForThumbnail, error) {
	camps, err := cps.SelectCamps(ct.db, args.IdCamps...)
	if err != nil {
		return nil, nil, utils.SQLError(err)
	}

	// safety check : disallow duplicate album
	for _, camp := range camps {
		if camp.AlbumID != "" {
			return nil, nil, fmt.Errorf("internal error: %s already has a photo album", camp.Label())
		}
	}

	api := immich.NewApi(ct.immich)

	out := make(map[cps.IdCamp]immich.AlbumAndLinks)
	var thumbnails []imageForThumbnail
	for _, camp := range camps {
		album, err := api.CreateAlbum(camp.Label())
		if err != nil {
			return nil, nil, err
		}
		out[camp.Id] = album

		// register the ID
		camp.AlbumID = string(album.Id)
		_, err = camp.Update(ct.db)
		if err != nil {
			return nil, nil, utils.SQLError(err)
		}

		if camp.ImageURL != "" {
			thumbnails = append(thumbnails, imageForThumbnail{camp.ImageURL, album.Id})
		}
	}

	return out, thumbnails, nil
}

func (ct *Controller) uploadAlbumThumbnails(l []imageForThumbnail) error {
	api := immich.NewApi(ct.immich)
	for _, album := range l {
		err := api.SetAlbumThumbnail(album.albumID, album.imageURL)
		if err != nil {
			return fmt.Errorf("setting thumbnail %s for album %s: %s", album.imageURL, album.albumID, err)
		}
	}
	return nil
}

func (ct *Controller) CampsDeleteAlbum(c echo.Context) error {
	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	err = ct.deleteAlbum(idCamp)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteAlbum(idCamp cps.IdCamp) error {
	camp, err := cps.SelectCamp(ct.db, idCamp)
	if err != nil {
		return utils.SQLError(err)
	}
	albumId := camp.AlbumID
	if albumId == "" {
		return errors.New("internal error: camp without album")
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		camp.AlbumID = ""
		_, err = camp.Update(tx)
		if err != nil {
			return err
		}
		err = immich.NewApi(ct.immich).DeleteAlbum(immich.AlbumID(albumId))
		if err != nil {
			return err
		}
		return nil
	})
}

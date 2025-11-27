package immich

import (
	"fmt"
	"net/http"
	"testing"

	"registro/config"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Immich {
	tu.LoadEnv(t, "../env.sh")
	out, err := config.NewImmich()
	tu.AssertNoErr(t, err)
	return out
}

func Test(t *testing.T) {
	api := NewApi(devCreds(t))

	var albumL []Album
	err := api.request(http.MethodGet, "/albums", nil, nil, &albumL)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albumL) == 0)

	alb1, err := api.CreateAlbum("Test - 2025")
	tu.AssertNoErr(t, err)

	albC, err := api.LoadAlbum(alb1.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, alb1.AlbumName == albC.AlbumName)

	alb2, err := api.CreateAlbum("Test - 2026")
	tu.AssertNoErr(t, err)

	list, err := api.LoadAlbums([]AlbumID{alb1.Id, alb2.Id})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(list) == 2)

	err = api.request(http.MethodDelete, fmt.Sprintf("/albums/%s", alb1.Id), nil, nil, nil)
	tu.AssertNoErr(t, err)
	err = api.request(http.MethodDelete, fmt.Sprintf("/albums/%s", alb2.Id), nil, nil, nil)
	tu.AssertNoErr(t, err)
}

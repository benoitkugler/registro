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

	var albumL []album
	err := api.request(http.MethodGet, "/albums", nil, nil, &albumL)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(albumL) == 0)

	var alb album
	err = api.request(http.MethodPost, "/albums", nil, map[string]string{"albumName": "TEST Album", "description": "Des souvenirs !"}, &alb)
	tu.AssertNoErr(t, err)

	var sharedLinkL []sharedLink
	err = api.request(http.MethodGet, "/shared-links", map[string]string{"albumId": string(alb.Id)}, nil, &sharedLinkL)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(sharedLinkL) == 0)

	var createdLink sharedLink
	err = api.request(http.MethodPost, "/shared-links", nil, map[string]any{
		"albumId":       alb.Id,
		"type":          "ALBUM",
		"allowDownload": true,
		"allowUpload":   false,
		"description":   "Lien de lecture",
	}, &createdLink)
	tu.AssertNoErr(t, err)

	fmt.Println(createdLink)

	err = api.request(http.MethodDelete, fmt.Sprintf("/albums/%s", alb.Id), nil, nil, nil)
	tu.AssertNoErr(t, err)
}

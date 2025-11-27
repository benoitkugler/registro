package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"registro/config"
)

// https://api.immich.app/endpoints

type Api struct {
	config config.Immich
}

func NewApi(config config.Immich) *Api { return &Api{config} }

func (api *Api) shareURL(key string) string {
	return fmt.Sprintf("%s/share/%s", api.config.BaseURL, key)
}

// if [dst] is non nil, the response body is read as JSON into [dst]
func (api *Api) request(method string, endpoint string, query map[string]string, body any, dst any) error {
	queryParams := url.Values{}
	// always add apiKey
	queryParams.Set("apiKey", api.config.ApiKey)
	for k, v := range query {
		queryParams.Set(k, v)
	}

	fullURL := fmt.Sprintf("%s/api%s?%s", api.config.BaseURL, endpoint, queryParams.Encode())

	var bodyReader io.Reader
	if body != nil {
		var bodyB bytes.Buffer
		err := json.NewEncoder(&bodyB).Encode(body)
		if err != nil {
			return fmt.Errorf("internal error: %s", err)
		}
		bodyReader = &bodyB
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("internal error: %s", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		b, _ := io.ReadAll(resp.Body)
		log.Println(string(b))
		return fmt.Errorf("Immich API at %s returned status code: %d", fullURL, resp.StatusCode)
	}

	if dst == nil {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(dst)
	return err
}

type Album struct {
	AlbumName string `json:"albumName"`
	// albumThumbnailAssetId 	string | Null
	// albumUsers 	AlbumUserResponseDto[]
	AssetCount int `json:"assetCount"`
	// assets 	AssetResponseDto[]
	// contributorCounts 	ContributorCountResponseDto[]
	CreatedAt time.Time `json:"createdAt"`
	// description 	string
	// endDate	time.Time
	// hasSharedLink 	bool
	Id AlbumID `json:"Id"`
	// isActivityEnabled 	bool
	// lastModifiedAssetTimestamp	time.Time
	Order string `json:"order"` // AssetOrder
	// owner 	UserResponseDto
	// ownerId 	string
	// shared 	bool
	// startDate	time.Time
	// updatedAt 	time.Time
}

type AlbumAndLinks struct {
	Album
	EquipiersURL, InscritsURL string
}

type sharedLink struct {
	Album         Album `json:"album"`
	AllowDownload bool  `json:"allowDownload"`
	AllowUpload   bool  `json:"allowUpload"`
	// assets AssetResponseDto[]
	// createdAt time.Time
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Id          AlbumID   `json:"id"`
	Key         string    `json:"key"`
	// Password    string    `json:"password"`
	// showMetadata bool
	// Slug string `json:"slug"`
	// token string | Null
	Type string `json:"type"` // SharedLinkType
	// userId string
}

type AlbumID string

// CreateAlbum creates a new album for the given camp, and
// setup shared links.
func (api *Api) CreateAlbum(campLabel string) (AlbumAndLinks, error) {
	// 1: Create Album
	var alb Album
	err := api.request(http.MethodPost, "/albums", nil, map[string]string{
		"albumName":   campLabel,
		"description": "Créé par Registro",
	}, &alb)
	if err != nil {
		return AlbumAndLinks{}, err
	}

	// 2: Create Equipiers link
	var equipiersLink sharedLink
	err = api.request(http.MethodPost, "/shared-links", nil, map[string]any{
		"albumId":       alb.Id,
		"type":          "ALBUM",
		"allowDownload": true,
		"allowUpload":   true,
		"description":   "Lien équipe - Créé par Registro",
	}, &equipiersLink)
	if err != nil {
		return AlbumAndLinks{}, err
	}

	// 3: Create Inscrits link
	var inscritsLink sharedLink
	err = api.request(http.MethodPost, "/shared-links", nil, map[string]any{
		"albumId":       alb.Id,
		"type":          "ALBUM",
		"allowDownload": true,
		"allowUpload":   false,
		"description":   "Lien incrits - Créé par Registro",
	}, &inscritsLink)
	if err != nil {
		return AlbumAndLinks{}, err
	}

	return AlbumAndLinks{
		Album:        alb,
		EquipiersURL: api.shareURL(equipiersLink.Key),
		InscritsURL:  api.shareURL(inscritsLink.Key),
	}, nil
}

func resolveLinks(list []sharedLink) (upload, download sharedLink, _ error) {
	if L := len(list); L != 2 {
		return sharedLink{}, sharedLink{}, fmt.Errorf("internal error: expected 2 links, found %d", L)
	}
	l1, l2 := list[0], list[1]
	if l1.AllowUpload == l2.AllowUpload {
		return sharedLink{}, sharedLink{}, fmt.Errorf("internal error: unexpected permissions for shared links")
	}
	if l1.AllowUpload {
		upload, download = l1, l2
	} else {
		upload, download = l2, l1
	}
	return
}

// LoadAlbum fetchs the metadata for the given album,
// including the shared links.
// Returns an error if there is not exactly two shared links configured,
// one with upload, the other without.
func (api *Api) LoadAlbum(id AlbumID) (AlbumAndLinks, error) {
	var alb Album
	err := api.request(http.MethodGet, fmt.Sprintf("/albums/%s", id), map[string]string{
		"withoutAssets": "true",
	}, nil, &alb)
	if err != nil {
		return AlbumAndLinks{}, err
	}

	var sharedLinkL []sharedLink
	err = api.request(http.MethodGet, "/shared-links", map[string]string{"albumId": string(alb.Id)}, nil, &sharedLinkL)
	if err != nil {
		return AlbumAndLinks{}, err
	}
	upload, download, err := resolveLinks(sharedLinkL)
	if err != nil {
		return AlbumAndLinks{}, err
	}

	return AlbumAndLinks{
		Album:        alb,
		EquipiersURL: api.shareURL(upload.Key),
		InscritsURL:  api.shareURL(download.Key),
	}, nil
}

// LoadAlbums is the same as [LoadAlbum], but is optimized
// for many albums.
func (api *Api) LoadAlbums(ids []AlbumID) (map[AlbumID]AlbumAndLinks, error) {
	var albumL []Album
	err := api.request(http.MethodGet, "/albums", map[string]string{
		"withoutAssets": "true",
	}, nil, &albumL)
	if err != nil {
		return nil, err
	}
	albumsById := map[AlbumID]Album{}
	for _, alb := range albumL {
		albumsById[alb.Id] = alb
	}

	var sharedLinkL []sharedLink
	err = api.request(http.MethodGet, "/shared-links", nil, nil, &sharedLinkL)
	if err != nil {
		return nil, err
	}
	linksById := map[AlbumID][]sharedLink{}
	for _, link := range sharedLinkL {
		linksById[link.Album.Id] = append(linksById[link.Album.Id], link)
	}

	out := make(map[AlbumID]AlbumAndLinks, len(ids))
	for _, id := range ids {
		alb, hasAlbum := albumsById[id]
		if !hasAlbum {
			return nil, fmt.Errorf("internal error: missing Album for id %s", id)
		}
		upload, download, err := resolveLinks(linksById[id])
		if err != nil {
			return nil, err
		}
		out[id] = AlbumAndLinks{
			Album:        alb,
			EquipiersURL: api.shareURL(upload.Key),
			InscritsURL:  api.shareURL(download.Key),
		}
	}

	return out, nil
}

func (api *Api) DeleteAlbum(id AlbumID) error {
	err := api.request(http.MethodDelete, fmt.Sprintf("/albums/%s", id), nil, nil, nil)
	return err
}

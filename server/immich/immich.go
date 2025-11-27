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

// if [dst] is non nil, the response body is read as JSON into [dst]
func (api *Api) request(method string, endpoint string, query map[string]string, body any, dst any) error {
	queryParams := url.Values{}
	// always add apiKey
	queryParams.Set("apiKey", api.config.ApiKey)
	for k, v := range query {
		queryParams.Set(k, v)
	}
	fullURL := fmt.Sprintf("%s%s?%s", api.config.BaseURL, endpoint, queryParams.Encode())

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

type album struct {
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

type sharedLink struct {
	//	album AlbumResponseDto
	AllowDownload bool `json:"allowDownload"`
	AllowUpload   bool `json:"allowUpload"`
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
func (api *Api) CreateAlbum()

// LoadAlbum fetchs the metadata for the given albums,
// including the shared links.
func (api *Api) LoadAlbum(id AlbumID)

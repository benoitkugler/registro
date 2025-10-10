package joomeo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"registro/config"
	"registro/utils"
)

const (
	messageDirecteur = `Vous avez maintenant accès à l'album de votre camp.
	Vous pouvez retrouver les informations liées à cet album sur votre espace directeur.
	`

	messageParents = `Bonjour,

    Envie d'avoir un aperçu des activités du %s ?
    Des photos sont disponibles (ou le seront très prochainement) sur un espace dédié.
    Vous y avez accès en suivant le lien ci-dessous. 
    
    Vous pouvez retrouver les informations de connexion sur votre espace d'inscription.

	Bon visionnage !
	`
)

// unix timestamp, in seconds
type date float64

func (d date) date() time.Time { return time.Unix(int64(d), 0) }

type joomeoError struct {
	Error json.RawMessage `json:"error"`
}

func checkError(resp *http.Response, successCode int) error {
	if resp.StatusCode == successCode {
		return nil
	}
	// decode the error
	var out joomeoError
	err := json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return fmt.Errorf("impossible de décoder l'erreur: %s", err)
	}
	return fmt.Errorf("erreur renvoyée par Joomeo: code %d, %s", resp.StatusCode, out.Error)
}

// if body is not nil, it is encoded as JSON
func sendRequest(method string, url string, body any, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		var bodyB bytes.Buffer
		err := json.NewEncoder(&bodyB).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("erreur interne: %s", err)
		}
		bodyReader = &bodyB
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("erreur interne: %s", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection à Joomeo: %s", err)
	}

	return resp, nil
}

type (
	AlbumId   = string
	FolderId  = string
	ContactId = string
)

// Api exposes the REST JSON Joomeo API
// See https://service.joomeo.com/doc
type Api struct {
	apiKey          string // copied from config
	rootFolderLabel string // copied from config

	sessionid string
	spaceName string
}

const baseUrl = "https://service.joomeo.com"

func NewApi(credences config.Joomeo) (*Api, error) {
	const url = baseUrl + "/session"

	type createSession struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	resp, err := sendRequest(http.MethodPost, url, createSession{
		Login:    credences.Login,
		Password: credences.Password,
	}, map[string]string{
		"X-API-KEY":      credences.Apikey,
		"X-PAYLOAD-TYPE": "manager",
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err = checkError(resp, 201); err != nil {
		return nil, err
	}

	var out struct {
		Sessionid   string `json:"sessionid"`
		SessionType int    `json:"sessionType"`
		// 0	Manager
		// 1	Guest
		// 3	Cart
		SpaceName string `json:"spaceName"`
		Admin     int    `json:"admin"`
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	return &Api{
		apiKey:          credences.Apikey,
		rootFolderLabel: credences.RootFolder,
		sessionid:       out.Sessionid,
		spaceName:       out.SpaceName,
	}, nil
}

func (api Api) SpaceURL() string {
	const baseURL = "https://private.joomeo.com/users/"
	return baseURL + api.spaceName
}

// Close termine la session, rendant `api` inutilisable.
func (api Api) Close() {
	const url = baseUrl + "/session"
	resp, err := sendRequest(http.MethodDelete, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		log.Println("internal error:", err)
		return
	}

	if err = checkError(resp, 204); err != nil {
		log.Println("internal error:", err)
	}
}

type folderJ struct {
	FolderId FolderId `json:"folderid"`
	Label    string   `json:"label"`
}

// pass an empty string for the root folder
func (api Api) getFolders(id FolderId) ([]folderJ, error) {
	url := fmt.Sprintf("%s/users/%s/folders?folderid=%s", baseUrl, api.spaceName, id)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}

	var response struct {
		TotalCount int       `json:"totalCount"`
		PageCount  int       `json:"pageCount"`
		List       []folderJ `json:"list,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if response.PageCount < response.TotalCount {
		return nil, fmt.Errorf("internal error: number of folders not supported")
	}

	return response.List, nil
}

// [GetSejoursFolder] return the sejours root folder id,
// found by comparing folder names to the one specified in the config.
func (api *Api) GetSejoursFolder() (FolderId, error) {
	rootFolders, err := api.getFolders("")
	if err != nil {
		return "", err
	}
	for _, folder := range rootFolders {
		if folder.Label == api.rootFolderLabel {
			return folder.FolderId, nil
		}
	}
	return "", fmt.Errorf("aucun dossier %s trouvé sur le serveur Joomeo", api.rootFolderLabel)
}

type albumJ struct {
	AlbumId  AlbumId `json:"albumid"`
	Label    string  `json:"label"`
	Date     date    `json:"date"`
	FolderId string  `json:"folderid"`
}

func queryByFolder(folderId FolderId) url.Values {
	query := make(url.Values)
	query.Set("filter", fmt.Sprintf("folderid=%s", folderId))
	return query
}

func queryByContact(contactId string) url.Values {
	query := make(url.Values)
	query.Set("contactid", contactId)
	return query
}

// returns the albums specified by [query]
func (api Api) getAlbums(query url.Values) ([]albumJ, error) {
	url := fmt.Sprintf("%s/users/%s/albums?%s", baseUrl, api.spaceName, query.Encode())
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "folderid,label,date",
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}

	var response struct {
		TotalCount int      `json:"totalCount"`
		PageCount  int      `json:"pageCount"`
		List       []albumJ `json:"list"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if response.PageCount < response.TotalCount {
		return nil, fmt.Errorf("internal error: number of folders not supported")
	}
	return response.List, nil
}

type Album struct {
	Id    AlbumId
	Label string
	Date  time.Time

	FilesCount int

	Contacts []ContactPermission
}

func newAlbum(alb albumJ, filesCount int, contacts []ContactPermission) Album {
	return Album{alb.AlbumId, alb.Label, alb.Date.date(), filesCount, contacts}
}

// FindContact performs a case insensitive match in [Contacts]
func (alb Album) FindContact(mail string) (ContactPermission, bool) {
	mail = strings.ToLower(mail)
	for _, contact := range alb.Contacts {
		if strings.ToLower(contact.Mail) == mail {
			return contact, true
		}
	}
	return ContactPermission{}, false
}

// LoadAlbums fetches the given album and contacts, which pay be rather slow.
func (api *Api) LoadAlbums(ids utils.Set[AlbumId]) (map[AlbumId]Album, error) {
	root, err := api.GetSejoursFolder()
	if err != nil {
		return nil, err
	}
	// fetch all albums at first, since Joomeo does not provide
	// a way to query a list of albums in one call.
	albums, err := api.getAlbums(queryByFolder(root))
	if err != nil {
		return nil, err
	}
	out := make(map[AlbumId]Album, len(albums))
	for _, alb := range albums {
		if !ids.Has(alb.AlbumId) {
			continue
		}
		contacts, err := api.loadContactsFor(alb.AlbumId)
		if err != nil {
			return nil, err
		}
		filesCount, err := api.resolveAlbumFilesCount(alb.AlbumId)
		if err != nil {
			return nil, err
		}
		out[alb.AlbumId] = newAlbum(alb, filesCount, contacts)
	}
	return out, nil
}

// LoadAlbum fetches one album. Use [LoadAlbums] to query a list.
func (api *Api) LoadAlbum(albumId AlbumId) (Album, error) {
	url := fmt.Sprintf("%s/users/%s/albums/%s", baseUrl, api.spaceName, albumId)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "folderid,label,date",
	})
	if err != nil {
		return Album{}, err
	}
	if err = checkError(resp, 200); err != nil {
		return Album{}, err
	}

	var out albumJ
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return Album{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	contacts, err := api.loadContactsFor(albumId)
	if err != nil {
		return Album{}, err
	}

	filesCount, err := api.resolveAlbumFilesCount(albumId)
	if err != nil {
		return Album{}, err
	}

	return newAlbum(out, filesCount, contacts), nil
}

type AccessRules struct {
	AllowCreateAlbum     bool `json:"allowCreateAlbum"`
	AllowDeleteAlbum     bool `json:"allowDeleteAlbum"`
	AllowDeleteFile      bool `json:"allowDeleteFile"`
	AllowEditFileCaption bool `json:"allowEditFileCaption"`
	AllowUpdateAlbum     bool `json:"allowUpdateAlbum"`
}

type AlbumAccessRules struct {
	AllowDownload   bool `json:"allowDownload"`
	AllowUpload     bool `json:"allowUpload"`
	AllowPrintOrder bool `json:"allowPrintOrder"`
	AllowComments   bool `json:"allowComments"`
}

// loadContactsFor renvoi les contacts de l'album demandé,
// avec les permissions correspondantes.
func (api *Api) loadContactsFor(albumId AlbumId) ([]ContactPermission, error) {
	url := fmt.Sprintf("%s/users/%s/contacts?albumid=%s", baseUrl, api.spaceName, albumId)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "email,password,albumAccessRules,accessRules,type",
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		TotalCount int                 `json:"totalCount"`
		PageCount  int                 `json:"pageCount"`
		List       []ContactPermission `json:"list"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if response.PageCount < response.TotalCount {
		return nil, fmt.Errorf("internal error: number of contacts not supported")
	}

	return response.List, nil
}

func (api *Api) resolveAlbumFilesCount(albumId AlbumId) (int, error) {
	url := fmt.Sprintf("%s/users/%s/albums/%s/files", baseUrl, api.spaceName, albumId)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return 0, err
	}
	if err = checkError(resp, 200); err != nil {
		return 0, err
	}

	var response struct {
		TotalCount int `json:"totalCount"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	return response.TotalCount, nil
}

// CreateAlbum crates a new album in [folderId].
func (api *Api) CreateAlbum(folderId FolderId, label string) (Album, error) {
	url := fmt.Sprintf("%s/users/%s/albums", baseUrl, api.spaceName)
	resp, err := sendRequest(http.MethodPost, url, map[string]any{
		"label":         label,
		"folderid":      folderId,
		"sortingKey":    "manual",
		"sortingType":   "ascending",
		"date":          time.Now().Unix(),
		"displayFormat": "1920",
		"watermark":     false,
	}, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return Album{}, err
	}
	if err = checkError(resp, 201); err != nil {
		return Album{}, err
	}

	var response albumJ
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return Album{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	return newAlbum(response, 0, nil), nil
}

func (api *Api) DeleteAlbum(id AlbumId) error {
	url := fmt.Sprintf("%s/users/%s/albums/%s", baseUrl, api.spaceName, id)
	resp, err := sendRequest(http.MethodDelete, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return err
	}
	return checkError(resp, 204)
}

func (api *Api) deleteContact(id string) error {
	url := fmt.Sprintf("%s/users/%s/contacts/%s", baseUrl, api.spaceName, id)
	resp, err := sendRequest(http.MethodDelete, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return err
	}
	return checkError(resp, 204)
}

// ContactPermission ajoute à un contact les permissions
// pour un album donné.
type ContactPermission struct {
	Contact
	AlbumPermissions AlbumAccessRules `json:"albumAccessRules"`
}

func (api *Api) loadContact(contactId string) (Contact, error) {
	url := fmt.Sprintf("%s/users/%s/contacts/%s", baseUrl, api.spaceName, contactId)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "email,password,accessRules,type",
	})
	if err != nil {
		return Contact{}, fmt.Errorf("loading contact: %s", err)
	}
	if err = checkError(resp, 200); err != nil {
		return Contact{}, fmt.Errorf("loading contact: %s", err)
	}
	defer resp.Body.Close()

	var contact Contact
	err = json.NewDecoder(resp.Body).Decode(&contact)
	if err != nil {
		return Contact{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	return contact, nil
}

// AddDirecteur ajoute un directeur comme super contact, et lui donne
// le droit d'écriture sur l'album donné.
// Renvoie le contact créé.
func (api *Api) AddDirecteur(albumid AlbumId, mailDirecteur string, envoiMail bool) (ContactPermission, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	resp, err := sendRequest(http.MethodPost, url, map[string]any{
		"email": mailDirecteur,
		"accessRules": AccessRules{
			AllowCreateAlbum:     false,
			AllowDeleteAlbum:     false,
			AllowUpdateAlbum:     false,
			AllowDeleteFile:      true,
			AllowEditFileCaption: true,
		},
		"returnDataIfExists": true,
		"type":               1,
	}, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return ContactPermission{}, fmt.Errorf("creating contact: %s", err)
	}
	if err = checkError(resp, 201); err != nil {
		return ContactPermission{}, fmt.Errorf("creating contact: %s", err)
	}

	// there seems to be a bug in Joomeo API:
	// when the contact alreay exists, its data is not properly returned
	// do it in two API calls then
	var res struct {
		Id string `json:"contactid"`
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return ContactPermission{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	contact, err := api.loadContact(res.Id)
	if err != nil {
		return ContactPermission{}, err
	}
	permissions, err := api.setContactUploader(albumid, contact.Id, envoiMail)
	if err != nil {
		return ContactPermission{}, err
	}

	return ContactPermission{contact, permissions}, nil
}

func (api *Api) setContactUploader(albumid, contactId string, mailForDirecteur bool) (AlbumAccessRules, error) {
	url := fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contactId, albumid)
	// invite with write access
	args := map[string]any{
		"allowDownload":   true,
		"allowUpload":     true,
		"allowPrintOrder": true,
		"allowComments":   true,
	}
	if mailForDirecteur {
		args["subject"] = "Album photos Joomeo"
		args["message"] = messageDirecteur
	}
	resp, err := sendRequest(http.MethodPut, url, args, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return AlbumAccessRules{}, fmt.Errorf("adding contact to album: %s", err)
	}
	defer resp.Body.Close()

	if err = checkError(resp, 200); err != nil {
		return AlbumAccessRules{}, fmt.Errorf("adding contact to album: %s", err)
	}

	var response AlbumAccessRules
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return AlbumAccessRules{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	return response, nil
}

// Renvoi le contact et la liste des albums authorisés du contact associé à l'adresse mail fournie.
// Le contact renvoyé peut être vide si [mail] n'est pas utilisé.
func (api *Api) GetContactByMail(mail string) (Contact, []string, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FILTER":    fmt.Sprintf("email=%s", mail),
		"X-FIELDS":    "email,password,accessRules,type",
	})
	if err != nil {
		return Contact{}, nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return Contact{}, nil, err
	}

	var response struct {
		List []Contact `json:"list"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return Contact{}, nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	if len(response.List) == 0 { // aucun contact n'est encore enregistré
		return Contact{}, nil, nil
	}
	contact := response.List[0]

	albums, err := api.getAlbums(queryByContact(contact.Id))
	if err != nil {
		return Contact{}, nil, err
	}
	labels := make([]string, len(albums))
	for i, album := range albums {
		labels[i] = album.Label
	}

	return contact, labels, nil
}

// AddContacts imite [AddDirecteur] mais pour plusieurs contacts,
// avec une permission lecture et un message adapté.
// Si le contact est déjà associé à l'album, il est ignoré.
func (api *Api) AddContacts(camp string, albumId AlbumId, mails []string, sendMail bool) error {
	// 0 - fetch already linked contacts
	// 1 - create all the contacts
	// 2 - send the invite one by one

	tmp, err := api.loadContactsFor(albumId)
	if err != nil {
		return err
	}
	alreadyAttached := utils.NewSet[ContactId]()
	for _, ct := range tmp {
		alreadyAttached.Add(ct.Id)
	}

	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	contacts := make([]map[string]any, len(mails))
	for i, mail := range mails {
		contacts[i] = map[string]any{
			"email":              mail,
			"returnDataIfExists": true,
			"type":               0,
		}
	}
	resp1, err := sendRequest(http.MethodPost, url, contacts, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-BULK":      "1",
		"X-FIELDS":    "email,password",
	})
	if err != nil {
		return fmt.Errorf("creating contact: %s", err)
	}
	defer resp1.Body.Close()
	if err = checkError(resp1, 201); err != nil {
		return fmt.Errorf("creating contact: %s", err)
	}
	var out struct {
		Successes []Contact         `json:"successes"`
		Failures  []json.RawMessage `json:"failures"`
	}
	err = json.NewDecoder(resp1.Body).Decode(&out)
	if err != nil {
		return fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if len(out.Failures) != 0 {
		return fmt.Errorf("creating contacts: %v", out.Failures)
	}

	// share the album
	args := map[string]any{
		"allowDownload":   true,
		"allowUpload":     false,
		"allowPrintOrder": true,
		"allowComments":   true,
	}
	if sendMail {
		args["subject"] = "Album photos"
		args["message"] = fmt.Sprintf(messageParents, camp)
	}

	for _, contact := range out.Successes {
		if alreadyAttached.Has(contact.Id) {
			continue
		}

		url = fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contact.Id, albumId)
		resp, err := sendRequest(http.MethodPut, url, args, map[string]string{
			"X-API-KEY":   api.apiKey,
			"X-SESSIONID": api.sessionid,
		})
		if err != nil {
			return fmt.Errorf("Invitation du contact %s impossible (%s).", contact.Mail, err)
		}
		if err = checkError(resp, 200); err != nil {
			return fmt.Errorf("Invitation du contact %s impossible (%s).", contact.Mail, err)
		}
		_ = resp.Body.Close()
	}

	return nil
}

//
// TODO - check the following
//

type folder struct {
	Id    FolderId `json:"folderid"`
	Label string   `json:"label"`
	id    int64
	// childs []AlbumExt
}

// func (api *Api) GetAllAlbumsContacts() (map[FolderId]folder, map[string]AlbumExt, map[string]Contact, error) {
// 	campsFolderId, err := api.GetSejoursFolder()
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}

// 	sejoursFolders, err := api.getFolders(campsFolderId)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}

// 	albumsList, err := api.getAlbumsOld("")
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}

// 	folders := make(map[FolderId]folder)
// 	for i, f := range sejoursFolders {
// 		fo := folder{
// 			Id:    f.FolderId,
// 			Label: f.Label,
// 			id:    int64(i),
// 		}
// 		folders[f.FolderId] = fo
// 	}

// 	albums := make(map[string]AlbumExt, len(albumsList))
// 	for _, album := range albumsList {
// 		if folder, isIn := folders[album.FolderId]; isIn {
// 			albums[album.AlbumId] = album
// 			folder.childs = append(folder.childs, album)
// 			folders[album.FolderId] = folder
// 		}
// 	}

// 	contactsList, err := api.getContacts()
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	contacts := make(map[string]Contact, len(contactsList))
// 	for _, c := range contactsList {
// 		contacts[c.Id] = Contact{
// 			Id:       c.Id,
// 			Mail:     c.Mail,
// 			Login:    c.Login,
// 			Password: c.Password,
// 		}
// 	}

// 	return folders, albums, contacts, nil
// }

// Contact représente un contact Joomeo
type Contact struct {
	Id       string `json:"contactid"`
	Mail     string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`

	AccesRules AccessRules `json:"accessRules"`
	Type       int         `json:"type"`
}

func (api Api) getContacts() ([]Contact, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "email,password",
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}

	var response struct {
		TotalCount int       `json:"totalCount"`
		PageCount  int       `json:"pageCount"`
		List       []Contact `json:"list"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if response.PageCount < response.TotalCount {
		return nil, fmt.Errorf("internal error: number of folders not supported")
	}
	return response.List, nil
}

// setAdvancedRights make sure existing contacts have advanced rights,
// required for directeur and write rights
func (api *Api) setAdvancedRights(contactId string) error {
	url := fmt.Sprintf("%s/users/%s/contacts/%s", baseUrl, api.spaceName, contactId)
	resp, err := sendRequest(http.MethodPut, url, map[string]any{
		"accessRules": AccessRules{
			AllowCreateAlbum:     false,
			AllowDeleteAlbum:     false,
			AllowUpdateAlbum:     false,
			AllowDeleteFile:      true,
			AllowEditFileCaption: true,
		},
		"type": 1,
	}, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}
	_ = resp.Body.Close()
	if err = checkError(resp, 200); err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}
	return nil
}

// SetContactUploader ajoute le contact comme uploader.
func (api *Api) SetContactUploader(albumId AlbumId, contactId ContactId) error {
	err := api.setAdvancedRights(contactId)
	if err != nil {
		return err
	}

	_, err = api.setContactUploader(albumId, contactId, false)
	return err
}

// UnlinkContact retire l'accès à l'album donné pour le contact donné.
// Le contact n'est pas supprimé et conserve son accès aux autres albums.
func (api *Api) UnlinkContact(albumId AlbumId, contactId ContactId) error {
	url := fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contactId, albumId)
	resp, err := sendRequest(http.MethodDelete, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return err
	}
	if err = checkError(resp, 204); err != nil {
		return err
	}
	return nil
}

package joomeo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"registro/config"
)

const (
	RootFolder       = "SEJOURS"
	messageDirecteur = `Vous avez maintenant accès à l'album de votre camp (permission lecture/écriture).
	Vous pouvez retrouver les informations liées à cette album sur votre espace directeur (https://acve.fr/directeurs).
	`
	messageParents = `Bonjour,

    Envie d'avoir un aperçu des activités du %s édition %d?
    Des photos sont disponibles (ou le seront très prochainement) sur un espace dédié.
    Vous y avez accès en suivant le lien ci-dessous. 
    
    Vous pouvez retrouver les informations de connexion sur votre espace personnel ACVE.

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
func sendRequest(method string, url string, body interface{}, headers map[string]string) (*http.Response, error) {
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

// ApiJoomeo exposes the REST JSON Joomeo API
// See https://service.joomeo.com/doc
type ApiJoomeo struct {
	apiKey    string
	sessionid string
	spaceName string
}

const baseUrl = "https://service.joomeo.com"

func InitApi(credences config.Joomeo) (*ApiJoomeo, error) {
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

	return &ApiJoomeo{
		apiKey:    credences.Apikey,
		sessionid: out.Sessionid,
		spaceName: out.SpaceName,
	}, nil
}

func (api ApiJoomeo) SpaceURL() string {
	const baseURL = "https://private.joomeo.com/users/"
	return baseURL + api.spaceName
}

// Close termine la session, rendant `api` inutilisable.
func (api ApiJoomeo) Close() {
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
	FolderId string `json:"folderid"`
	Label    string `json:"label"`
}

// pass an empty string for the root folder
func (api ApiJoomeo) getFolders(id string) ([]folderJ, error) {
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

type Date float64

func (d Date) Time() time.Time { return time.Unix(int64(d/1000), 0) }

// ContactPermission ajoute à un contact les permissions
// pour un album donné.
type ContactPermission struct {
	Contact
	Allowupload   int `json:"allowupload,omitempty"`
	Allowdownload int `json:"allowdownload,omitempty"`
}

type Album struct {
	AlbumId  string `json:"albumid"`
	Label    string `json:"label"`
	Date     date   `json:"date"`
	FolderId string `json:"folderid"`

	NbFiles int // not returned by JOOMEO
}

// if contactid is not empty, restrict to the ones shared to it
func (api ApiJoomeo) getAlbums(contactid string) ([]Album, error) {
	url := fmt.Sprintf("%s/users/%s/albums", baseUrl, api.spaceName)
	if contactid != "" {
		url += "?contactid=" + contactid
	}
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "folderid",
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}

	var response struct {
		TotalCount int     `json:"totalCount"`
		PageCount  int     `json:"pageCount"`
		List       []Album `json:"list"`
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

// GetAlbumMetadatas renvoi un dictionnaire contenant un résumé des informations
// de l'album demandé
func (api *ApiJoomeo) GetAlbumMetadatas(albumid string) (Album, error) {
	url := fmt.Sprintf("%s/users/%s/albums/%s", baseUrl, api.spaceName, albumid)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "folderid",
	})
	if err != nil {
		return Album{}, err
	}
	if err = checkError(resp, 200); err != nil {
		return Album{}, err
	}

	var out Album
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return Album{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	// resolve number of files
	url = fmt.Sprintf("%s/users/%s/albums/%s/files", baseUrl, api.spaceName, albumid)
	resp2, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return Album{}, err
	}
	if err = checkError(resp2, 200); err != nil {
		return Album{}, err
	}

	var response2 struct {
		TotalCount int `json:"totalCount"`
	}
	err = json.NewDecoder(resp2.Body).Decode(&response2)
	if err != nil {
		return Album{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	out.NbFiles = response2.TotalCount

	return out, nil
}

type Folder struct {
	FolderId string `json:"folderid"`
	Label    string `json:"label"`
	id       int64
	childs   []Album
}

func (api *ApiJoomeo) GetAllAlbumsContacts() (map[string]Folder, map[string]Album, map[string]Contact, error) {
	rootFolders, err := api.getFolders("")
	if err != nil {
		return nil, nil, nil, err
	}
	var campsFolderid string
	for _, folder := range rootFolders {
		if folder.Label == RootFolder {
			campsFolderid = folder.FolderId
		}
	}
	if campsFolderid == "" {
		return nil, nil, nil, fmt.Errorf("aucun dossier %s trouvé sur le serveur Joomeo", RootFolder)
	}
	sejoursFolders, err := api.getFolders(campsFolderid)
	if err != nil {
		return nil, nil, nil, err
	}

	albumsList, err := api.getAlbums("")
	if err != nil {
		return nil, nil, nil, err
	}

	folders := make(map[string]Folder)
	for i, folder := range sejoursFolders {
		fo := Folder{
			FolderId: folder.FolderId,
			Label:    folder.Label,
			id:       int64(i),
		}
		folders[folder.FolderId] = fo
	}

	albums := make(map[string]Album, len(albumsList))
	for _, album := range albumsList {
		if folder, isIn := folders[album.FolderId]; isIn {
			albums[album.AlbumId] = album
			folder.childs = append(folder.childs, album)
			folders[album.FolderId] = folder
		}
	}

	contactsList, err := api.getContacts()
	if err != nil {
		return nil, nil, nil, err
	}
	contacts := make(map[string]Contact, len(contactsList))
	for _, c := range contactsList {
		contacts[c.ContactId] = Contact{
			ContactId: c.ContactId,
			Mail:      c.Mail,
			Login:     c.Login,
			Password:  c.Password,
		}
	}

	return folders, albums, contacts, nil
}

// Contact représente un contact Joomeo
type Contact struct {
	ContactId string `json:"contactid"`
	Mail      string `json:"email"`
	Login     string `json:"login"`
	Password  string `json:"password"`
}

func (api ApiJoomeo) getContacts() ([]Contact, error) {
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
func (api *ApiJoomeo) setAdvancedRights(contactId string) error {
	url := fmt.Sprintf("%s/users/%s/contacts/%s", baseUrl, api.spaceName, contactId)
	resp2, err := sendRequest(http.MethodPut, url, map[string]interface{}{
		"accessRules": map[string]bool{
			"allowCreateAlbum":     false,
			"allowDeleteAlbum":     false,
			"allowUpdateAlbum":     false,
			"allowDeleteFile":      true,
			"allowEditFileCaption": true,
		},
		"type": 1,
	}, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}
	_ = resp2.Body.Close()
	if err = checkError(resp2, 200); err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}
	return nil
}

// AjouteDirecteur ajoute un directeur en contact avec droit d'écriture.
// Renvoie le contact créé.
func (api *ApiJoomeo) AjouteDirecteur(albumid, mailDirecteur string, envoiMail bool) (Contact, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	resp1, err := sendRequest(http.MethodPost, url, map[string]interface{}{
		"email": mailDirecteur,
		"accessRules": map[string]bool{
			"allowCreateAlbum":     false,
			"allowDeleteAlbum":     false,
			"allowUpdateAlbum":     false,
			"allowDeleteFile":      true,
			"allowEditFileCaption": true,
		},
		"returnDataIfExists": true,
		"type":               1,
	}, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "email,password",
	})
	if err != nil {
		return Contact{}, fmt.Errorf("creating contact: %s", err)
	}
	defer resp1.Body.Close()
	if err = checkError(resp1, 201); err != nil {
		return Contact{}, fmt.Errorf("creating contact: %s", err)
	}
	var contact Contact
	err = json.NewDecoder(resp1.Body).Decode(&contact)
	if err != nil {
		return Contact{}, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	out := Contact(contact)

	err = api.setContactUploader(albumid, contact.ContactId, true)
	if err != nil {
		return Contact{}, err
	}

	return out, nil
}

// AjouteContacts imite AjouteDirecteur mais pour plusieurs contacts,
// avec une permission lecture et un message adapté.
// Renvoie une liste d'erreurs (mail par mail) et une erreur globale.
func (api *ApiJoomeo) AjouteContacts(campNom string, campAnnee int, albumid string, mails []string, envoiMail bool) ([]string, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)

	contacts := make([]map[string]interface{}, len(mails))
	for i, mail := range mails {
		contacts[i] = map[string]interface{}{
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
		return nil, fmt.Errorf("creating contact: %s", err)
	}
	defer resp1.Body.Close()
	if err = checkError(resp1, 201); err != nil {
		return nil, fmt.Errorf("creating contact: %s", err)
	}

	var out struct {
		Successes []Contact         `json:"successes"`
		Failures  []json.RawMessage `json:"failures"`
	}
	err = json.NewDecoder(resp1.Body).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}

	var allErrors []string
	for _, err := range out.Failures {
		allErrors = append(allErrors, string(err))
	}

	// share the album
	args := map[string]interface{}{
		"allowDownload":   true,
		"allowUpload":     false,
		"allowPrintOrder": true,
		"allowComments":   true,
	}
	if envoiMail {
		args["subject"] = "Album photos ACVE"
		args["message"] = fmt.Sprintf(messageParents, campNom, campAnnee)
	}

	for _, contact := range out.Successes {
		url = fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contact.ContactId, albumid)
		resp3, err := sendRequest(http.MethodPut, url, args, map[string]string{
			"X-API-KEY":   api.apiKey,
			"X-SESSIONID": api.sessionid,
		})
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Invitation du contact %s impossible (%s).", contact.Mail, err))
		}
		defer resp3.Body.Close()
		if err = checkError(resp3, 200); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Invitation du contact %s impossible (%s).", contact.Mail, err))
		}
	}

	return allErrors, nil
}

// GetContactsFor renvoi les contacts de l'album demandé,
// avec les permissions correspondantes.
func (api *ApiJoomeo) GetContactsFor(albumid string) ([]ContactPermission, error) {
	url := fmt.Sprintf("%s/users/%s/contacts?albumid=%s", baseUrl, api.spaceName, albumid)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FIELDS":    "email,password,albumAccessRules",
	})
	if err != nil {
		return nil, err
	}
	if err = checkError(resp, 200); err != nil {
		return nil, err
	}

	type contactPermission struct {
		Contact
		AlbumAccessRules struct {
			AllowDownload bool `json:"allowDownload"`
			AllowUpload   bool `json:"allowUpload"`
		} `json:"albumAccessRules"`
	}

	var response struct {
		TotalCount int                 `json:"totalCount"`
		PageCount  int                 `json:"pageCount"`
		List       []contactPermission `json:"list"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("impossible de décoder la réponse Joomeo: %s", err)
	}
	if response.PageCount < response.TotalCount {
		return nil, fmt.Errorf("internal error: number of folders not supported")
	}

	bToI := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	out := make([]ContactPermission, len(response.List))
	for i, c := range response.List {
		out[i] = ContactPermission{
			Contact:       Contact(c.Contact),
			Allowupload:   bToI(c.AlbumAccessRules.AllowUpload),
			Allowdownload: bToI(c.AlbumAccessRules.AllowDownload),
		}
	}
	return out, nil
}

// SetContactUploader ajoute le contact comme uploader.
func (api *ApiJoomeo) SetContactUploader(albumid, contactid string) error {
	return api.setContactUploader(albumid, contactid, false)
}

// also calls setAdvancedRights
func (api *ApiJoomeo) setContactUploader(albumid, contactId string, mailForDirecteur bool) error {
	err := api.setAdvancedRights(contactId)
	if err != nil {
		return err
	}

	// invite with write access
	args := map[string]interface{}{
		"allowDownload":   true,
		"allowUpload":     true,
		"allowPrintOrder": true,
		"allowComments":   true,
	}
	if mailForDirecteur {
		args["subject"] = "Album photos Joomeo"
		args["message"] = messageDirecteur
	}
	url := fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contactId, albumid)
	resp, err := sendRequest(http.MethodPut, url, args, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
	})
	if err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}
	defer resp.Body.Close()
	if err = checkError(resp, 200); err != nil {
		return fmt.Errorf("updating contact: %s", err)
	}

	return nil
}

// RemoveContact retire l'accès à l'album donné pour le contact donné.
// Le contact n'est pas supprimé et conserve son accès aux autres albums.
func (api *ApiJoomeo) RemoveContact(albumid, contactid string) error {
	url := fmt.Sprintf("%s/users/%s/contacts/%s/albums/%s", baseUrl, api.spaceName, contactid, albumid)
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

// Renvoi le contact et la liste des albums authorisés du contact associé à l'adresse mail fournie.
// Le contact peut être zéro.
func (api *ApiJoomeo) GetLoginFromMail(mail string) (Contact, []Album, error) {
	url := fmt.Sprintf("%s/users/%s/contacts", baseUrl, api.spaceName)
	resp, err := sendRequest(http.MethodGet, url, nil, map[string]string{
		"X-API-KEY":   api.apiKey,
		"X-SESSIONID": api.sessionid,
		"X-FILTER":    fmt.Sprintf("email=%s", mail),
		"X-FIELDS":    "email,password",
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

	albums, err := api.getAlbums(contact.ContactId)
	if err != nil {
		return Contact{}, nil, err
	}

	return Contact(contact), albums, nil
}

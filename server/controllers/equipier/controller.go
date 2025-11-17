package equipier

import (
	"database/sql"
	"errors"
	"fmt"

	"registro/config"
	filesAPI "registro/controllers/files"
	"registro/crypto"
	"registro/joomeo"
	"registro/logic"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	db *sql.DB

	key    crypto.Encrypter
	files  fs.FileSystem
	joomeo config.Joomeo
}

func NewController(db *sql.DB, key crypto.Encrypter, files fs.FileSystem, joomeo config.Joomeo) *Controller {
	return &Controller{db, key, files, joomeo}
}

func (ct *Controller) Load(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[cps.IdEquipier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	out, err := ct.load(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Camp struct {
	Nom       string
	DateDebut shared.Date
	Duree     int
}

type DemandeEquipier struct {
	Demande     fs.Demande
	Optionnelle bool
	Files       []logic.PublicFile // uploaded by the user
}

type EquipierExt struct {
	Equipier cps.Equipier
	Personne pr.Etatcivil
	Camp     Camp

	Demandes []DemandeEquipier
}

func (ct *Controller) load(id cps.IdEquipier) (EquipierExt, error) {
	equipier, err := cps.SelectEquipier(ct.db, id)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, equipier.IdPersonne)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	camp, err := cps.SelectCamp(ct.db, equipier.IdCamp)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}

	links, err := fs.SelectDemandeEquipiersByIdEquipiers(ct.db, id)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	files, demandesM, err := filesAPI.LoadFilesPersonnes(ct.db, ct.key, links.IdDemandes(), equipier.IdPersonne)
	if err != nil {
		return EquipierExt{}, err
	}

	demandes := make([]DemandeEquipier, len(links))
	for i, link := range links {
		demandes[i] = DemandeEquipier{
			Demande:     demandesM[link.IdDemande],
			Optionnelle: link.Optionnelle,
			Files:       files[equipier.IdPersonne][link.IdDemande],
		}
	}

	out := EquipierExt{equipier, personne.Etatcivil, Camp{camp.Nom, camp.DateDebut, camp.Duree}, demandes}
	return out, nil
}

// LoadJoomeo loads the Joomeo data,
// which may be quite slow
func (ct *Controller) LoadJoomeo(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[cps.IdEquipier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}
	out, err := ct.loadJoomeo(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Joomeo struct {
	SpaceURL string
	Login    string
	Password string
}

func (ct *Controller) loadJoomeo(id cps.IdEquipier) (Joomeo, error) {
	equipier, err := cps.SelectEquipier(ct.db, id)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, equipier.IdPersonne)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}

	api, err := joomeo.NewApi(ct.joomeo)
	if err != nil {
		return Joomeo{}, err
	}
	defer api.Close()

	contact, _, err := api.GetContactByMail(personne.Mail)
	if err != nil {
		return Joomeo{}, fmt.Errorf("accès Joomeo: %s", err)
	}

	return Joomeo{api.SpaceURL(), contact.Login, contact.Password}, nil
}

type UpdateIn struct {
	Token string

	Personne pr.Etatcivil
	Presence cps.PresenceOffsets
}

// Update updates the Equipier and the underlying Personne
func (ct *Controller) Update(c echo.Context) error {
	var args UpdateIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	id, err := crypto.DecryptID[cps.IdEquipier](ct.key, args.Token)
	if err != nil {
		return err
	}

	err = ct.update(id, args.Personne, args.Presence)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) update(id cps.IdEquipier, argsP pr.Etatcivil, argsPre cps.PresenceOffsets) error {
	equipier, err := cps.SelectEquipier(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, equipier.IdPersonne)
	if err != nil {
		return utils.SQLError(err)
	}
	camp, err := cps.SelectCamp(ct.db, equipier.IdCamp)
	if err != nil {
		return utils.SQLError(err)
	}

	// sanitize presence
	if argsPre.Debut-argsPre.Fin >= camp.Duree {
		return errors.New("invalid Presence")
	}

	equipier.FormStatus = cps.Answered
	equipier.Presence = argsPre

	personne.Etatcivil = argsP

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = personne.Update(tx)
		if err != nil {
			return err
		}
		_, err = equipier.Update(tx)
		if err != nil {
			return err
		}
		return nil
	})
}

func (ct *Controller) UpdateCharte(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[cps.IdEquipier](ct.key, token)
	if err != nil {
		return err
	}
	accept := utils.QueryParamBool(c, "accept")
	err = ct.updateCharte(id, accept)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateCharte(id cps.IdEquipier, accept bool) error {
	equipier, err := cps.SelectEquipier(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	equipier.AccepteCharte = sql.NullBool{Valid: true, Bool: accept}
	_, err = equipier.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) UploadDocument(c echo.Context) error {
	token := c.QueryParam("token")
	idEquipier, err := crypto.DecryptID[cps.IdEquipier](ct.key, token)
	if err != nil {
		return err
	}
	idDemande, err := utils.QueryParamInt[fs.IdDemande](c, "idDemande")
	if err != nil {
		return err
	}
	header, err := c.FormFile("document")
	if err != nil {
		return err
	}
	content, filename, err := filesAPI.ReadUpload(header)
	if err != nil {
		return err
	}

	filePub, err := ct.uploadDocument(idEquipier, idDemande, content, filename)
	if err != nil {
		return err
	}

	return c.JSON(200, filePub)
}

func (ct *Controller) uploadDocument(idEquipier cps.IdEquipier, idDemande fs.IdDemande, content []byte, filename string) (logic.PublicFile, error) {
	equipier, err := cps.SelectEquipier(ct.db, idEquipier)
	if err != nil {
		return logic.PublicFile{}, utils.SQLError(err)
	}

	file, err := filesAPI.SaveFileFor(ct.files, ct.db, equipier.IdPersonne, idDemande, content, filename)
	if err != nil {
		return logic.PublicFile{}, err
	}

	return logic.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DeleteDocument(c echo.Context) error {
	key := c.QueryParam("key")
	_, err := filesAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

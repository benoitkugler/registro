package espaceperso

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	filesAPI "registro/controllers/files"
	fsAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/controllers/services"
	"registro/crypto"
	"registro/mails"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// fiches sanitaires

type FichesanitaireExt struct {
	Personne             string
	IsLocked             bool
	State                pr.FichesanitaireState
	Fichesanitaire       pr.Fichesanitaire
	RespoTels            pr.Tels
	RespoSecuriteSociale string

	VaccinsDemande fs.Demande
	VaccinsFiles   []fsAPI.PublicFile
}

func (ct *Controller) LoadFichesanitaires(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	out, err := ct.loadFichesanitaires(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadFichesanitaires(id ds.IdDossier) (out []FichesanitaireExt, _ error) {
	dossier, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}
	responsable := dossier.Responsable()
	idPersonnes := dossier.Participants.IdPersonnes()

	// load existing vaccins
	vaccins, vaccinDemande, err := fsAPI.LoadVaccins(ct.db, ct.key, idPersonnes)
	if err != nil {
		return nil, err
	}

	fiches, err := pr.SelectFichesanitairesByIdPersonnes(ct.db, idPersonnes...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	fichesByPersonne := fiches.ByIdPersonne()
	// make sure the struct is initialized for every [Personne], restricting to < 18 years old
	for _, pers := range dossier.Personnes()[1:] {
		if pers.IsTemp { // wait for the validation
			continue
		}
		camp, ok := dossier.FirstCampFor(pers.Id)
		if !ok {
			continue
		}
		if camp.AgeDebutCamp(pers.DateNaissance) >= 18 {
			continue
		}

		fiche := fichesByPersonne[pers.Id]
		fiche.IdPersonne = pers.Id // init ID for empty fiche
		fsExt := FichesanitaireExt{
			pers.PrenomNOM(),
			isFichesanitaireLocked(responsable.Mail, fiche.Mails),
			fiche.State(dossier.Dossier.MomentInscription),
			fiche,
			responsable.Tels,
			responsable.SecuriteSociale,

			vaccinDemande,
			vaccins[pers.Id],
		}
		if fsExt.IsLocked { // hide sensitive information
			fsExt.Fichesanitaire = pr.Fichesanitaire{
				IdPersonne: fiche.IdPersonne,
				Mails:      fiche.Mails,
				LastModif:  fiche.LastModif,
			}
		}
		out = append(out, fsExt)
	}
	return out, nil
}

type UpdateFichesanitaireIn struct {
	Token          string
	Fichesanitaire pr.Fichesanitaire
	// for simplicity, it is updated for each participant
	SecuriteSocialeResponsable string
}

func (ct *Controller) UpdateFichesanitaire(c echo.Context) error {
	var args UpdateFichesanitaireIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateFichesanitaire(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func isFichesanitaireLocked(mailRespo string, mails []string) bool {
	if len(mails) == 0 {
		return false
	}
	for _, mail := range mails {
		if strings.ToLower(mail) == strings.ToLower(mailRespo) {
			return false
		}
	}
	return true
}

func (ct *Controller) updateFichesanitaire(args UpdateFichesanitaireIn) error {
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return err
	}
	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	responsable := dossier.Responsable()
	idPersonne := args.Fichesanitaire.IdPersonne
	// check Id is valid
	if !slices.Contains(dossier.Participants.IdPersonnes(), idPersonne) {
		return errors.New("access forbidden")
	}
	// for security concern, load the existing fs
	fs, _, err := pr.SelectFichesanitaireByIdPersonne(ct.db, idPersonne)
	if err != nil {
		return utils.SQLError(err)
	}

	if isFichesanitaireLocked(responsable.Mail, fs.Mails) {
		return errors.New("access forbidden")
	}

	// le premier responsable à modifier devient le proprio de la fiche
	if len(fs.Mails) == 0 {
		fs.Mails = []string{responsable.Mail}
	}
	args.Fichesanitaire.Mails = fs.Mails
	args.Fichesanitaire.LastModif = time.Now()

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		responsable.SecuriteSociale = args.SecuriteSocialeResponsable
		_, err = responsable.Update(tx)
		if err != nil {
			return err
		}
		_, err = pr.DeleteFichesanitairesByIdPersonnes(tx, idPersonne)
		if err != nil {
			return err
		}
		err = args.Fichesanitaire.Insert(tx)
		return err
	})
}

func (ct *Controller) UploadVaccin(c echo.Context) error {
	token := c.QueryParam("token")
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	idPersonne, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	header, err := c.FormFile("file")
	if err != nil {
		return err
	}
	content, name, err := fsAPI.ReadUpload(header)
	if err != nil {
		return err
	}
	out, err := ct.uploadVaccin(idDossier, idPersonne, content, name)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) uploadVaccin(idDossier ds.IdDossier, idPersonne pr.IdPersonne, content []byte, filename string) (fsAPI.PublicFile, error) {
	vaccinDemande, err := fsAPI.DemandeVaccin(ct.db)
	if err != nil {
		return fsAPI.PublicFile{}, err
	}

	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return fsAPI.PublicFile{}, err
	}
	// check Id is valid
	if !slices.Contains(dossier.Participants.IdPersonnes(), idPersonne) {
		return fsAPI.PublicFile{}, errors.New("access forbidden")
	}

	file, err := filesAPI.SaveFileFor(ct.files, ct.db, idPersonne, vaccinDemande.Id, content, filename)
	if err != nil {
		return fsAPI.PublicFile{}, err
	}

	return fsAPI.NewPublicFile(ct.key, file), nil
}

func (ct *Controller) DeleteVaccin(c echo.Context) error {
	key := c.QueryParam("key")
	err := fsAPI.Delete(ct.db, ct.key, ct.files, key)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// TransfertFicheSanitaire envoie un mail de demande de transfert
func (ct *Controller) TransfertFicheSanitaire(c echo.Context) error {
	token := c.QueryParam("token")
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	idPersonne, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	err = ct.transfertFicheSanitaire(c.Request().Host, idDossier, idPersonne)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

type transfertFicheSanitaireArgs struct {
	IdPersonne pr.IdPersonne // le profil à modifier
	NewMail    string        // le nouvel accès à autoriser
}

func (ct *Controller) transfertFicheSanitaire(host string, idDossier ds.IdDossier, idPersonne pr.IdPersonne) error {
	dossier, err := logic.LoadDossier(ct.db, idDossier)
	if err != nil {
		return err
	}
	// check Id is valid
	if !slices.Contains(dossier.Participants.IdPersonnes(), idPersonne) {
		return errors.New("access forbidden")
	}
	personne, err := pr.SelectPersonne(ct.db, idPersonne)
	if err != nil {
		return utils.SQLError(err)
	}
	fiche, ok, err := pr.SelectFichesanitaireByIdPersonne(ct.db, idPersonne)
	if err != nil {
		return utils.SQLError(err)
	}
	if !ok {
		return errors.New("internal error: missing Fichesanitaire")
	}
	newMail := dossier.Responsable().Mail
	token, err := ct.key.EncryptJSON(transfertFicheSanitaireArgs{idPersonne, newMail})
	if err != nil {
		return err
	}
	url := utils.BuildUrl(host, services.EndpointServices,
		utils.QPInt("service", services.TransfertFicheSanitaire),
		utils.QP("token", token),
	)
	html, err := mails.TransfertFicheSanitaire(ct.asso, url, newMail, personne.PrenomNOM())
	if err != nil {
		return err
	}
	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return err
	}
	defer pool.Close()
	for _, owner := range fiche.Mails {
		err = pool.SendMail(owner, "Partage d'une fiche sanitaire", html, nil, mails.DefaultReplyTo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ct *Controller) ValideTransfertFicheSanitaire(c echo.Context) error {
	token := c.QueryParam("token")
	var args transfertFicheSanitaireArgs
	if err := ct.key.DecryptJSON(token, &args); err != nil {
		return err
	}
	err := ct.valideTransfertFicheSanitaire(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) valideTransfertFicheSanitaire(args transfertFicheSanitaireArgs) error {
	fiche, found, err := pr.SelectFichesanitaireByIdPersonne(ct.db, args.IdPersonne)
	if err != nil {
		return err
	}
	if !found { // should not happen
		fiche = pr.Fichesanitaire{IdPersonne: args.IdPersonne}
	}
	if slices.Contains(fiche.Mails, args.NewMail) {
		return nil // nothing to do
	}
	fiche.Mails = append(fiche.Mails, args.NewMail)
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = pr.DeleteFichesanitairesByIdPersonnes(tx, args.IdPersonne)
		if err != nil {
			return err
		}
		err = fiche.Insert(tx)
		return err
	})
}

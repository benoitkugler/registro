package espaceperso

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"registro/controllers/services"
	"registro/crypto"
	"registro/logic"
	"registro/mails"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// fiches sanitaires

// only asks for fiche sanitaires and vaccins for majeur
func asksFichesanitaire(dossier logic.Dossier, personne pr.Personne) bool {
	camp, ok := dossier.FirstCampFor(personne.Id)
	if !ok {
		return false
	}
	return camp.AgeDebutCamp(personne.DateNaissance) < 18
}

type FichesanitaireExt struct {
	Personne       string
	IsLocked       bool
	State          pr.FichesanitaireState
	Fichesanitaire pr.Fichesanitaire

	ResponsableNom  string
	ResponsableTels pr.Tels
}

func loadFichesanitaires(db ds.DB, dossier logic.Dossier) (out []FichesanitaireExt, _ error) {
	responsable := dossier.Responsable()
	idPersonnes := dossier.Participants.IdPersonnes()

	fiches, err := pr.SelectFichesanitairesByIdPersonnes(db, idPersonnes...)
	if err != nil {
		return out, utils.SQLError(err)
	}
	fichesByPersonne := fiches.ByIdPersonne()
	// make sure the struct is initialized for every [Personne], restricting to < 18 years old
	for _, pers := range dossier.Personnes()[1:] {
		if pers.IsTemp { // wait for the validation
			continue
		}
		if !asksFichesanitaire(dossier, pers) {
			continue
		}

		fiche := fichesByPersonne[pers.Id]
		fiche.IdPersonne = pers.Id // init ID for empty fiche
		fsExt := FichesanitaireExt{
			pers.PrenomN(),
			isFichesanitaireLocked(responsable.Mail, fiche.Owners),
			fiche.State(dossier.Dossier.MomentInscription),
			fiche,
			responsable.PrenomNOM(),
			responsable.Tels,
		}
		if fsExt.IsLocked { // hide sensitive information
			fsExt.Fichesanitaire = pr.Fichesanitaire{
				IdPersonne: fiche.IdPersonne,
				Owners:     fiche.Owners,
				Modified:   fiche.Modified,
			}
		}
		out = append(out, fsExt)
	}
	return out, nil
}

type UpdateFichesanitaireIn struct {
	Token          string
	Fichesanitaire pr.Fichesanitaire
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

	if isFichesanitaireLocked(responsable.Mail, fs.Owners) {
		return errors.New("access forbidden")
	}

	// le premier responsable à modifier devient le proprio de la fiche
	if len(fs.Owners) == 0 {
		fs.Owners = []string{responsable.Mail}
	}
	args.Fichesanitaire.Owners = fs.Owners
	args.Fichesanitaire.Modified = time.Now()

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = pr.DeleteFichesanitairesByIdPersonnes(tx, idPersonne)
		if err != nil {
			return err
		}
		err = args.Fichesanitaire.Insert(tx)
		return err
	})
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
	for _, owner := range fiche.Owners {
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
	if slices.Contains(fiche.Owners, args.NewMail) {
		return nil // nothing to do
	}
	fiche.Owners = append(fiche.Owners, args.NewMail)
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = pr.DeleteFichesanitairesByIdPersonnes(tx, args.IdPersonne)
		if err != nil {
			return err
		}
		err = fiche.Insert(tx)
		return err
	})
}

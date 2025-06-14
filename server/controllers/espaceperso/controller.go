package espaceperso

import (
	"database/sql"
	"errors"
	"slices"
	"time"

	"registro/config"
	"registro/controllers/directeurs"
	filesAPI "registro/controllers/files"
	fsAPI "registro/controllers/files"
	"registro/crypto"
	"registro/generators/pdfcreator"
	"registro/joomeo"
	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// Une modification après J-7 (début du camp)
// entraine une notification par email
const updateLimitation = 7 * 24 * time.Hour

type Controller struct {
	db *sql.DB

	key    crypto.Encrypter
	smtp   config.SMTP
	asso   config.Asso
	files  fs.FileSystem
	joomeo config.Joomeo
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso, fs fs.FileSystem,
	joomeo config.Joomeo,
) *Controller {
	return &Controller{db, key, smtp, asso, fs, joomeo}
}

func (ct *Controller) Load(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return errors.New("Lien invalide.")
	}

	out, err := ct.load(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Data struct {
	Dossier                   logic.DossierExt
	DocumentsToFillCount      int // number to fill
	FichesanitaireToFillCount int
	IsPaiementOpen            bool
}

func (ct *Controller) load(id ds.IdDossier) (Data, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return Data{}, err
	}
	documents, err := loadDocuments(ct.db, ct.key, dossier.Dossier)
	if err != nil {
		return Data{}, err
	}
	fiches, err := ct.loadFichesanitaires(id)
	if err != nil {
		return Data{}, err
	}
	return Data{
		dossier.Publish(ct.key),
		documents.ToFillCount,
		fiches.ToFillCount,
		false, // TODO
	}, nil
}

type SendMessageIn struct {
	Token string

	Message string
}

// SendMessage inscrit un nouveau message, sans notifications
func (ct *Controller) SendMessage(c echo.Context) error {
	var args SendMessageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.sendMessage(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) sendMessage(args SendMessageIn) (logic.Event, error) {
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return logic.Event{}, err
	}
	event, message, err := events.CreateMessage(ct.db, id, time.Now(), args.Message, events.FromEspaceperso, events.OptIdCamp{})
	if err != nil {
		return logic.Event{}, utils.SQLError(err)
	}

	return logic.Event{
		Id:      event.Id,
		Created: event.Created,
		Content: logic.Message{
			Message: message,
		},
	}, nil
}

type UpdateParticipantsIn struct {
	Token        string
	Participants []cps.Participant
}

// UpdateParticipants met à jour les champs [Navette], [Commentaire] et
// [OptionPrix] de chaque participant donnés.
//
// Une notification est envoyée au directeur si le séjour approche.
func (ct *Controller) UpdateParticipants(c echo.Context) error {
	var args UpdateParticipantsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateParticipants(c.Request().Host, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateParticipants(host string, args UpdateParticipantsIn) error {
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return err
	}
	participants, err := cps.SelectParticipantsByIdDossiers(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	camps, err := cps.SelectCamps(ct.db, participants.IdCamps()...)
	if err != nil {
		return utils.SQLError(err)
	}
	tmp, err := cps.SelectEquipiersByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return utils.SQLError(err)
	}
	equipiersbyCamp := tmp.ByIdCamp()

	personnes, err := pr.SelectPersonnes(ct.db, append(tmp.IdPersonnes(), participants.IdPersonnes()...)...)
	if err != nil {
		return utils.SQLError(err)
	}

	urlDirecteur := utils.BuildUrl(host, directeurs.EndpointDirecteur)

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		participantsByCamp := map[cps.IdCamp][]string{}
		for _, newP := range args.Participants {
			current, ok := participants[newP.Id]
			if !ok {
				return errors.New("access forbidden")
			}
			current.Navette = newP.Navette
			current.Details = newP.Details
			current.OptionPrix = newP.OptionPrix // TODO: sanitize
			_, err = current.Update(tx)
			if err != nil {
				return err
			}
			participantsByCamp[current.IdCamp] = append(participantsByCamp[current.IdCamp], personnes[current.IdPersonne].PrenomNOM())
		}

		for idCamp, participants := range participantsByCamp {
			camp := camps[idCamp]
			dir, hasDir := equipiersbyCamp[camp.Id].Directeur()
			if time.Until(camp.DateDebut.Time()) < updateLimitation && hasDir {
				// send a notification
				directeur := personnes[dir.IdPersonne]

				html, err := mails.NotifieModificationOptions(ct.asso, directeur.Etatcivil, camp.Label(), participants, urlDirecteur)
				if err != nil {
					return err
				}
				err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(directeur.Mail, "Modification des options", html, nil, nil)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (ct *Controller) GetStructureaides(c echo.Context) error {
	out, err := cps.SelectAllStructureaides(ct.db)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// CreateAide déclare une aide (non validée).
func (ct *Controller) CreateAide(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}

	var args cps.Aide
	err = utils.FormValueJSON(c, "aide", &args)
	if err != nil {
		return err
	}

	header, err := c.FormFile("document")
	if err != nil {
		return err
	}

	content, name, err := fsAPI.ReadUpload(header)
	if err != nil {
		return err
	}

	err = ct.createAide(id, args, content, name)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) createAide(id ds.IdDossier, args cps.Aide, fileContent []byte, fileName string) error {
	participant, err := cps.SelectParticipant(ct.db, args.IdParticipant)
	if err != nil {
		return utils.SQLError(err)
	}
	if participant.IdDossier != id {
		return errors.New("access forbidden")
	}

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		aide, err := cps.Aide{
			IdStructureaide: args.IdStructureaide,
			IdParticipant:   args.IdParticipant,
			Valide:          false,
			Valeur:          args.Valeur,
			ParJour:         args.ParJour,
			NbJoursMax:      args.NbJoursMax,
		}.Insert(tx)
		if err != nil {
			return err
		}

		file, err := fs.File{}.Insert(tx)
		if err != nil {
			return err
		}
		err = fs.FileAide{IdFile: file.Id, IdAide: aide.Id}.Insert(tx)
		if err != nil {
			return err
		}

		_, err = fs.UploadFile(ct.files, tx, file.Id, fileContent, fileName)
		if err != nil {
			return err
		}

		return nil
	})
}

type Joomeo struct {
	SpaceURL string
	Loggin   string   // may be empty
	Password string   // may be empty
	Albums   []string // may be empty
}

func (ct *Controller) LoadJoomeo(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	out, err := ct.loadJoomeo(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadJoomeo(id ds.IdDossier) (Joomeo, error) {
	dossier, err := ds.SelectDossier(ct.db, id)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}
	responsable, err := pr.SelectPersonne(ct.db, dossier.IdResponsable)
	if err != nil {
		return Joomeo{}, utils.SQLError(err)
	}

	api, err := joomeo.InitApi(ct.joomeo)
	if err != nil {
		return Joomeo{}, err
	}
	defer api.Close()

	contact, albums, err := api.GetLoginFromMail(responsable.Mail)
	if err != nil {
		return Joomeo{}, err
	}

	albumsLabels := make([]string, len(albums))
	for i, album := range albums {
		albumsLabels[i] = album.Label
	}

	return Joomeo{
		SpaceURL: api.SpaceURL(),
		Loggin:   contact.Login,
		Password: contact.Password,
		Albums:   albumsLabels,
	}, nil
}

// Sondages

func (ct *Controller) LoadSondages(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	out, err := ct.loadSondages(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type SondageExt struct {
	Camp    string
	Sondage cps.Sondage // Id is zero for no answer
}

// Renvoie les sondages existant.
// Pour les séjours qui n'en n'ont pas encore, on renvoie
// un sondage vide si le camp est "ouvert au sondage" pour le dossier,
// et si au moins un participant est en liste principal.
func (ct *Controller) loadSondages(id ds.IdDossier) ([]SondageExt, error) {
	dossier, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}
	camps := dossier.Camps()
	participantsByCamp := dossier.Participants.ByIdCamp()

	campsToAdd := utils.NewSet(camps.IDs()...)
	sondages, err := cps.SelectSondagesByIdDossiers(ct.db, id)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	var out []SondageExt
	// 1 : always include existing sondages
	for _, sondage := range sondages {
		out = append(out, SondageExt{camps[sondage.IdCamp].Label(), sondage})
		campsToAdd.Delete(sondage.IdCamp)
	}
	// 2 : complete with "open" camps
	for event := range logic.IterContentBy[logic.Sondage](dossier.Events) {
		if !campsToAdd.Has(event.Content.IdCamp) { // already included
			continue
		}
		hasInscrit := false
		for _, part := range participantsByCamp[event.Content.IdCamp] {
			if part.Statut == cps.Inscrit {
				hasInscrit = true
				break
			}
		}
		if !hasInscrit {
			continue
		}
		out = append(out, SondageExt{camps[event.Content.IdCamp].Label(), cps.Sondage{
			IdCamp:    event.Content.IdCamp,
			IdDossier: id,
		}})
	}
	// sort to avoid strange switch on the client
	slices.SortFunc(out, func(a, b SondageExt) int { return int(a.Sondage.IdCamp - b.Sondage.IdCamp) })
	return out, nil
}

type UpdateSondageIn struct {
	Token string

	Id      cps.IdSondage // may be 0
	IdCamp  cps.IdCamp    // ignored if Id != 0
	Reponse cps.ReponseSondage
}

func (ct *Controller) UpdateSondages(c echo.Context) error {
	var args UpdateSondageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	idDossier, err := crypto.DecryptID[ds.IdDossier](ct.key, args.Token)
	if err != nil {
		return err
	}
	err = ct.updateSondage(idDossier, args.Id, args.IdCamp, args.Reponse)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateSondage(idDossier ds.IdDossier, idSondage cps.IdSondage, idCamp cps.IdCamp, reponse cps.ReponseSondage,
) error {
	if idSondage == 0 { // create mode
		_, err := cps.Sondage{
			IdCamp:         idCamp,
			IdDossier:      idDossier,
			Modified:       time.Now(),
			ReponseSondage: reponse,
		}.Insert(ct.db)
		if err != nil {
			return utils.SQLError(err)
		}
	} else { // update mode
		existing, err := cps.SelectSondage(ct.db, idSondage)
		if err != nil {
			return utils.SQLError(err)
		}
		existing.Modified = time.Now()
		existing.ReponseSondage = reponse
		_, err = existing.Update(ct.db)
		if err != nil {
			return utils.SQLError(err)
		}
	}
	return nil
}

func (ct *Controller) DownloadAttestationPresence(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	content, err := ct.renderAttestationPresence(id)
	if err != nil {
		return err
	}
	mimeType := filesAPI.SetBlobHeader(c, content, "Attestation de présence.pdf")
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) renderAttestationPresence(id ds.IdDossier) ([]byte, error) {
	dossier, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}
	// restrict to inscrits with started camp
	var filtered []cps.ParticipantCamp
	for _, p := range dossier.ParticipantsExt() {
		if p.Participant.Statut != cps.Inscrit {
			continue
		}
		if hasStarted := p.Camp.DateDebut.Time().Before(time.Now()); !hasStarted {
			continue
		}
		filtered = append(filtered, p)
	}
	responsable := dossier.Responsable()
	destinataire := pdfcreator.Destinataire{
		NomPrenom:  responsable.NOMPrenom(),
		Adresse:    responsable.Adresse,
		CodePostal: responsable.CodePostal,
		Ville:      responsable.Ville,
	}
	content, err := pdfcreator.CreateAttestationPresence(ct.asso, destinataire, filtered)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (ct *Controller) DownloadFacture(c echo.Context) error {
	token := c.QueryParam("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	if err != nil {
		return err
	}
	content, err := ct.renderFacture(id)
	if err != nil {
		return err
	}
	mimeType := filesAPI.SetBlobHeader(c, content, "Facture.pdf")
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) renderFacture(id ds.IdDossier) ([]byte, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return nil, err
	}
	responsable := dossier.Responsable()
	destinataire := pdfcreator.Destinataire{
		NomPrenom:  responsable.NOMPrenom(),
		Adresse:    responsable.Adresse,
		CodePostal: responsable.CodePostal,
		Ville:      responsable.Ville,
	}
	// restrict to inscrits with started camp
	var filtered []cps.ParticipantCamp
	for _, p := range dossier.ParticipantsExt() {
		if p.Participant.Statut != cps.Inscrit {
			continue
		}
		filtered = append(filtered, p)
	}
	// sort by time
	finances := dossier.Publish(ct.key)
	var paiements []ds.Paiement
	for _, p := range finances.Paiements {
		paiements = append(paiements, p)
	}
	slices.SortFunc(paiements, func(a, b ds.Paiement) int { return a.Time.Compare(b.Time) })

	content, err := pdfcreator.CreateFacture(ct.asso, destinataire, filtered, finances.Bilan, paiements)
	if err != nil {
		return nil, err
	}
	return content, nil
}

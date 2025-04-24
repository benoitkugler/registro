package espaceperso

import (
	"database/sql"
	"errors"
	"time"

	"registro/config"
	"registro/controllers/directeurs"
	fsAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/crypto"
	"registro/joomeo"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	"registro/sql/files"
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
	fs     files.FileSystem
	joomeo config.Joomeo
}

func NewController(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso, fs files.FileSystem,
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
	Dossier logic.DossierExt
}

func (ct *Controller) load(id ds.IdDossier) (Data, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return Data{}, err
	}

	return Data{
		dossier.Publish(ct.key),
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

		file, err := files.File{}.Insert(tx)
		if err != nil {
			return err
		}
		err = files.FileAide{IdFile: file.Id, IdAide: aide.Id}.Insert(tx)
		if err != nil {
			return err
		}

		_, err = files.UploadFile(ct.fs, tx, file.Id, fileContent, fileName)
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

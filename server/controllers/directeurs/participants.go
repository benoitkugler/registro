package directeurs

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	fsAPI "registro/controllers/files"
	"registro/generators/pdfcreator"
	"registro/generators/sheets"
	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) ParticipantsGet(c echo.Context) error {
	user := JWTUser(c)

	out, err := ct.getParticipants(user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type ParticipantsOut struct {
	Participants []logic.ParticipantExt
	Dossiers     map[ds.IdDossier]logic.DossierReglement
}

func (ct *Controller) getParticipants(id cps.IdCamp) (ParticipantsOut, error) {
	participants, dossiers, _, err := logic.LoadParticipants(ct.db, id)
	if err != nil {
		return ParticipantsOut{}, err
	}
	finances, err := logic.LoadDossiersFinances(ct.db, dossiers.IDs()...)
	if err != nil {
		return ParticipantsOut{}, err
	}
	reglements := make(map[ds.IdDossier]logic.DossierReglement)
	for id := range dossiers {
		dossier := finances.For(id)
		reglements[id] = dossier.Reglement()
	}
	return ParticipantsOut{participants, reglements}, nil
}

// ParticipantsUpdate modifie les champs d'un participant.
//
// Seuls les champs Details et Navette sont pris en compte.
//
// Le statut est modifié sans aucune notification.
func (ct *Controller) ParticipantsUpdate(c echo.Context) error {
	var args cps.Participant
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateParticipant(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateParticipant(args cps.Participant) error {
	current, err := cps.SelectParticipant(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	current.Commentaire = args.Commentaire
	current.Navette = args.Navette
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) ParticipantsGetFichesSanitaires(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.loadFichesSanitaires(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type FicheSanitaireExt struct {
	IdParticipant cps.IdParticipant
	Personne      string
	State         pr.FichesanitaireState
	Fiche         pr.Fichesanitaire
	// Vaccins       []logic.PublicFile
}

func (ct *Controller) loadFichesSanitaires(user cps.IdCamp) ([]FicheSanitaireExt, error) {
	participants, err := cps.SelectParticipantsByIdCamps(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers, err := ds.SelectDossiers(ct.db, participants.IdDossiers()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	pIds := participants.IdPersonnes()
	personnes, err := pr.SelectPersonnes(ct.db, pIds...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	tmp, err := pr.SelectFichesanitairesByIdPersonnes(ct.db, pIds...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	fiches := tmp.ByIdPersonne()

	out := make([]FicheSanitaireExt, 0, len(personnes))
	for _, participant := range participants {
		personne := personnes[participant.IdPersonne]
		fiche := fiches[personne.Id]
		dossier := dossiers[participant.IdDossier]
		out = append(out, FicheSanitaireExt{
			participant.Id,
			personne.NOMPrenom(),
			fiche.State(dossier.MomentInscription),
			fiche,
		})
	}

	slices.SortFunc(out, func(a, b FicheSanitaireExt) int { return strings.Compare(a.Personne, b.Personne) })

	return out, nil
}

func (ct *Controller) ParticipantsDownloadFicheSanitaire(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[cps.IdParticipant](c, "idParticipant")
	if err != nil {
		return err
	}
	content, name, err := ct.downloadFicheSanitaire(user, id)
	if err != nil {
		return err
	}
	mimeType := fsAPI.SetBlobHeader(c, content, name)
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) downloadFicheSanitaire(user cps.IdCamp, id cps.IdParticipant) ([]byte, string, error) {
	// check the access is legal
	participant, err := cps.SelectParticipant(ct.db, id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	if participant.IdCamp != user {
		return nil, "", errors.New("access forbidden")
	}
	dossier, err := ds.SelectDossier(ct.db, participant.IdDossier)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	responsable, err := pr.SelectPersonne(ct.db, dossier.IdResponsable)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, participant.IdPersonne)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	fiche, _, err := pr.SelectFichesanitaireByIdPersonne(ct.db, personne.Id)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	content, err := pdfcreator.CreateFicheSanitaires(ct.asso, []pdfcreator.FicheSanitaire{
		{Personne: personne.Etatcivil, FicheSanitaire: fiche, Responsable: responsable.Etatcivil},
	})
	name := fmt.Sprintf("Fiche sanitaire %s.pdf", personne.NOMPrenom())
	return content, name, nil
}

func (ct *Controller) ParticipantsMessagesLoad(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.loadMessages(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Messages struct {
	Messages         []MessageExt // sorted by time, most recent first
	Dossiers         map[ds.IdDossier]DossierPersonnes
	NewMessagesCount int
}

func (m *Messages) setNewMessagesCount(idCamp cps.IdCamp) {
	count := 0
	for _, message := range m.Messages {
		isOwn := message.Content.Message.OrigineCamp.Is(idCamp)
		wasSeen := slices.Contains(message.Content.VuParCampsIDs, idCamp)
		if !isOwn && !wasSeen {
			count++
		}
	}
	m.NewMessagesCount = count
}

type MessageExt = logic.EventExt[logic.Message]

type DossierPersonnes struct {
	Responsable  string
	Participants []string
}

func (ct *Controller) loadMessages(idCamp cps.IdCamp) (Messages, error) {
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return Messages{}, err
	}
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers()...)
	if err != nil {
		return Messages{}, err
	}
	out := Messages{Dossiers: make(map[ds.IdDossier]DossierPersonnes)}
	for idDossier := range dossiers.Dossiers {
		dossier := dossiers.For(idDossier)

		// hide fonds de soutien
		for message := range logic.IterEventsBy[logic.Message](dossier.Events) {
			if message.Content.Message.Origine == evs.FondSoutien || message.Content.Message.OnlyToFondSoutien {
				continue
			}
			out.Messages = append(out.Messages, message)
		}

		item := DossierPersonnes{
			Responsable: dossier.Responsable().NOMPrenom(),
		}
		// only show participants of this camp
		for _, part := range dossier.ParticipantsExt() {
			if part.Participant.IdCamp != idCamp {
				continue
			}
			item.Participants = append(item.Participants, part.Personne.PrenomN())
		}
		out.Dossiers[idDossier] = item
	}

	slices.SortFunc(out.Messages, func(a, b MessageExt) int { return b.Event.Created.Compare(a.Event.Created) })

	out.setNewMessagesCount(idCamp)
	return out, nil
}

func (ct *Controller) ParticipantsMessageSetSeen(c echo.Context) error {
	user := JWTUser(c)
	idMessage, err := utils.QueryParamInt[evs.IdEvent](c, "idEvent")
	if err != nil {
		return err
	}
	seen := utils.QueryParamBool(c, "seen")
	// TODO: we should check that this camp has acces to the message
	out, err := ct.setMessageSeen(user, idMessage, seen)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) setMessageSeen(user cps.IdCamp, idEvent evs.IdEvent, seen bool) (MessageExt, error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		// always delete to avoid unique constraint error
		err := evs.EventMessageVu{IdEvent: idEvent, IdCamp: user}.Delete(tx)
		if err != nil {
			return err
		}
		if seen {
			err = evs.EventMessageVu{IdEvent: idEvent, IdCamp: user}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return MessageExt{}, err
	}

	return loadMessage(ct.db, idEvent)
}

func loadMessage(db evs.DB, id evs.IdEvent) (MessageExt, error) {
	event, err := logic.LoadEvent(db, id)
	if err != nil {
		return MessageExt{}, err
	}
	m, ok := event.Content.(logic.Message)
	if !ok { // should never happen since the event has Kind == Message
		return MessageExt{}, errors.New("internal error : expected Message")
	}
	return MessageExt{Event: event.Raw(), Content: m}, nil
}

type CreateMessageIn struct {
	Contenu   string
	IdDossier ds.IdDossier
}

func (ct *Controller) ParticipantsMessagesCreate(c echo.Context) error {
	user := JWTUser(c)

	var args CreateMessageIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createMessage(c.Request().Host, user, args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createMessage(host string, idCamp cps.IdCamp, args CreateMessageIn) (MessageExt, error) {
	dossier, err := logic.LoadDossier(ct.db, args.IdDossier)
	if err != nil {
		return MessageExt{}, utils.SQLError(err)
	}
	url := logic.EspacePersoURL(ct.key, host, args.IdDossier)
	var event evs.Event
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		event, _, err = evs.CreateMessage(tx, args.IdDossier, time.Now(), evs.EventMessage{Contenu: args.Contenu, Origine: evs.Directeur, OrigineCamp: idCamp.Opt()})
		if err != nil {
			return err
		}
		// notifie le responsable
		resp := dossier.Responsable()
		body, err := mails.NotifieMessage(ct.asso, mails.NewContact(&resp), args.Contenu, url)
		if err != nil {
			return err
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(resp.Mail, "Nouveau message", body, dossier.Dossier.CopiesMails, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return MessageExt{}, err
	}
	return loadMessage(ct.db, event.Id)
}

func (ct *Controller) ParticipantsDownloadListe(c echo.Context) error {
	user := JWTUser(c)
	content, name, err := ct.exportListeParticipants(user)
	if err != nil {
		return err
	}
	mimeType := fsAPI.SetBlobHeader(c, content, name)
	return c.Blob(200, mimeType, content)
}

func (ct *Controller) exportListeParticipants(user cps.IdCamp) ([]byte, string, error) {
	camp, err := cps.LoadCamp(ct.db, user)
	if err != nil {
		return nil, "", err
	}
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, user)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	links, err := cps.SelectGroupeParticipantsByIdCamps(ct.db, user)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	participantToGroupe := make(map[cps.IdParticipant]cps.Groupe)
	for _, link := range links {
		participantToGroupe[link.IdParticipant] = groupes[link.IdGroupe]
	}
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers()...)
	if err != nil {
		return nil, "", err
	}
	showNationnaliteSuisse := ct.asso.AskNationnalite
	content, err := sheets.ListeParticipantsCamp(camp.Camp, camp.Participants(true), dossiers, participantToGroupe, showNationnaliteSuisse)
	if err != nil {
		return nil, "", err
	}
	name := fmt.Sprintf("Participants %s.xlsx", camp.Camp.Label())
	return content, name, nil
}

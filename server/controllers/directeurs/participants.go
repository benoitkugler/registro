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
	sh "registro/sql/shared"
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
	Statistiques cps.StatistiquesInscrits
}

func (ct *Controller) getParticipants(id cps.IdCamp) (ParticipantsOut, error) {
	participants, dossiers, camp, err := logic.LoadParticipants(ct.db, id)
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

	return ParticipantsOut{participants, reglements, camp.Stats()}, nil
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

type MessageExt = logic.EventExt[logic.MessageEvt]

type DossierPersonnes struct {
	Responsable  string
	Participants []string
}

func (ct *Controller) loadMessages(idCamp cps.IdCamp) (Messages, error) {
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return Messages{}, err
	}
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers())
	if err != nil {
		return Messages{}, err
	}
	out := Messages{Dossiers: make(map[ds.IdDossier]DossierPersonnes)}
	for idDossier := range dossiers.Dossiers {
		dossier := dossiers.For(idDossier)

		// hide fonds de soutien
		for message := range logic.IterEventsBy[logic.MessageEvt](dossier.Events) {
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
	m, ok := event.Content.(logic.MessageEvt)
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
	dossiers, err := logic.LoadDossiers(ct.db, camp.IdDossiers())
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

type GroupesOut struct {
	Groupes              cps.Groupes
	ParticipantsToGroupe map[cps.IdParticipant]cps.GroupeParticipant
	MinHint, MaxHint     sh.Date
}

func (ct *Controller) GroupesGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getGroupes(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getGroupes(id cps.IdCamp) (GroupesOut, error) {
	camp, err := cps.LoadCamp(ct.db, id)
	if err != nil {
		return GroupesOut{}, err
	}
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, id)
	if err != nil {
		return GroupesOut{}, utils.SQLError(err)
	}
	participantsGroupes, err := cps.SelectGroupeParticipantsByIdCamps(ct.db, id)
	if err != nil {
		return GroupesOut{}, utils.SQLError(err)
	}
	min, _, max := groupesRangeHint(camp, groupes)

	return GroupesOut{groupes, participantsGroupes.ByIdParticipant(), min, max}, nil
}

func groupesRangeHint(camp cps.CampData, groupes cps.Groupes) (min, toCreate, max sh.Date) {
	inscrits := camp.Participants(true)

	if len(inscrits) == 0 { // default to the Camp age range
		ageMin := camp.Camp.AgeMin
		ageMax := camp.Camp.AgeMax
		if ageMin <= 0 {
			ageMin = 6
		}
		if ageMax <= 0 {
			ageMax = 18
		}
		ageMiddle := (ageMin + ageMax) / 2

		max = camp.Camp.DateDebut.AddDays(-ageMin * 365)
		min = camp.Camp.DateDebut.AddDays(-ageMax * 365)
		toCreate = camp.Camp.DateDebut.AddDays(-ageMiddle * 365)
	} else {
		// use inscrits
		minT := inscrits[0].Personne.DateNaissance.Time()
		maxT := minT
		for _, inscrit := range inscrits {
			d := inscrit.Personne.DateNaissance.Time()
			if d.Before(minT) {
				minT = d
			}
			if maxT.Before(d) {
				maxT = d
			}
		}
		middleT := minT.Add(maxT.Sub(minT) / 2)
		min, toCreate, max = sh.NewDateFrom(minT), sh.NewDateFrom(middleT), sh.NewDateFrom(maxT)
	}

	// always includes already defined groupes
	for _, g := range groupes {
		minToCreate := g.Fin.AddDays(365)
		fin := g.Fin.Time()
		if fin.Before(min.Time()) {
			min = g.Fin
		}
		if max.Time().Before(fin) {
			max = g.Fin
		}
		if toCreate.Time().Before(minToCreate.Time()) {
			toCreate = minToCreate
		}
	}

	return
}

func (ct *Controller) GroupeCreate(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.createGroupe(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createGroupe(idCamp cps.IdCamp) (cps.Groupe, error) {
	// for a better UX, choose a decent default Fin
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return cps.Groupe{}, err
	}
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, idCamp)
	if err != nil {
		return cps.Groupe{}, utils.SQLError(err)
	}
	_, middle, _ := groupesRangeHint(camp, groupes)

	out, err := cps.Groupe{
		IdCamp:  idCamp,
		Nom:     "Groupe " + utils.RandString(6, false),
		Couleur: utils.RandColor(),
		Fin:     middle,
	}.Insert(ct.db)
	if err != nil {
		return out, utils.SQLError(err)
	}
	return out, nil
}

// GroupeUpdate does not update [Plage], see [GroupeUpdatePlages]
func (ct *Controller) GroupeUpdate(c echo.Context) error {
	user := JWTUser(c)
	var args cps.Groupe
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateGroupe(user, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateGroupe(user cps.IdCamp, args cps.Groupe) error {
	current, err := cps.SelectGroupe(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	if current.IdCamp != user {
		return errors.New("access forbidden")
	}
	current.Nom = args.Nom
	current.Couleur = args.Couleur
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) GroupeDelete(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[cps.IdGroupe](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteGroupe(user, id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteGroupe(user cps.IdCamp, id cps.IdGroupe) error {
	groupe, err := cps.SelectGroupe(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	if groupe.IdCamp != user {
		return errors.New("access forbidden")
	}
	_, err = cps.DeleteGroupeById(ct.db, id) // links cascade
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

type UpdateFinsIn struct {
	Fins map[cps.IdGroupe]sh.Date
	// if OverrideManuel is true, even the participant
	// with a manually affected groupe are updated.
	OverrideManuel bool
}

// GroupeUpdatePlages modifie les plages de tous les groupes donnés,
// et met à jour les affectations des inscrits.
func (ct *Controller) GroupeUpdatePlages(c echo.Context) error {
	user := JWTUser(c)
	var args UpdateFinsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateGroupesPlages(user, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateGroupesPlages(idCamp cps.IdCamp, args UpdateFinsIn) error {
	// affectations courantes
	camp, err := cps.LoadCamp(ct.db, idCamp)
	if err != nil {
		return err
	}
	links, err := cps.SelectGroupeParticipantsByIdCamps(ct.db, idCamp)
	if err != nil {
		return utils.SQLError(err)
	}
	byParticipant := links.ByIdParticipant()

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		for idGroupe, fin := range args.Fins {
			current, err := cps.SelectGroupe(ct.db, idGroupe)
			if err != nil {
				return err
			}
			if current.IdCamp != idCamp {
				return errors.New("access forbidden")
			}
			current.Fin = fin
			_, err = current.Update(tx)
			if err != nil {
				return err
			}
		}

		groupes, err := cps.SelectGroupesByIdCamps(tx, idCamp)
		if err != nil {
			return err
		}

		// compute the new affectations
		var newLinks cps.GroupeParticipants
		for _, participant := range camp.Participants(true) {
			// should we skip updating ?
			if current := byParticipant[participant.Participant.Id]; current.Manuel && !args.OverrideManuel {
				newLinks = append(newLinks, current)
			} else if newGroupe, ok := groupes.TrouveGroupe(participant.Personne.DateNaissance); ok {
				newLinks = append(newLinks, cps.GroupeParticipant{IdParticipant: participant.Participant.Id, IdGroupe: newGroupe.Id, IdCamp: idCamp})
			}
		}

		_, err = cps.DeleteGroupeParticipantsByIdCamps(tx, idCamp)
		if err != nil {
			return err
		}
		err = cps.InsertManyGroupeParticipants(tx, newLinks...)
		return err
	})
}

func (ct *Controller) ParticipantSetGroupe(c echo.Context) error {
	user := JWTUser(c)
	idParticipant, err := utils.QueryParamInt[cps.IdParticipant](c, "idParticipant")
	if err != nil {
		return err
	}
	idGroupe, err := utils.QueryParamInt[cps.IdGroupe](c, "idGroupe") // <= 0 to remove the groupe
	if err != nil {
		return err
	}
	err = ct.setParticipantGroupe(user, idParticipant, idGroupe)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) setParticipantGroupe(user cps.IdCamp, idParticipant cps.IdParticipant, idGroupe cps.IdGroupe) error {
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err := cps.DeleteGroupeParticipantsByIdParticipants(ct.db, idParticipant)
		if err != nil {
			return err
		}
		if idGroupe > 0 {
			err = cps.GroupeParticipant{IdCamp: user, IdParticipant: idParticipant, IdGroupe: idGroupe, Manuel: true}.Insert(ct.db)
			return err
		}
		return nil
	})
}

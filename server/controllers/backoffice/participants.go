package backoffice

import (
	"database/sql"
	"errors"
	"time"

	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type CampsLoadOut struct {
	Camp         cps.CampExt
	Participants []logic.ParticipantExt
}

func (ct *Controller) CampsLoad(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	out, err := ct.getParticipants(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getParticipants(id cps.IdCamp) (CampsLoadOut, error) {
	participants, _, camp, err := logic.LoadParticipants(ct.db, id)
	if err != nil {
		return CampsLoadOut{}, err
	}
	return CampsLoadOut{Camp: camp, Participants: participants}, nil
}

type ParticipantsCreateIn struct {
	IdDossier  ds.IdDossier
	IdCamp     cps.IdCamp
	IdPersonne pr.IdPersonne
}

// ParticipantsCreate ajoute un participant au séjour donné,
// en résolvant statut et groupe.
func (ct *Controller) ParticipantsCreate(c echo.Context) error {
	var args ParticipantsCreateIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createParticipant(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// checkParticipantDouble returns a better error message if already present
func checkParticipantDouble(db cps.DB, idCamp cps.IdCamp, idPersonne pr.IdPersonne) error {
	_, alreadyHere, err := cps.SelectParticipantByIdCampAndIdPersonne(db, idCamp, idPersonne)
	if err != nil {
		return utils.SQLError(err)
	}
	if alreadyHere {
		return errors.New("Ce profil est déjà présent sur ce camp !")
	}
	return nil
}

func (ct *Controller) createParticipant(args ParticipantsCreateIn) (logic.ParticipantExt, error) {
	dossier, err := ds.SelectDossier(ct.db, args.IdDossier)
	if err != nil {
		return logic.ParticipantExt{}, utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, args.IdPersonne)
	if err != nil {
		return logic.ParticipantExt{}, utils.SQLError(err)
	}

	if err := checkParticipantDouble(ct.db, args.IdCamp, args.IdPersonne); err != nil {
		return logic.ParticipantExt{}, err
	}

	// resolve Groupe...
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, args.IdCamp)
	if err != nil {
		return logic.ParticipantExt{}, utils.SQLError(err)
	}
	groupe, hasGroupe := groupes.TrouveGroupe(personne.DateNaissance)

	// ... and Statut
	camp, err := cps.LoadCampPersonnes(ct.db, args.IdCamp)
	if err != nil {
		return logic.ParticipantExt{}, err
	}
	statut := camp.Status([]pr.Personne{personne})[0]
	participant := cps.Participant{
		IdDossier:  args.IdDossier,
		IdCamp:     args.IdCamp,
		IdPersonne: args.IdPersonne,

		IdTaux: dossier.IdTaux,
		Statut: statut.Hint(),
	}

	// if the dossier is empty (for instance if manually created), we want to allow
	// a different taux (the one of the camp) to be used
	existingP, err := cps.SelectParticipantsByIdDossiers(ct.db, args.IdDossier)
	if err != nil {
		return logic.ParticipantExt{}, utils.SQLError(err)
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		if len(existingP) == 0 {
			// update the dossier ...
			dossier.IdTaux = camp.Camp.IdTaux
			_, err = dossier.Update(tx)
			if err != nil {
				return err
			}
			// ... and use this taux
			participant.IdTaux = camp.Camp.IdTaux
		}

		participant, err = participant.Insert(tx)
		if err != nil {
			return err
		}
		if hasGroupe {
			err = cps.GroupeParticipant{IdGroupe: groupe.Id, IdCamp: groupe.IdCamp, IdParticipant: participant.Id}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return logic.NewParticipantExt(participant, personne, camp.Camp, dossier), err
}

// ParticipantsUpdate modifie les champs d'un participant.
//
// Les champs [IdTaux] et [IdCamp] sont ignorés.
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
	current.IdPersonne = args.IdPersonne
	current.IdDossier = args.IdDossier
	current.Statut = args.Statut
	current.Remises = args.Remises
	current.QuotientFamilial = args.QuotientFamilial
	current.OptionPrix = args.OptionPrix
	current.Commentaire = args.Commentaire
	current.Navette = args.Navette
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

// ParticipantsDelete supprime le participant donné.
// Si la personne liée est temporaire, elle est aussi supprimée.
// Si le participant n'est pas encore validé et que la personne
// n'est pas référencé ailleurs, elle est aussi supprimée.
func (ct *Controller) ParticipantsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdParticipant](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteParticipant(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteParticipant(id cps.IdParticipant) error {
	// cleanup aides files and temp personne; the other items will cascade
	aides, err := cps.SelectAidesByIdParticipants(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	links, err := files.SelectFileAidesByIdAides(ct.db, aides.IDs()...)
	if err != nil {
		return utils.SQLError(err)
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		events, err := evs.DeleteEventPlaceLibereesByIdParticipants(tx, id)
		if err != nil {
			return err
		}
		_, err = evs.DeleteEventsByIDs(tx, events.IdEvents()...)
		if err != nil {
			return err
		}

		deleted, err := files.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}
		err = ct.files.Delete(deleted...)
		if err != nil {
			return err
		}

		participant, err := cps.DeleteParticipantById(tx, id)
		if err != nil {
			return err
		}
		dossier, err := ds.SelectDossier(ct.db, participant.IdDossier)
		if err != nil {
			return err
		}
		personne, err := pr.SelectPersonne(tx, participant.IdPersonne)
		if err != nil {
			return err
		}
		refs, err := logic.CheckPersonneReferences(tx, personne.Id)
		if err != nil {
			return err
		}
		if personne.IsTemp || (!dossier.IsValidated && refs.Empty()) { // cleanup
			_, err = pr.DeletePersonneById(tx, personne.Id)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type ParticipantsMoveIn struct {
	Id     cps.IdParticipant
	Target cps.IdCamp
}

// ParticipantsMove change le participant donné de camp,
// calculant automatiquement un groupe et un statut.
func (ct *Controller) ParticipantsMove(c echo.Context) error {
	var args ParticipantsMoveIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.moveParticipant(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) moveParticipant(args ParticipantsMoveIn) error {
	participant, err := cps.SelectParticipant(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, participant.IdPersonne)
	if err != nil {
		return utils.SQLError(err)
	}

	if participant.IdCamp == args.Target {
		return errors.New("same Camp")
	}

	// make sure the profil is not already in the target camp
	if err := checkParticipantDouble(ct.db, args.Target, participant.IdPersonne); err != nil {
		return err
	}

	// resolve Groupe...
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, args.Target)
	if err != nil {
		return utils.SQLError(err)
	}
	groupe, hasGroupe := groupes.TrouveGroupe(personne.DateNaissance)

	// ... and Statut
	camp, err := cps.LoadCampPersonnes(ct.db, args.Target)
	if err != nil {
		return err
	}
	statut := camp.Status([]pr.Personne{personne})[0].Hint()

	// also reset options which wont match the new camp

	participant.IdCamp = args.Target
	participant.Statut = statut
	participant.OptionPrix = cps.OptionPrixParticipant{}

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		participant, err = participant.Insert(tx)
		if err != nil {
			return err
		}
		if hasGroupe {
			err = cps.GroupeParticipant{IdGroupe: groupe.Id, IdCamp: groupe.IdCamp, IdParticipant: participant.Id}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// ParticipantsSetPlaceLiberee change the participant status, creates an event,
// and send a mail notification.
func (ct *Controller) ParticipantsSetPlaceLiberee(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdParticipant](c, "id")
	if err != nil {
		return err
	}
	out, err := ct.setPlaceLiberee(c.Request().Host, id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) setPlaceLiberee(host string, id cps.IdParticipant) (cps.Participant, error) {
	participant, err := cps.SelectParticipant(ct.db, id)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}
	dossier, responsable, err := dossierAndResp(ct.db, participant.IdDossier)
	if err != nil {
		return cps.Participant{}, err
	}
	camp, err := cps.LoadCampPersonnes(ct.db, participant.IdCamp)
	if err != nil {
		return cps.Participant{}, err
	}

	if participant.Statut == cps.Inscrit || participant.Statut == cps.EnAttenteReponse {
		return cps.Participant{}, errors.New("invalid Statut")
	}

	participant.Statut = cps.EnAttenteReponse

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		ev, err := evs.Event{IdDossier: participant.IdDossier, Kind: evs.PlaceLiberee, Created: time.Now()}.Insert(tx)
		if err != nil {
			return err
		}
		err = evs.EventPlaceLiberee{IdEvent: ev.Id, IdParticipant: participant.Id}.Insert(tx)
		if err != nil {
			return err
		}
		participant, err = participant.Update(tx)
		if err != nil {
			return err
		}
		// notifie par mail
		url := logic.URLEspacePerso(ct.key, host, participant.IdDossier,
			utils.QPInt("idEvent", ev.Id))
		html, err := mails.NotifiePlaceLiberee(ct.asso, mails.NewContact(&responsable), camp.Camp.Label(), url)
		if err != nil {
			return err
		}
		err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(responsable.Mail,
			"Place disponible", html, dossier.CopiesMails, nil)
		if err != nil {
			return err
		}

		return nil
	})

	return participant, err
}

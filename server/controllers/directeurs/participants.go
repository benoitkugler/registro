package directeurs

import (
	"errors"
	"fmt"

	fsAPI "registro/controllers/files"
	"registro/controllers/logic"
	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
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

func (ct *Controller) getParticipants(id cps.IdCamp) ([]logic.ParticipantExt, error) {
	participants, _, err := logic.LoadParticipants(ct.db, id)
	if err != nil {
		return nil, err
	}
	return participants, nil
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
	current.Details = args.Details
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
	Personne string
	State    pr.FichesanitaireState
	Fiche    pr.Fichesanitaire
	Vaccins  []fsAPI.PublicFile
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
	// load the vaccins
	vaccins, _, err := fsAPI.LoadVaccins(ct.db, ct.key, pIds)
	if err != nil {
		return nil, err
	}
	out := make([]FicheSanitaireExt, 0, len(personnes))
	for _, participant := range participants {
		personne := personnes[participant.IdPersonne]
		fiche := fiches[personne.Id]
		dossier := dossiers[participant.IdDossier]
		out = append(out, FicheSanitaireExt{
			Personne: personne.NOMPrenom(),
			State:    fiche.State(dossier.MomentInscription),
			Fiche:    fiche,
			Vaccins:  vaccins[personne.Id],
		})
	}
	return out, nil
}

func (ct *Controller) ParticipantDownloadFicheSanitaire(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[cps.IdParticipant](c, "idParticipant")
	if err != nil {
		return err
	}
	content, name, err := ct.downloadFicheSanitaire(user, id)
	if err != nil {
		return err
	}
	return fsAPI.SendBlob(c, content, name)
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

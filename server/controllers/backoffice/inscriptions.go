package backoffice

import (
	"database/sql"
	"errors"
	"slices"

	"registro/controllers/logic"
	"registro/controllers/search"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// InscriptionsGet returns the [Dossier]s to be validated.
func (ct *Controller) InscriptionsGet(c echo.Context) error {
	out, err := ct.getInscriptions()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Inscription struct {
	Dossier      ds.Dossier
	Message      string // le message (optionnel) du formulaire d'inscription
	Responsable  pr.Personne
	Participants []cps.ParticipantExt
}

func (ct *Controller) getInscriptions() ([]Inscription, error) {
	dossiers, err := ds.SelectAllDossiers(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers.RestrictByValidated(false)

	return ct.loadInscriptionsContent(dossiers.IDs()...)
}

func (ct *Controller) loadInscriptionsContent(ids ...ds.IdDossier) ([]Inscription, error) {
	loader, err := logic.LoadDossiers(ct.db, ids...)
	if err != nil {
		return nil, err
	}

	out := make([]Inscription, 0, len(ids))
	for _, id := range ids {
		dossier := loader.For(id)

		var message string
		// at this point, since the espace perso is in
		// readonly mode, there is at most one message
		if l := dossier.Events.By(evs.Message); len(l) != 0 {
			content := l[0].Content.(logic.Message).Message
			if content.Origine == evs.FromEspaceperso {
				message = content.Contenu
			}
		}

		out = append(out, Inscription{
			Dossier:      dossier.Dossier,
			Responsable:  dossier.Responsable(),
			Participants: dossier.ParticipantsExt(),
			Message:      message,
		})
	}

	// sort by time
	slices.SortFunc(out, func(a, b Inscription) int { return a.Dossier.MomentInscription.Compare(b.Dossier.MomentInscription) })
	return out, nil
}

func (ct *Controller) InscriptionsSearchSimilaires(c echo.Context) error {
	id, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	out, err := ct.searchSimilaires(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) searchSimilaires(id pr.IdPersonne) ([]search.ScoredPersonne, error) {
	const maxCount = 5
	personnes, err := pr.SelectAllPersonnes(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	input := personnes[id]

	_, filtered := search.ChercheSimilaires(utils.MapValues(personnes), search.NewPatternsSimilarite(input))
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	return filtered, nil
}

type InscriptionIdentifieIn struct {
	IdDossier ds.IdDossier
	Target    IdentTarget
}

// InscriptionsIdentifiePersonne identifie et renvoie l'inscription
// mise à jour
func (ct *Controller) InscriptionsIdentifiePersonne(c echo.Context) error {
	var args InscriptionIdentifieIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.identifieInscriptionPersonne(args)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) identifieInscriptionPersonne(args InscriptionIdentifieIn) (Inscription, error) {
	err := IdentifiePersonne(ct.db, args.Target)
	if err != nil {
		return Inscription{}, err
	}

	l, err := ct.loadInscriptionsContent(args.IdDossier)
	if err != nil {
		return Inscription{}, err
	}

	return l[0], nil
}

// IdentTarget indique comment identifier une personne temporaire.
// Si `Rattache` vaut false, la personne est simplement marquée comme non temporaire.
// Sinon, le profil [IdTemporaire] est supprimé et toutes ses occurences sont remplacées
// par [RattacheTo]. [RattacheTo] est mis à jour pour prendre en compte le contenu de [IdTemporaire],
// en utilisant [search.Merge]
type IdentTarget struct {
	IdTemporaire pr.IdPersonne // le profil à rattacher

	Rattache   bool
	RattacheTo pr.IdPersonne // only valid if [Rattache] is true
}

func IdentifiePersonne(db *sql.DB, args IdentTarget) error {
	temporaire, err := pr.SelectPersonne(db, args.IdTemporaire)
	if err != nil {
		return utils.SQLError(err)
	}

	if !args.Rattache {
		// on marque simplement la personne 'entrante' comme non temporaire
		temporaire.IsTemp = false
		_, err = temporaire.Update(db)
		if err != nil {
			return utils.SQLError(err)
		}
		return nil
	}

	if args.IdTemporaire == args.RattacheTo {
		return errors.New("internal error: same target and origin profil")
	}
	err = utils.InTx(db, func(tx *sql.Tx) error {
		existant, err := pr.SelectPersonne(tx, args.RattacheTo)
		if err != nil {
			return err
		}
		if existant.IsTemp {
			return errors.New("internal error: target is temporary")
		}

		// 1) on applique les modifications de la fusion
		existant.Etatcivil, _ = search.Merge(temporaire.Etatcivil, existant.Etatcivil)
		_, err = existant.Update(tx)
		if err != nil {
			return err
		}

		// 2) redirige les occurrences de [IdTemporaire]
		if err = cps.SwitchParticipantPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = cps.SwitchEquipierPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = ds.SwitchDossierPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = files.SwitchDemandePersonne(tx, existant.Id.Opt(), temporaire.Id.Opt()); err != nil {
			return err
		}
		if err = files.SwitchFilePersonnePersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}

		// 3) supprime la personne temporaire désormais inutile
		_, err = pr.DeletePersonneById(tx, temporaire.Id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// InscriptionsValide marque l'inscription comme validée, après s'être assuré
// qu'aucune personne impliquée n'est temporaire.
//
// Le statut des participants est aussi mis à jour (de manière automatique).
func (ct *Controller) InscriptionsValide(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	err = ct.valideInscription(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) valideInscription(id ds.IdDossier) error {
	data, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return err
	}

	// on s'assure qu'aucune personne n'est temporaire
	for _, pe := range data.Personnes() {
		if pe.IsTemp {
			return errors.New("internal error: personne should not be temporary")
		}
	}

	// le status est calculé camp par camp
	dossierByCamp := data.Participants.ByIdCamp()

	// on calcule le statut des participants (requiert les participants et personnes déjà inscrites)
	loaders, err := cps.LoadCamps(ct.db, data.Camps().IDs()...)
	if err != nil {
		return err
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		for _, loader := range loaders {
			incommingPa := utils.MapValues(dossierByCamp[loader.Camp.Id])
			incommingPe := data.PersonnesFor(incommingPa)
			for index, status := range loader.Status(incommingPe) {
				listeAttente := status.Hint()
				part := incommingPa[index]
				// update the participant
				part.Statut = listeAttente
				_, err = part.Update(tx)
				if err != nil {
					return err
				}
			}
		}
		dossier := data.Dossier
		dossier.IsValidated = true
		_, err = dossier.Update(tx)

		return err
	})

	return err
}

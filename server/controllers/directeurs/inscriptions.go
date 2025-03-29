package directeurs

import (
	"errors"
	"slices"

	"registro/controllers/backoffice"
	"registro/controllers/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) InscriptionsGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getInscriptions(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// sortParticipants affiche les participants du camp en premier
func sortParticipants(insc logic.Inscription, user cps.IdCamp) {
	slices.SortFunc(insc.Participants, func(a, b cps.ParticipantCamp) int {
		if a.Camp.Id == user {
			return -1
		} else if b.Camp.Id == user {
			return 1
		}
		return 0
	})
}

func (ct *Controller) getInscriptions(user cps.IdCamp) ([]logic.Inscription, error) {
	parts, err := cps.SelectParticipantsByIdCamps(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers, err := ds.SelectDossiers(ct.db, parts.IdDossiers()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// restrict to new inscriptions
	dossiers.RestrictByValidated(false)

	out, err := logic.LoadInscriptions(ct.db, dossiers.IDs()...)
	if err != nil {
		return nil, err
	}

	// sort participant by camp
	for _, insc := range out {
		sortParticipants(insc, user)
	}

	return out, nil
}

func (ct *Controller) InscriptionsSearchSimilaires(c echo.Context) error {
	id, err := utils.QueryParamInt[pr.IdPersonne](c, "idPersonne")
	if err != nil {
		return err
	}
	out, err := logic.SearchSimilaires(ct.db, id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type InscriptionIdentifieIn = backoffice.InscriptionIdentifieIn

// InscriptionsIdentifiePersonne identifie et renvoie l'inscription
// mise à jour
func (ct *Controller) InscriptionsIdentifiePersonne(c echo.Context) error {
	user := JWTUser(c)

	var args InscriptionIdentifieIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := logic.IdentifiePersonne(ct.db, args.Target)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, args.IdDossier)
	if err != nil {
		return err
	}
	out := l[0]
	sortParticipants(out, user)

	return c.JSON(200, out)
}

// InscriptionsHintValide renvoie la suggestion automatique
// de statut pour chaque participant du dossier donné.
//
// Voir aussi [InscriptionsValide] pour l'action effective.
func (ct *Controller) InscriptionsHintValide(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	out, err := ct.hintValideInscription(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) hintValideInscription(id ds.IdDossier) (logic.StatutHints, error) {
	loader, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}

	return loader.PrepareValideInscription(ct.db)
}

type InscriptionsValideIn = backoffice.InscriptionsValideIn

// InscriptionsValide marque l'inscription comme validée, après s'être assuré
// qu'aucune personne impliquée n'est temporaire.
//
// Le statut des participants est aussi mis à jour (de manière automatique),
// et un mail d'accusé de réception est envoyé.
func (ct *Controller) InscriptionsValide(c echo.Context) error {
	user := JWTUser(c)

	var args InscriptionsValideIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.valideInscription(args, user)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, args.IdDossier)
	if err != nil {
		return err
	}
	out := l[0]
	sortParticipants(out, user)

	return c.JSON(200, out)
}

func (ct *Controller) valideInscription(args InscriptionsValideIn, idCamp cps.IdCamp) error {
	loader, err := logic.LoadDossier(ct.db, args.IdDossier)
	if err != nil {
		return err
	}

	// on s'assure qu'aucune personne n'est temporaire
	for _, pe := range loader.Personnes() {
		if pe.IsTemp {
			return errors.New("internal error: Personne should not be temporary")
		}
	}

	// TODO:
	// tmp, err := loader.PrepareValideInscription(ct.db)
	// if err != nil {
	// 	return err
	// }

	// // only update the participant for the given camp
	// participants := tmp.ByIdCamp()[idCamp]

	// // validate the dossier if the other camp are already validated
	// otherValidated := true
	// for _, part := range loader.Participants {
	// 	if part.IdCamp != idCamp && part.Statut == cps.AStatuer {
	// 		otherValidated = false
	// 		break
	// 	}
	// }

	// err = utils.InTx(ct.db, func(tx *sql.Tx) error {
	// 	for _, part := range participants {
	// 		_, err = part.Update(tx)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	if otherValidated {
	// 		loader.Dossier.IsValidated = true
	// 		_, err = loader.Dossier.Update(tx)
	// 	}

	// 	return err
	// })

	return err
}

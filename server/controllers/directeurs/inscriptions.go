package directeurs

import (
	"database/sql"
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

// InscriptionsValide met mis à jour le statut des participants
// du camp (de manière automatique).
// Si tous les participants ont été validés (statut non zéro),
// l'inscription est marquée comme validée.
func (ct *Controller) InscriptionsValide(c echo.Context) error {
	user := JWTUser(c)
	id, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	err = ct.valideInscription(id, user)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, id)
	if err != nil {
		return err
	}
	out := l[0]
	sortParticipants(out, user)

	return c.JSON(200, out)
}

func (ct *Controller) valideInscription(id ds.IdDossier, idCamp cps.IdCamp) error {
	loader, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return err
	}

	tmp, err := loader.PrepareValideInscription(ct.db)
	if err != nil {
		return err
	}

	// only update the participant for the given camp
	participants := tmp.ByIdCamp()[idCamp]

	// validate the dossier if the other camp are already validated
	otherValidated := true
	for _, part := range loader.Participants {
		if part.IdCamp != idCamp && part.Statut == cps.AStatuer {
			otherValidated = false
			break
		}
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		for _, part := range participants {
			_, err = part.Update(tx)
			if err != nil {
				return err
			}
		}
		if otherValidated {
			loader.Dossier.IsValidated = true
			_, err = loader.Dossier.Update(tx)
		}

		return err
	})

	return err
}

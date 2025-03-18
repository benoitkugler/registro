package backoffice

import (
	"database/sql"

	"registro/controllers/logic"
	ds "registro/sql/dossiers"
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

func (ct *Controller) getInscriptions() ([]logic.Inscription, error) {
	dossiers, err := ds.SelectAllDossiers(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers.RestrictByValidated(false)

	return logic.LoadInscriptions(ct.db, dossiers.IDs()...)
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

type InscriptionIdentifieIn struct {
	IdDossier ds.IdDossier
	Target    logic.IdentTarget
}

// InscriptionsIdentifiePersonne identifie et renvoie l'inscription
// mise à jour
func (ct *Controller) InscriptionsIdentifiePersonne(c echo.Context) error {
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

	return c.JSON(200, out)
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
	loader, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return err
	}

	participants, err := loader.PrepareValideInscription(ct.db)
	if err != nil {
		return err
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		for _, part := range participants {
			_, err = part.Update(tx)
			if err != nil {
				return err
			}
		}
		loader.Dossier.IsValidated = true
		_, err = loader.Dossier.Update(tx)

		return err
	})

	return err
}

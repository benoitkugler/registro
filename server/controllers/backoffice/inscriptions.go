package backoffice

import (
	"registro/logic"
	cps "registro/sql/camps"
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

	return logic.LoadInscriptions(ct.db, backofficeRights, dossiers.IDs()...)
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

	l, err := logic.LoadInscriptions(ct.db, backofficeRights, args.IdDossier)
	if err != nil {
		return err
	}
	out := l[0]

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
	out, err := logic.HintValideInscription(ct.db, backofficeRights, id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

var backofficeRights = logic.StatutBypassRights{ProfilInvalide: true, CampComplet: true, Inscrit: true}

// InscriptionsValide marque l'inscription comme validée, après s'être assuré
// qu'aucune personne impliquée n'est temporaire.
//
// Le statut des participants est mis à jour
// et un mail d'accusé de réception est envoyé.
func (ct *Controller) InscriptionsValide(c echo.Context) error {
	var args logic.InscriptionsValideIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.valideInscription(c.Request().Host, args)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, backofficeRights, args.IdDossier)
	if err != nil {
		return err
	}
	out := l[0]

	return c.JSON(200, out)
}

func (ct *Controller) valideInscription(host string, args logic.InscriptionsValideIn) error {
	return logic.ValideInscription(ct.db, ct.key, ct.smtp, ct.asso,
		host, args, backofficeRights, cps.OptIdCamp{})
}

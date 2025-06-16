package directeurs

import (
	"slices"

	"registro/controllers/backoffice"
	"registro/logic"
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

// sortParticipants affiche les participants du séjour en premier
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

// for now, this is hard-coded
var directeursBypass = logic.StatutBypassRights{ProfilInvalide: false, CampComplet: true, Inscrit: false}

func (ct *Controller) hintValideInscription(id ds.IdDossier) (logic.StatutHints, error) {
	loader, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return nil, err
	}

	return loader.StatutHints(ct.db, directeursBypass)
}

type InscriptionsValideIn = logic.InscriptionsValideIn

func (ct *Controller) InscriptionsValide(c echo.Context) error {
	user := JWTUser(c)

	var args InscriptionsValideIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.valideInscription(c.Request().Host, args, user)
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

func (ct *Controller) valideInscription(host string, args InscriptionsValideIn, idCamp cps.IdCamp) error {
	return logic.ValideInscription(ct.db, ct.key, ct.smtp, ct.asso,
		host, args, directeursBypass, idCamp.Opt())
}

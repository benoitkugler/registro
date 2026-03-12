package backoffice

import (
	"registro/logic"
	"registro/logic/search"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// InscriptionsGet returns the [Dossier]s to be validated.
func (ct *Controller) InscriptionsGet(c echo.Context) error {
	_, isFondsSoutien := JWTUser(c)
	out, err := ct.getInscriptions(isFondsSoutien)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getInscriptions(isFondsSoutien bool) ([]logic.InscriptionExt, error) {
	participants, err := cps.SelectParticipantsByStatut(ct.db, cps.AStatuer)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	return logic.LoadInscriptions(ct.db, backofficeRights, isFondsSoutien, participants.IdDossiers()...)
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
	_, isFondsSoutien := JWTUser(c)
	var args InscriptionIdentifieIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := logic.IdentifiePersonne(ct.db, args.Target)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, backofficeRights, isFondsSoutien, args.IdDossier)
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
	_, isFondsSoutien := JWTUser(c)

	var args logic.InscriptionsValideIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.valideInscription(c.Request().Host, args)
	if err != nil {
		return err
	}

	l, err := logic.LoadInscriptions(ct.db, backofficeRights, isFondsSoutien, args.IdDossier)
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

// InscriptionsSearchDoublons parcourt la table "inscriptions" à la recherche
// d'un même participant inscrit sur plusieurs séjours.
// Les séjours concernés sont uniquement ceux ouverts aux inscriptions.
func (ct *Controller) InscriptionsSearchDoublons(c echo.Context) error {
	out, err := ct.searchInscriptionsDoublons()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type InscriptionsDoublonsOut struct {
	Participants [][]in.InscriptionParticipant // each item is non empty

	Inscriptions in.Inscriptions // enough for ids in [Participants]
	Camps        cps.Camps       // enough for ids in [Participants]
}

func (ct *Controller) searchInscriptionsDoublons() (InscriptionsDoublonsOut, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return InscriptionsDoublonsOut{}, utils.SQLError(err)
	}
	camps.RestrictOpen()

	participants, err := in.SelectInscriptionParticipantsByIdCamps(ct.db, camps.IDs()...)
	if err != nil {
		return InscriptionsDoublonsOut{}, utils.SQLError(err)
	}
	// build a crible, keyed by participant identity
	crible := make(map[search.PatternsSimilarite][]in.InscriptionParticipant)
	for _, part := range participants {
		key := search.NewPatternsSimilarite(part.Identite())
		crible[key] = append(crible[key], part)
	}

	// now restrict to doublons
	var out [][]in.InscriptionParticipant
	inscriptionsIds := utils.NewSet[in.IdInscription]()
	for _, inscList := range crible {
		if len(inscList) < 2 {
			continue
		}
		out = append(out, inscList)
		for _, insc := range inscList {
			inscriptionsIds.Add(insc.IdInscription)
		}
	}

	inscriptions, err := in.SelectInscriptions(ct.db, inscriptionsIds.Keys()...)
	if err != nil {
		return InscriptionsDoublonsOut{}, utils.SQLError(err)
	}

	return InscriptionsDoublonsOut{out, inscriptions, camps}, nil
}

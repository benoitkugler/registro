package backoffice

import (
	"slices"
	"strings"

	"registro/controllers/logic"
	"registro/controllers/search"
	"registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

type QueryAttente uint8

const (
	EmptyQA         QueryAttente = iota // Indifférent
	AvecAttente                         // Avec liste d'attente
	AvecInscrits                        // Avec inscrits
	AvecAttenteOnly                     // Seulement avec liste d'attente
)

type QueryReglement uint8

const (
	EmptyQR QueryReglement = iota // Indifférent
	Zero                          // Non commencé
	Partiel                       // En cours
	Total                         // Complété
)

// The zero value defaults to returning everything
type SearchDossierIn struct {
	Pattern   string // Responsable et participants
	IdCamp    events.OptIdCamp
	Attente   QueryAttente
	Reglement QueryReglement
}

type SearchDossierOut struct {
	Dossiers []DossierHeader // passing the query
	Total    int             // all dossiers in the DB, not just passing the query
}

// DossiersSearch returns a list of [Dossier] headers
// matching the given query, sorted by activity time (defined by the messages)
func (ct *Controller) DossiersSearch(c echo.Context) error {
	var args SearchDossierIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.searchDossiers(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type DossierHeader struct {
	Id           ds.IdDossier
	Responsable  string
	Participants string
	NewMessages  int
}

func newDossierHeader(dossier logic.DossierFinance) DossierHeader {
	personnes := dossier.Personnes()
	// extract participants
	chunks := make([]string, 0, len(personnes)-1)
	for _, pe := range personnes[1:] {
		chunks = append(chunks, pe.PrenomNOM())
	}
	return DossierHeader{
		Id:           dossier.Dossier.Dossier.Id,
		Responsable:  personnes[0].PrenomNOM(),
		Participants: strings.Join(chunks, ", "),
		NewMessages:  len(dossier.Events.NewMessagesForBackoffice()),
	}
}

func (ct *Controller) searchDossiers(query SearchDossierIn) (SearchDossierOut, error) {
	dossiers, err := ds.SelectAllDossiers(ct.db)
	if err != nil {
		return SearchDossierOut{}, utils.SQLError(err)
	}
	dossiers.RestrictByValidated(true)
	ids := dossiers.IDs()

	data, err := logic.NewDossiersFinances(ct.db, ids...)
	if err != nil {
		return SearchDossierOut{}, err
	}

	queryText := search.NewQuery(query.Pattern)

	var filtered []logic.DossierFinance
	for _, id := range ids {
		dossier := data.For(id)
		if match(dossier, queryText, query.IdCamp, query.Attente, query.Reglement) {
			filtered = append(filtered, dossier)
		}
	}

	// sort by messages time
	slices.SortFunc(filtered, func(a, b logic.DossierFinance) int { return a.Time().Compare(b.Time()) })

	// paginate and return the headers only
	const maxCount = 50
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	out := make([]DossierHeader, len(filtered))
	for i, v := range filtered {
		out[i] = newDossierHeader(v)
	}
	return SearchDossierOut{out, len(ids)}, nil
}

func match(dossier logic.DossierFinance,
	text search.Query, idCamp events.OptIdCamp, attente QueryAttente, reglement QueryReglement,
) bool {
	// critère camp
	if idCamp.Valid {
		_, hasCamp := dossier.Camps()[idCamp.Id]
		if !hasCamp {
			return false
		}
	}

	// critère texte
	matchText := false
	for _, personne := range dossier.Personnes() {
		if search.QueryMatch(text, personne) {
			matchText = true
			break
		}
	}
	if !matchText {
		return false
	}

	// critère liste d'attente
	if attente != EmptyQA {
		var (
			hasAtLeastOneAttente, hasAtLeastOneInscrit = false, false
			hasAllAttente                              = true
		)
		for _, part := range dossier.Participants {
			// ignore les participants en dehors du camp sélectionné
			if idCamp.Valid && idCamp.Id != part.IdCamp {
				continue
			}
			if part.Statut == camps.Inscrit {
				hasAtLeastOneInscrit = true
				hasAllAttente = false
			} else {
				hasAtLeastOneAttente = true
			}
		}
		switch attente {
		case AvecAttente:
			return hasAtLeastOneAttente
		case AvecInscrits:
			return hasAtLeastOneInscrit
		case AvecAttenteOnly:
			return hasAllAttente
		}
	}

	// critère financier
	if reglement != EmptyQR {
		matchStatut := dossier.Bilan().StatutPaiement() == logic.StatutPaiement(reglement)
		if !matchStatut {
			return false
		}
	}

	// we have a match !
	return true
}

func (ct *Controller) DossiersLoad(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "id")
	if err != nil {
		return err
	}
	out, err := ct.loadDossier(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type DossierExt struct{}

// also marks the message as seen
func (ct *Controller) loadDossier(id ds.IdDossier) (DossierExt, error) {
	panic("unimplemented")
}

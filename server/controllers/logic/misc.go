package logic

import (
	"slices"

	"registro/controllers/search"
	cps "registro/sql/camps"
	pr "registro/sql/personnes"
	"registro/utils"
)

type CampItem struct {
	Id    cps.IdCamp
	Label string
	IsOld bool // true if the end is passed by 45 jours
}

func NewCampItem(camp cps.Camp) CampItem {
	const deltaOld = 45
	return CampItem{camp.Id, camp.Label(), camp.IsPassedBy(deltaOld)}
}

func LoadCamps(db cps.DB) ([]CampItem, error) {
	camps, err := cps.SelectAllCamps(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	list := utils.MapValues(camps)
	slices.SortFunc(list, func(a, b cps.Camp) int { return -a.DateDebut.Time().Compare(b.DateDebut.Time()) })

	out := make([]CampItem, len(list))
	for i, camp := range list {
		out[i] = NewCampItem(camp)
	}
	return out, nil
}

func SelectPersonne(db pr.DB, pattern string, removeTemp bool) ([]search.PersonneHeader, error) {
	const maxCount = 10
	personnes, err := pr.SelectAllPersonnes(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	if removeTemp {
		personnes.RemoveTemp()
	}
	out := search.FilterPersonnes(personnes, pattern)
	if len(out) > maxCount {
		out = out[:maxCount]
	}
	return out, nil
}

func SearchSimilaires(db pr.DB, id pr.IdPersonne) ([]search.ScoredPersonne, error) {
	const maxCount = 5
	personnes, err := search.SelectAllFieldsForSimilaires(db)
	if err != nil {
		return nil, err
	}
	input := personnes[id]

	_, filtered := search.ChercheSimilaires(personnes, search.NewPatternsSimilarite(input))
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	return filtered, nil
}

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

func LoadCamps(db cps.DB) ([]CampItem, error) {
	const deltaOld = 45
	camps, err := cps.SelectAllCamps(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	list := utils.MapValues(camps)
	slices.SortFunc(list, func(a, b cps.Camp) int { return -a.DateDebut.Time().Compare(b.DateDebut.Time()) })

	out := make([]CampItem, len(list))
	for i, camp := range list {
		out[i] = CampItem{camp.Id, camp.Label(), camp.IsPassedBy(deltaOld)}
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

package logic

import (
	"slices"

	cps "registro/sql/camps"
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

package dons

import "registro/sql/shared"

type OptIdOrganisme = shared.OptID[IdOrganisme]

func (id IdOrganisme) Opt() OptIdOrganisme { return OptIdOrganisme{Id: id, Valid: true} }

package dons

import (
	"registro/sql/dossiers"
	"registro/sql/shared"
)

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type IdDon int64

type Don struct {
	Id           IdDon
	Montant      dossiers.Montant
	ModePaiement dossiers.ModePaiement
	Date         shared.Date
	Affectation  string // indicatif
	Details      string // détails additionels
	Remercie     bool   // `true` si le remerciement a été envoyé

	// champ caché, optionnel
	IdPaiementHelloasso int32
}

// TODO: organismes et donateurs

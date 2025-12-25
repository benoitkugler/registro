package dons

import (
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type (
	IdDon       int64
	IdOrganisme int64
)

//
// gomacro:SQL ADD CHECK(IdPersonne <> null OR IdOrganisme <> null)
// gomacro:SQL ADD CHECK(IdPersonne = null OR IdOrganisme = null)

type Don struct {
	Id IdDon

	// Donateur
	IdPersonne  pr.OptIdPersonne `gomacro-sql-foreign:"Personne"`
	IdOrganisme OptIdOrganisme   `gomacro-sql-foreign:"Organisme"`

	Montant      dossiers.Montant
	ModePaiement dossiers.ModePaiement
	Date         shared.Date
	Affectation  string // indicatif
	Details      string // détails additionels

	Remercie bool // `true` si le remerciement a été envoyé

	// champ caché, optionnel
	IdPaiementHelloasso int32
}

// Organisme désigne un groupe de personne (typiquement une église),
// à l'origine d'un don. Un don d'un organisme de donne pas lieu à un reçu fiscal.
type Organisme struct {
	Id  IdOrganisme
	Nom string

	Mail       string
	Adresse    string
	CodePostal string
	Ville      string
	Pays       pr.Pays
}

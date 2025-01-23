package personnes

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type IdPersonne int64

// Personne représente les attributs d'une personne
type Personne struct {
	Id IdPersonne

	Etatcivil

	// used for equipiers

	Diplome           Diplome
	Approfondissement Approfondissement

	FicheSanitaire FicheSanitaire

	// IsTemp is `true` for non verified profils,
	// which may require to be merged to an existant one
	IsTemp bool
}

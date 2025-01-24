package personnes

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type IdPersonne int64

// Personne représente les attributs d'une personne
type Personne struct {
	Id IdPersonne

	Etatcivil

	// IsTemp is `true` for non verified profils,
	// which may require to be merged to an existant one
	IsTemp bool
}

// RemoveTemp delete from [m] the temporary profiles
func (m Personnes) RemoveTemp() {
	for k, v := range m {
		if v.IsTemp {
			delete(m, k)
		}
	}
}

// Fichesanitaire stores information as declared on the personnal space.
//
// Information from the responsable legal will be required to display
// the complete document.
//
// gomacro:SQL ADD UNIQUE(IdPersonne)
type Fichesanitaire struct {
	IdPersonne IdPersonne `gomacro-sql-on-delete:"CASCADE"`

	TraitementMedical bool
	Maladies          Maladies
	Allergies         Allergies
	DifficultesSante  string
	Recommandations   string
	Handicap          bool
	Tel               Tel // added to the one of the responsable
	Medecin           Medecin

	LastModif Time  // dernière modification
	Mails     Mails // owners
}

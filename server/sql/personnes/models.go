package personnes

import "time"

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

type IdPersonne int64

// Personne représente les attributs d'une personne
//
// required by Fichesanitaire
// gomacro:SQL ADD UNIQUE(Id, IsTemp)
type Personne struct {
	Id IdPersonne

	Etatcivil

	// Publicite est utilisé pour des exports automatiques
	// TODO: https://github.com/benoitkugler/registro/issues/11
	Publicite Publicite

	CharteAccepted time.Time // last time a participant has accepted the asso charte

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

// SelectByMail returns the profiles with the given [mail].
func SelectByMail(db DB, mail string) (Personnes, error) {
	rows, err := db.Query("SELECT * FROM personnes WHERE Mail = $1", mail)
	if err != nil {
		return nil, err
	}
	return ScanPersonnes(rows)
}

// Fichesanitaire stores information as declared on the personnal space.
//
// Information from the responsable legal will be required to display
// the complete document.
//
// gomacro:SQL ADD UNIQUE(IdPersonne)
//
// Temp people must not have one Fichesanitaire
// gomacro:SQL ADD FOREIGN KEY (IdPersonne, guard) REFERENCES Personne(Id,IsTemp)
type Fichesanitaire struct {
	IdPersonne IdPersonne `gomacro-sql-on-delete:"CASCADE"`

	DifficultesSante      string
	AllergiesAlimentaires string
	TraitementMedical     string

	Medecin      NomTel
	AutreContact NomTel // added to the responsable

	// TODO: maybe add assurances

	Modified time.Time
	Owners   Mails // used for security

	guard bool `gomacro-sql-guard:"false"`
}

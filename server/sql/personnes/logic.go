package personnes

import "time"

func (r Etatcivil) Age() int { return r.DateNaissance.Age(time.Now()) }

type FichesanitaireState uint8

const (
	NoFiche FichesanitaireState = iota
	Outdated
	UpToDate
)

// State returns the state of the fiche sanitaire with respect to
// the inscription time.
func (fs Fichesanitaire) State(inscription time.Time) FichesanitaireState {
	if fs.LastModif.IsZero() { // never filled
		return NoFiche
	}
	if fs.LastModif.Before(inscription) { // filled some time ago
		return Outdated
	}
	return UpToDate
}

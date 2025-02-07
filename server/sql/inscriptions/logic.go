package inscriptions

import (
	"errors"
	"strings"
	"time"
)

type InscriptionExt struct {
	Inscription
	Participants []InscriptionParticipant
}

func (insc *InscriptionExt) Check() error {
	if strings.TrimSpace(insc.Responsable.Nom) == "" {
		return errors.New("missing Nom")
	}
	if strings.TrimSpace(insc.Responsable.Prenom) == "" {
		return errors.New("missing Prenom")
	}
	if insc.Responsable.DateNaissance.Time().IsZero() {
		return errors.New("missing DateNaissance")
	}
	if len(insc.Participants) == 0 {
		return errors.New("missing Participants")
	}
	age := insc.Responsable.DateNaissance.Age(time.Now())
	if age < 18 {
		return errors.New("invalid Age")
	}
	for _, part := range insc.Participants {
		if part.DateNaissance.Time().IsZero() {
			return errors.New("missing Participant.DateNaissance")
		}
	}
	return nil
}

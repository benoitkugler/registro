package backoffice

import (
	"slices"

	evAPI "registro/controllers/events"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

// InscriptionsGet returns the [Dossier]s to be validated.
func (ct *Controller) InscriptionsGet(c echo.Context) error {
	out, err := ct.getInscriptions()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type Inscription struct {
	Dossier      ds.Dossier
	Message      string // le message (optionnel) du formulaire d'inscription
	Responsable  pr.Personne
	Participants []cps.ParticipantExt
}

func (ct *Controller) getInscriptions() ([]Inscription, error) {
	dossiers, err := ds.SelectAllDossiers(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	dossiers.RestrictByValidated(false)

	// select the participants and associated people
	links, err := cps.SelectParticipantsByIdDossiers(ct.db, dossiers.IDs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	participants := links.ByIdDossier()

	personnes, err := pr.SelectPersonnes(ct.db, append(dossiers.IdResponsables(), links.IdPersonnes()...)...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	// load the camps
	camps, err := cps.SelectCamps(ct.db, links.IdCamps()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	// load the messages
	ld, err := evAPI.NewLoaderFor(ct.db, dossiers.IDs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	out := make([]Inscription, 0, len(dossiers))
	for _, dossier := range dossiers {
		var ps []cps.ParticipantExt
		for _, part := range participants[dossier.Id] {
			ps = append(ps, cps.ParticipantExt{
				Participant: part,
				Camp:        camps[part.IdCamp],
				Personne:    personnes[part.IdPersonne],
			})
		}

		var message string
		if l := ld.EventsFor(dossier.Id).By(evs.Message); len(l) != 0 {
			content := l[0].Content.(evAPI.Message).Message
			if content.Origine == evs.FromEspaceperso {
				message = content.Contenu
			}
		}

		out = append(out, Inscription{
			Dossier:      dossier,
			Responsable:  personnes[dossier.IdResponsable],
			Participants: ps,
			Message:      message,
		})
	}

	// sort by time
	slices.SortFunc(out, func(a, b Inscription) int { return a.Dossier.MomentInscription.Compare(b.Dossier.MomentInscription) })
	return out, nil
}

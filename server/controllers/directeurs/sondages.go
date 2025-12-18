package directeurs

import (
	"registro/logic"
	cps "registro/sql/camps"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) SondagesGet(c echo.Context) error {
	user := JWTUser(c)
	out, err := ct.getSondages(user)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type SondageMoyennes struct {
	InfosAvantSejour   float32 // entre 1 et 4
	InfosPendantSejour float32 // entre 1 et 4
	Hebergement        float32 // entre 1 et 4
	Activites          float32 // entre 1 et 4
	Theme              float32 // entre 1 et 4
	Nourriture         float32 // entre 1 et 4
	Hygiene            float32 // entre 1 et 4
	Ambiance           float32 // entre 1 et 4
	Ressenti           float32 // entre 1 et 4
}

func (sm *SondageMoyennes) add(sd cps.Sondage) {
	sm.InfosAvantSejour += float32(sd.InfosAvantSejour)
	sm.InfosPendantSejour += float32(sd.InfosPendantSejour)
	sm.Hebergement += float32(sd.Hebergement)
	sm.Activites += float32(sd.Activites)
	sm.Theme += float32(sd.Theme)
	sm.Nourriture += float32(sd.Nourriture)
	sm.Hygiene += float32(sd.Hygiene)
	sm.Ambiance += float32(sd.Ambiance)
	sm.Ressenti += float32(sd.Ressenti)
}

func (sm *SondageMoyennes) normalize(L int) {
	sm.InfosAvantSejour /= float32(L)
	sm.InfosPendantSejour /= float32(L)
	sm.Hebergement /= float32(L)
	sm.Activites /= float32(L)
	sm.Theme /= float32(L)
	sm.Nourriture /= float32(L)
	sm.Hygiene /= float32(L)
	sm.Ambiance /= float32(L)
	sm.Ressenti /= float32(L)
}

type SondageExt struct {
	ResponsableNom  string
	ResponsableMail string
	Participants    []string // du séjour
	Sondage         cps.Sondage
}

type SondagesOut struct {
	Moyennes SondageMoyennes
	Sondages []SondageExt
}

func (ct *Controller) getSondages(idCamp cps.IdCamp) (SondagesOut, error) {
	sondages, err := cps.SelectSondagesByIdCamps(ct.db, idCamp)
	if err != nil {
		return SondagesOut{}, utils.SQLError(err)
	}
	dossiers, err := logic.LoadDossiers(ct.db, sondages.IdDossiers())
	if err != nil {
		return SondagesOut{}, err
	}

	out := SondagesOut{Sondages: make([]SondageExt, 0, len(sondages))}
	for _, sondage := range sondages {
		dossier := dossiers.For(sondage.IdDossier)
		resp := dossier.Responsable()
		item := SondageExt{ResponsableNom: resp.PrenomNOM(), ResponsableMail: resp.Mail, Sondage: sondage}
		for _, part := range dossier.ParticipantsExt() {
			if part.Camp.Id == idCamp && part.Participant.Statut == cps.Inscrit {
				item.Participants = append(item.Participants, part.Personne.PrenomN())
			}
		}

		out.Moyennes.add(sondage)
		out.Sondages = append(out.Sondages, item)
	}

	if len(sondages) != 0 {
		out.Moyennes.normalize(len(sondages))
	}

	return out, nil
}

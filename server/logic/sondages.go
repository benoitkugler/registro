package logic

import (
	cps "registro/sql/camps"
	"registro/utils"
)

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
	Participants    []string // inscrits, du séjour
	Sondage         cps.Sondage
}

type Sondages struct {
	sondages map[cps.IdCamp]cps.Sondages
	dossiers Dossiers
}

func LoadSondages(db cps.DB, camps []cps.IdCamp) (Sondages, error) {
	sondages, err := cps.SelectSondagesByIdCamps(db, camps...)
	if err != nil {
		return Sondages{}, utils.SQLError(err)
	}
	dossiers, err := LoadDossiers(db, sondages.IdDossiers())
	if err != nil {
		return Sondages{}, err
	}
	return Sondages{sondages.ByIdCamp(), dossiers}, nil
}

type CampSondages struct {
	Moyennes SondageMoyennes
	Sondages []SondageExt
}

func (sd Sondages) For(idCamp cps.IdCamp) CampSondages {
	sondages := sd.sondages[idCamp]
	out := CampSondages{Sondages: make([]SondageExt, 0, len(sondages))}
	for _, sondage := range sondages {
		dossier := sd.dossiers.For(sondage.IdDossier)
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

	return out
}

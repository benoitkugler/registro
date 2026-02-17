package logic

import (
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/utils"
)

type ParticipantExt struct {
	cps.ParticipantPersonne
	Age         int  // âge au premier jour du séjour
	HasBirthday bool // anniversaire pendant le séjour ?

	MomentInscription time.Time
}

func NewParticipantExt(participant cps.Participant, personne pr.Personne, camp cps.Camp, dossier ds.Dossier) ParticipantExt {
	return ParticipantExt{
		cps.ParticipantPersonne{Participant: participant, Personne: personne},
		camp.AgeDebutCamp(personne.DateNaissance),
		camp.Plage().HasBirthday(personne.DateNaissance),
		dossier.MomentInscription,
	}
}

func LoadParticipants(db cps.DB, id cps.IdCamp) ([]ParticipantExt, ds.Dossiers, cps.CampData, error) {
	camp, err := cps.LoadCamp(db, id)
	if err != nil {
		return nil, nil, cps.CampData{}, err
	}

	dossiers, err := ds.SelectDossiers(db, camp.IdDossiers()...)
	if err != nil {
		return nil, nil, cps.CampData{}, utils.SQLError(err)
	}

	pp := camp.Participants(false)
	l := make([]ParticipantExt, len(pp))
	for i, p := range pp {
		l[i] = NewParticipantExt(p.Participant, p.Personne, camp.Camp, dossiers[p.Participant.IdDossier])
	}

	return l, dossiers, camp, nil
}

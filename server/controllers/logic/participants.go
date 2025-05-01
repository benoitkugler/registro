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

func LoadParticipants(db cps.DB, id cps.IdCamp) ([]ParticipantExt, cps.CampExt, error) {
	camp, err := cps.LoadCamp(db, id)
	if err != nil {
		return nil, cps.CampExt{}, err
	}

	dossiers, err := ds.SelectDossiers(db, camp.IdDossiers()...)
	if err != nil {
		return nil, cps.CampExt{}, utils.SQLError(err)
	}

	pp := camp.Participants()
	l := make([]ParticipantExt, len(pp))
	for i, p := range pp {
		l[i] = NewParticipantExt(p.Participant, p.Personne, camp.Camp, dossiers[p.Participant.IdDossier])
	}

	return l, camp.Camp.Ext(), nil
}

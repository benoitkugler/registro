package camps

import (
	"testing"

	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func part(p pr.IdPersonne, statut ListeAttente) Participant {
	return Participant{IdPersonne: p, Statut: statut}
}

func pers(s pr.Sexe, n pr.Nationnalite) pr.Personne {
	return pr.Personne{Etatcivil: pr.Etatcivil{Sexe: s, Nationnalite: n}}
}

func TestCampLoader_Stats(t *testing.T) {
	tests := []struct {
		Participants Participants
		Personnes    pr.Personnes
		want         StatistiquesInscrits
	}{
		{
			Participants{},
			pr.Personnes{},
			StatistiquesInscrits{},
		},
		{
			Participants{
				1: part(1, AStatuer),
				2: part(1, Inscrit),
				3: part(1, Refuse), 4: part(1, AttenteProfilInvalide),
				5: part(1, AttenteCampComplet), 6: part(1, EnAttenteReponse),
			},
			pr.Personnes{1: pers(0, 0)},
			StatistiquesInscrits{Inscriptions: 6, Valides: 1, Refus: 1, AStatuer: 1, Exceptions: 1, Attente: 2},
		},
		{
			Participants{
				1: part(1, Inscrit), 2: part(2, Inscrit), 6: part(3, Inscrit),
				3: part(1, Refuse), 4: part(2, Refuse), 5: part(3, Refuse),
			},
			pr.Personnes{1: pers(0, 0), 2: pers(pr.Woman, pr.Francaise), 3: pers(pr.Woman, pr.Suisse)},
			StatistiquesInscrits{
				Inscriptions: 6, InscriptionsFilles: 4, InscriptionsSuisses: 2,
				Valides: 3, ValidesFilles: 2, ValidesSuisses: 1,
				Refus: 3,
			},
		},
	}
	for _, tt := range tests {
		cd := CampLoader{
			Participants: tt.Participants,
			Personnes:    tt.Personnes,
		}
		tu.Assert(t, cd.Stats() == tt.want)
	}
}

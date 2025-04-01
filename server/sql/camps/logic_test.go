package camps

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

func part(p pr.IdPersonne, statut StatutParticipant) Participant {
	return Participant{IdPersonne: p, Statut: statut}
}

func pers(s pr.Sexe, n pr.Nationnalite) pr.Personne {
	return pr.Personne{Etatcivil: pr.Etatcivil{Sexe: s, Nationnalite: n}}
}

func pers2(s pr.Sexe, age int) pr.Personne {
	now := time.Now()
	dateNaissace := shared.NewDate(now.Year()-age, now.Month(), now.Day())
	return pr.Personne{Etatcivil: pr.Etatcivil{Sexe: s, DateNaissance: dateNaissace}}
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
			Camp:         Camp{},
			participants: tt.Participants,
			personnes:    tt.Personnes,
		}
		tu.Assert(t, cd.Stats() == tt.want)
	}
}

func TestCamp_isTerminated(t *testing.T) {
	now := time.Now()
	y, m, d := now.Year(), now.Month(), now.Day()
	tu.Assert(t, (&Camp{DateDebut: shared.NewDate(y, m, d), Duree: 1}).IsPassedBy(1) == false)
	tu.Assert(t, (&Camp{DateDebut: shared.NewDate(y, m, d-1), Duree: 1}).IsPassedBy(1) == true)
}

func TestCampLoader_Status(t *testing.T) {
	campNoGF := Camp{
		AgeMin:          6,
		AgeMax:          12,
		Places:          5,
		NeedEquilibreGF: false,
		DateDebut:       shared.NewDateFrom(time.Now()),
	}
	campGF := Camp{
		AgeMin:          6,
		AgeMax:          12,
		Places:          5,
		NeedEquilibreGF: true,
		DateDebut:       shared.NewDateFrom(time.Now()),
	}
	personnes := pr.Personnes{1: pers(pr.Man, 0), 2: pers(pr.Woman, 0)}
	participants := Participants{
		1: part(1, Inscrit), 2: part(2, Inscrit), 3: part(2, Inscrit), // 3 inscrits
		4: part(1, Refuse), // ignored
	}
	tests := []struct {
		camp         Camp
		participants []pr.Personne
		want         []StatutCauses
	}{
		{campNoGF, nil, []StatutCauses{}},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, 10)},
			[]StatutCauses{{true, true, true, true}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, 18)},
			[]StatutCauses{{true, false, true, true}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, 4)},
			[]StatutCauses{{false, true, true, true}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, 10), pers2(pr.Man, 10)},
			[]StatutCauses{{true, true, true, true}, {true, true, true, true}},
		},
		{ // places manquantes
			campNoGF,
			[]pr.Personne{pers2(pr.Man, 10), pers2(pr.Man, 10), pers2(pr.Man, 10)},
			[]StatutCauses{{true, true, true, false}, {true, true, true, false}, {true, true, true, false}},
		},
		{
			campGF,
			[]pr.Personne{pers2(pr.Man, 10)},
			[]StatutCauses{{true, true, true, true}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Woman, 10), pers2(pr.Woman, 10), pers2(pr.Woman, 10)},
			[]StatutCauses{{true, true, false, false}, {true, true, false, false}, {true, true, false, false}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Man, 10), pers2(pr.Woman, 10), pers2(pr.Woman, 10)},
			[]StatutCauses{{true, true, false, false}, {true, true, false, false}, {true, true, false, false}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Man, 10), pers2(pr.Man, 10), pers2(pr.Woman, 10)},
			[]StatutCauses{{true, true, true, false}, {true, true, true, false}, {true, true, true, false}},
		},
	}
	for _, tt := range tests {
		cd := CampLoader{
			Camp:         tt.camp,
			participants: participants,
			personnes:    personnes,
		}
		fmt.Println(tt.want)
		tu.Assert(t, reflect.DeepEqual(cd.Status(tt.participants), tt.want))
	}
}

func TestAide_Resolve(t *testing.T) {
	type fields struct {
		Valeur     int
		ParJour    bool
		NbJoursMax int
	}
	tests := []struct {
		fields  fields
		nbJours int
		want    int
	}{
		{fields{ParJour: true}, 0, 0},
		{fields{Valeur: 100}, 20, 100},
		{fields{ParJour: true, Valeur: 10}, 5, 50},
		{fields{ParJour: true, Valeur: 10, NbJoursMax: 8}, 5, 50},
		{fields{ParJour: true, Valeur: 10, NbJoursMax: 4}, 5, 40},
	}
	for _, tt := range tests {
		ai := Aide{
			Valeur:     dossiers.Montant{Cent: tt.fields.Valeur},
			ParJour:    tt.fields.ParJour,
			NbJoursMax: tt.fields.NbJoursMax,
		}
		tu.Assert(t, ai.Resolve(tt.nbJours) == dossiers.Montant{Cent: tt.want})
	}
}

package camps

import (
	"reflect"
	"testing"
	"time"

	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"
)

func part(p pr.IdPersonne, statut StatutParticipant) Participant {
	return Participant{IdPersonne: p, Statut: statut}
}

func pers(s pr.Sexe, n pr.Nationnalite) pr.Personne {
	return pr.Personne{Identite: pr.Identite{Sexe: s, Nationnalite: n}}
}

func pers2(s pr.Sexe, now sh.Date, age int) pr.Personne {
	n := now.Time()
	dateNaissace := sh.NewDate(n.Year()-age, n.Month(), n.Day())
	return pr.Personne{Identite: pr.Identite{Sexe: s, DateNaissance: dateNaissace}}
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
			pr.Personnes{1: pers(0, pr.Nationnalite{})},
			StatistiquesInscrits{Inscriptions: 6, Valides: 1, Refus: 1, AStatuer: 1, Exceptions: 1, Attente: 2},
		},
		{
			Participants{
				1: part(1, Inscrit), 2: part(2, Inscrit), 6: part(3, Inscrit),
				3: part(1, Refuse), 4: part(2, Refuse), 5: part(3, Refuse),
			},
			pr.Personnes{1: pers(0, pr.Nationnalite{}), 2: pers(pr.Woman, pr.Nationnalite{}), 3: pers(pr.Woman, pr.Nationnalite{IsSuisse: true})},
			StatistiquesInscrits{
				Inscriptions: 6, InscriptionsFilles: 4, InscriptionsSuisses: 2,
				Valides: 3, ValidesFilles: 2, ValidesSuisses: 1,
				Refus: 3,
			},
		},
	}
	for _, tt := range tests {
		cd := CampData{
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
	tu.Assert(t, (&Camp{DateDebut: sh.NewDate(y, m, d), Duree: 1}).IsPassedBy(1) == false)
	tu.Assert(t, (&Camp{DateDebut: sh.NewDate(y, m, d-1), Duree: 1}).IsPassedBy(1) == true)
}

func TestCampLoader_Status(t *testing.T) {
	now := sh.NewDate(2026, time.March, 7)
	campNoGF := Camp{
		AgeMin:          6,
		AgeMax:          12,
		Places:          5,
		NeedEquilibreGF: false,
		DateDebut:       now,
		Duree:           1,
	}
	campGF := Camp{
		AgeMin:          6,
		AgeMax:          12,
		Places:          5,
		NeedEquilibreGF: true,
		DateDebut:       now,
		Duree:           1,
	}
	personnes := pr.Personnes{1: pers(pr.Man, pr.Nationnalite{}), 2: pers(pr.Woman, pr.Nationnalite{})}
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
			[]pr.Personne{pers2(pr.Man, now, 10)},
			[]StatutCauses{{true, true, true, CauseAge{}}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, now, 14)},
			[]StatutCauses{{false, true, true, CauseAge{Jeune: false, Age: 14, EcartInDays: 366}}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, now, 4)},
			[]StatutCauses{{false, true, true, CauseAge{Jeune: true, Age: 4, EcartInDays: 730}}},
		},
		{
			campNoGF,
			[]pr.Personne{pers2(pr.Man, now, 10), pers2(pr.Man, now, 10)},
			[]StatutCauses{{true, true, true, CauseAge{}}, {true, true, true, CauseAge{}}},
		},
		{ // places manquantes
			campNoGF,
			[]pr.Personne{pers2(pr.Man, now, 10), pers2(pr.Man, now, 10), pers2(pr.Man, now, 10)},
			[]StatutCauses{{true, true, false, CauseAge{}}, {true, true, false, CauseAge{}}, {true, true, false, CauseAge{}}},
		},
		{
			campGF,
			[]pr.Personne{pers2(pr.Man, now, 10)},
			[]StatutCauses{{true, true, true, CauseAge{}}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Woman, now, 10), pers2(pr.Woman, now, 10), pers2(pr.Woman, now, 10)},
			[]StatutCauses{{true, false, false, CauseAge{}}, {true, false, false, CauseAge{}}, {true, false, false, CauseAge{}}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Man, now, 10), pers2(pr.Woman, now, 10), pers2(pr.Woman, now, 10)},
			[]StatutCauses{{true, false, false, CauseAge{}}, {true, false, false, CauseAge{}}, {true, false, false, CauseAge{}}},
		},
		{ // equlibre actuel : 1G / 2F
			campGF,
			[]pr.Personne{pers2(pr.Man, now, 10), pers2(pr.Man, now, 10), pers2(pr.Woman, now, 10)},
			[]StatutCauses{{true, true, false, CauseAge{}}, {true, true, false, CauseAge{}}, {true, true, false, CauseAge{}}},
		},
	}
	for _, tt := range tests {
		cd := CampData{
			Camp:         tt.camp,
			participants: participants,
			personnes:    personnes,
		}
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

func TestSlug(t *testing.T) {
	for _, test := range []struct {
		nom      string
		annee    int
		expected string
	}{
		{"C2", 2023, "c2-2023"},
		{"MinI's", 2000, "minis-2000"},
		{"La Bessone", 2000, "la-bessone-2000"},
		{"Canoë", 2000, "canoe-2000"},
		{"école", 2000, "ecole-2000"},
	} {
		camp := Camp{Nom: test.nom, DateDebut: sh.NewDate(test.annee, 1, 1)}
		tu.Assert(t, camp.Slug() == test.expected)
	}
}

func TestGroupes_TrouveGroupe(t *testing.T) {
	tests := []struct {
		groupes       Groupes
		dateNaissance sh.Date
		want          IdGroupe
		want2         bool
	}{
		{
			Groupes{},
			sh.NewDate(2000, 1, 1),
			0, false,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2000, 1, 1)}},
			sh.NewDate(2001, 1, 1),
			0, false,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2001, 1, 1)}},
			sh.NewDate(2001, 1, 1),
			1, true,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2001, 1, 1)}},
			sh.NewDate(1999, 1, 1),
			1, true,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2000, 1, 1)}, 2: {Id: 2, Fin: sh.NewDate(2002, 1, 1)}},
			sh.NewDate(2000, 1, 1),
			1, true,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2000, 1, 1)}, 2: {Id: 2, Fin: sh.NewDate(2002, 1, 1)}},
			sh.NewDate(2001, 1, 1),
			2, true,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2000, 1, 1)}, 2: {Id: 2, Fin: sh.NewDate(2002, 1, 1)}},
			sh.NewDate(2002, 1, 1),
			2, true,
		},
		{
			Groupes{1: {Id: 1, Fin: sh.NewDate(2000, 1, 1)}, 2: {Id: 2, Fin: sh.NewDate(2002, 1, 1)}},
			sh.NewDate(2002, 1, 2),
			0, false,
		},
	}
	for _, tt := range tests {
		got, got2 := tt.groupes.TrouveGroupe(tt.dateNaissance)
		tu.Assert(t, got.Id == tt.want && got2 == tt.want2)
	}
}

func TestCamp_IsAgeValide(t *testing.T) {
	debut := sh.NewDate(2025, time.March, 5)
	tests := []struct {
		plage         sh.Plage
		dateNaissance sh.Date
		wantValid     bool
		wantCause     CauseAge
	}{
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2012, time.March, 1), false, CauseAge{false, 13, 5}}, // 13 ans
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2015, time.March, 1), true, CauseAge{}},              // 10 ans

		// cas "fins"
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2019, time.March, 14), true, CauseAge{}},             // 6 ans le dernier jour
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2012, time.March, 6), true, CauseAge{}},              // 13 ans le deuxième jour
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2019, time.March, 15), false, CauseAge{true, 5, 1}},  // 6 ans juste après le dernier jour
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2012, time.March, 5), false, CauseAge{false, 13, 1}}, // 13 ans le premier jour
		{sh.Plage{From: debut, Duree: 10}, sh.NewDate(2012, time.March, 3), false, CauseAge{false, 13, 3}},
	}
	for _, tt := range tests {
		cp := Camp{DateDebut: tt.plage.From, Duree: tt.plage.Duree, AgeMin: 6, AgeMax: 12}
		gotValid, gotCause := cp.IsAgeValide(tt.dateNaissance)
		tu.Assert(t, gotValid == tt.wantValid)
		tu.Assert(t, gotCause == tt.wantCause)
	}
}

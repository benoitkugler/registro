package sheets

import (
	"testing"
	"time"

	"registro/logic"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestStyle(t *testing.T) {
	m := map[Style]int{}
	m[Style{}] = 5 // check for crashes

	b := NewBuilder()
	_, err := b.register(Style{})
	tu.AssertNoErr(t, err)
}

func TestCreateTable(t *testing.T) {
	liste := [][]Cell{
		{{Value: "lmkeke", Bold: true}, {Value: "Blue", Color: "#AA00BB"}},
		{{ValueF: 5.56, NumFormat: FrancsSuisse}, {ValueF: 3, NumFormat: Int}},
		{{ValueF: 5.56, NumFormat: Euros}, {ValueF: 3, NumFormat: Euros}},
		{{Value: "5.56"}, {ValueF: 0.255, NumFormat: Percentage}},
		{{NumFormat: Float}, {}},
	}
	content, err := CreateTable([]string{"nom", "Prénom"}, liste)
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_Table.xlsx", content)

	content, err = CreateTableTotal([]string{"nom", "Prénom"}, liste, "123.5€")
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_TableTotal.xlsx", content)
}

func TestCreateComplex(t *testing.T) {
	// TODO: consolidate with real data

	content, err := SuiviFinancierCamp([][]Cell{
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
	}, "1354.5€", "1546.4€")
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_SuiviFinancierCamp.xlsx", content)
}

func TestFormatTime(t *testing.T) {
	for _, test := range []struct {
		t        time.Time
		expected string
	}{
		{time.Time{}, ""},
		{time.Date(2000, time.January, 3, 1, 1, 12, 0, time.Local), "03/01/2000 01:01:12"},
	} {
		tu.Assert(t, formatTime(test.t) == test.expected)
	}
}

func TestListeParticipants(t *testing.T) {
	camp := cps.Camp{DateDebut: shared.NewDateFrom(time.Now())}
	p1 := pr.Personne{Etatcivil: pr.Etatcivil{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Man,
		DateNaissance: shared.NewDate(2000, time.August, 5),
		Nationnalite:  pr.Nationnalite{IsSuisse: true},
	}}
	p2 := p1
	p2.Nationnalite.IsSuisse = false
	inscrit1 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 1, IdDossier: 1, Commentaire: utils.RandString(10, true), Navette: cps.AllerRetour},
		Personne:    p1,
	}
	inscrit2 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 2, IdDossier: 1, Commentaire: utils.RandString(10, true)},
		Personne:    p2,
	}
	g1, g2 := cps.Groupe{Nom: "Groupe 1", Couleur: "#AA12BB"}, cps.Groupe{Nom: "Groupe 2"}
	liste := []cps.ParticipantPersonne{inscrit1, inscrit2, inscrit1}
	dossiers := logic.Dossiers{Dossiers: dossiers.Dossiers{
		1: dossiers.Dossier{MomentInscription: time.Now(), IdResponsable: 2},
	}}

	content, err := ListeParticipantsCamp(camp, liste, dossiers, map[cps.IdParticipant]cps.Groupe{1: g1, 2: g2}, false)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamp_1.xlsx", content)

	content, err = ListeParticipantsCamp(camp, liste, dossiers, map[cps.IdParticipant]cps.Groupe{1: g1, 2: g2}, true)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamp_2.xlsx", content)
}

func TestListeParticipantsCamps(t *testing.T) {
	camp := cps.Camp{Id: 123, Nom: "Mini's", DateDebut: shared.NewDateFrom(time.Now())}
	p1 := pr.Personne{Etatcivil: pr.Etatcivil{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Man,
		DateNaissance: shared.NewDate(2000, time.August, 5),
		Nationnalite:  pr.Nationnalite{IsSuisse: true},
	}}
	p2 := p1
	p2.Nationnalite.IsSuisse = false
	pa1 := cps.Participant{Id: 1, IdDossier: 1, Commentaire: utils.RandString(10, true), Navette: cps.AllerRetour}
	pa2 := cps.Participant{Id: 2, IdDossier: 2, Commentaire: utils.RandString(10, true), Remises: cps.Remises{Famille: 30}}
	inscrit1 := cps.ParticipantCamp{
		Camp: camp,
		ParticipantPersonne: cps.ParticipantPersonne{
			Participant: pa1,
			Personne:    p1,
		},
	}
	inscrit2 := cps.ParticipantCamp{
		Camp: camp,
		ParticipantPersonne: cps.ParticipantPersonne{
			Participant: pa2,
			Personne:    p2,
		},
	}
	liste := []cps.ParticipantCamp{inscrit1, inscrit2, inscrit1}
	dossiers := logic.DossiersFinances{Dossiers: logic.Dossiers{Dossiers: dossiers.Dossiers{
		1: dossiers.Dossier{Id: 1, MomentInscription: time.Now(), IdResponsable: 2},
		2: dossiers.Dossier{Id: 2, MomentInscription: time.Now().Add(time.Hour), IdResponsable: 1},
	}}}
	remisesHints := map[cps.IdParticipant]cps.Remises{
		pa1.Id: {Equipiers: 10, Famille: 5},
	}

	content, err := ListeParticipantsCamps(liste, dossiers, remisesHints, false)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamps_1.xlsx", content)

	content, err = ListeParticipantsCamps(liste, dossiers, remisesHints, true)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipantsCamps_2.xlsx", content)
}

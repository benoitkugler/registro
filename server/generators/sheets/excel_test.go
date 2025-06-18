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

func TestListeParticipants(t *testing.T) {
	camp := cps.Camp{DateDebut: shared.NewDateFrom(time.Now())}
	p1 := pr.Personne{Etatcivil: pr.Etatcivil{
		Nom: utils.RandString(12, false), Prenom: utils.RandString(12, false),
		Sexe:          pr.Man,
		DateNaissance: shared.NewDate(2000, time.August, 5),
	}}
	inscrit1 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 1, IdDossier: 1, Commentaire: utils.RandString(10, true), Navette: cps.AllerRetour},
		Personne:    p1,
	}
	inscrit2 := cps.ParticipantPersonne{
		Participant: cps.Participant{Id: 2, IdDossier: 1, Commentaire: utils.RandString(10, true)},
		Personne:    p1,
	}
	g1, g2 := cps.Groupe{Nom: "Groupe 1", Couleur: "#AA12BB"}, cps.Groupe{Nom: "Groupe 2"}
	content, err := ListeParticipants(camp, []cps.ParticipantPersonne{
		inscrit1, inscrit2, inscrit1,
	}, logic.Dossiers{Dossiers: dossiers.Dossiers{
		1: dossiers.Dossier{MomentInscription: time.Now(), IdResponsable: 2},
	}}, map[cps.IdParticipant]cps.Groupe{1: g1, 2: g2})
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_ListeParticipants.xlsx", content)
}

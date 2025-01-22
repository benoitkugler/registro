package sheets

import (
	"testing"

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

	content, err := CreateSuiviFinancierCamp([][]Cell{
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}},
	}, "1354.5€", "1546.4€")
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_SuiviFinancierCamp.xlsx", content)

	liste := [][]Cell{
		{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
		{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
	}
	content, err = CreateListeParticipants(liste, liste)
	tu.AssertNoErr(t, err)
	tu.Write(t, "registro_ListeParticipants.xlsx", content)
}

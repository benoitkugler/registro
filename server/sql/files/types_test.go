package files

import (
	"reflect"
	"testing"

	cp "registro/sql/camps"
	tu "registro/utils/testutils"
)

func TestBuiltins(t *testing.T) {
	_, err := Demandes{}.builtins()
	tu.AssertErr(t, err)
	_, err = Demandes{1: {Categorie: NoBuiltin}, 2: {Categorie: Vaccins}}.builtins()
	tu.AssertErr(t, err)
	_, err = Demandes{
		1:  {Id: 1, Categorie: 1},
		2:  {Id: 2, Categorie: 2},
		3:  {Id: 3, Categorie: 3},
		4:  {Id: 4, Categorie: 4},
		5:  {Id: 5, Categorie: 5},
		6:  {Id: 6, Categorie: 6},
		7:  {Id: 7, Categorie: 7},
		8:  {Id: 8, Categorie: 8},
		9:  {Id: 9, Categorie: 9},
		10: {Id: 10, Categorie: 10},
		11: {Id: 11, Categorie: 11},
		12: {Id: 12, Categorie: 12},
		13: {Id: 13, Categorie: 13},
		14: {Id: 14, Categorie: 14},
	}.builtins()
	tu.AssertNoErr(t, err)
}

func TestBuiltins_Defaut(t *testing.T) {
	b := Builtins{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	tests := []struct {
		equipier cp.Equipier
		want     DemandeEquipiers
	}{
		{cp.Equipier{Id: 1, Roles: cp.Roles{}}, nil},
		{cp.Equipier{Id: 1, Roles: cp.Roles{cp.Direction}}, DemandeEquipiers{
			{1, b[CarteId], true}, {1, b[Permis], true}, {1, b[SB], true}, {1, b[Bafa], true}, {1, b[Bafd], true}, {1, b[CarteVitale], true}, {1, b[Vaccins], true}, {1, b[BafdEquiv], true},
		}},
		{cp.Equipier{Id: 1, Roles: cp.Roles{cp.Direction, cp.Infirmerie}}, DemandeEquipiers{
			{1, b[CarteId], true}, {1, b[Permis], true}, {1, b[SB], true}, {1, b[Secourisme], false}, {1, b[Bafa], true}, {1, b[Bafd], true}, {1, b[CarteVitale], true}, {1, b[Vaccins], true}, {1, b[BafdEquiv], true},
		}},
	}
	for _, tt := range tests {
		got := b.Defaut(tt.equipier)
		tu.Assert(t, reflect.DeepEqual(got, tt.want))
	}
}

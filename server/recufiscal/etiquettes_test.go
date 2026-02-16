package recufiscal

import (
	"os"
	"testing"

	"registro/generators/pdfcreator"
	pr "registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestEtiquettes(t *testing.T) {
	err := pdfcreator.Init(os.TempDir(), "../assets")
	tu.AssertNoErr(t, err)

	b, err := genereEtiquettes([]pr.Identite{
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Die\nufuek\n sdsds", CodePostal: "20160", Ville: "la begude mazeklsd slkdj skljs dsklj"},
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude ded mazenc et a "},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: utils.RandString(10, true), Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
		{Nom: "KUGLER", Prenom: "Benoit", Adresse: "Le béal route de Dieufuek", CodePostal: "20160", Ville: "la begude maze"},
	})
	tu.AssertNoErr(t, err)
	tu.Write(t, "test_etiquettes.pdf", b)
}

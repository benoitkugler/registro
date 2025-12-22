package helloasso

import (
	"fmt"
	"testing"

	"registro/config"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Helloasso {
	tu.LoadEnv(t, "../env.sh")
	creds, err := config.NewHelloasso()
	tu.AssertNoErr(t, err)
	return creds
}

func TestPing(t *testing.T) {
	err := PingHelloAsso(devCreds(t))
	tu.AssertNoErr(t, err)
}

func TestFetchDons(t *testing.T) {
	db := tu.NewTestDB(t, "../sql/personnes/gen_create.sql", "../sql/dossiers/gen_create.sql", "../sql/dons/gen_create.sql")
	defer db.Remove()

	l1, err := ImportDonsHelloasso(devCreds(t), db, true)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l1) != 0)

	l2, err := ImportDonsHelloasso(devCreds(t), db, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l2) != 0)

	tu.Assert(t, len(l1) < len(l2))
}

type formV5 struct {
	FormSlug string `json:"formSlug"`
	FormType string `json:"formType"`
	Url      string `json:"url"`
	Title    string `json:"title"`
}

type formsV5 struct {
	Data       []formV5   `json:"data"`
	Pagination pagination `json:"pagination"`
}

func fetchForms(accesToken string) (formsV5, error) {
	url := "https://api.helloasso.com/v5/organizations/acve/forms"

	var out formsV5
	err := getJSON(url, accesToken, &out)

	return out, err
}

func TestShowForms(t *testing.T) {
	accesToken, err := getAccessToken(devCreds(t))
	tu.AssertNoErr(t, err)

	forms, err := fetchForms(accesToken)
	tu.AssertNoErr(t, err)

	for _, form := range forms.Data {
		fmt.Println(form.Title, form.FormSlug, form.FormType, form.Url)
	}
}

func TestAPIV5(t *testing.T) {
	accesToken, err := getAccessToken(devCreds(t))
	tu.AssertNoErr(t, err)

	forms, err := fetchForms(accesToken)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(forms.Data) > 0)

	var allPaiements []paiementHelloAsso
	for _, form := range formsHelloAsso {
		l, err := fetchAllFormPaiements(accesToken, form.formType, form.formSlug, false)
		tu.AssertNoErr(t, err)
		allPaiements = append(allPaiements, l...)
	}

	t.Logf("%d paiements", len(allPaiements))

	for _, paiement := range allPaiements {
		for _, item := range paiement.Items {
			tu.Assert(t, item.Type == "Donation" || item.Type == "MonthlyDonation")
		}

		_, err := parseDateHelloAsso(paiement.Date)
		tu.AssertNoErr(t, err)

		tu.Assert(t, paiement.Payer.FirstName != "")
	}
}

func TestIDV5ToV3(t *testing.T) {
	tu.Assert(t, idV5ToV3(23239900) == "000232399003")
	tu.Assert(t, idV5ToV3(7382535) == "000073825353")

	v, err := IdV3ToV5("000232399003")
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == 23239900)
	v, err = IdV3ToV5("000073825353")
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == 7382535)
	_, err = IdV3ToV5("000073825352")
	tu.AssertErr(t, err)
	_, err = IdV3ToV5("0000738.5353")
	tu.AssertErr(t, err)
}

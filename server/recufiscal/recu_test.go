package recufiscal

import (
	"fmt"
	"os"
	"testing"
	"time"

	"registro/generators/pdfcreator"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	sh "registro/sql/shared"
	tu "registro/utils/testutils"

	"github.com/benoitkugler/pdf/formfill"
	"github.com/benoitkugler/pdf/model"
)

var taux = ds.Taux{Euros: 1000}

func init() {
	if err := Init("templateRecuAcve.pdf"); err != nil {
		panic(err)
	}
}

func TestFieldLength(t *testing.T) {
	field := recuTemplate.Catalog.AcroForm.Flatten()["z1"].Field
	tu.Assert(t, field.FT.(model.FormFieldText).MaxLen == nil)
}

func TestRecu(t *testing.T) {
	b, err := fillPdf(recuFiscal{
		Montant: taux.Convertible(ds.NewEuros(45.4)),
		Mode:    ds.Virement,
		Date:    time.Now().AddDate(-2, 3, -5),
	},
		pr.Personne{Id: 1, Identite: pr.Identite{Nom: "GER", Prenom: "EMLZKEs"}},
	)
	tu.AssertNoErr(t, err)

	tu.Write(t, "testRecu.pdf", b)
}

func TestFields(t *testing.T) {
	champNum := champPdf{id: "z1", valeur: formfill.FDFText(numero(4))}
	donateur := pr.Personne{
		Identite: pr.Identite{
			Nom:        "')='à=(kmlrk'",
			Prenom:     "mldmskld8+-*",
			Adresse:    "lmemzkd\ndlss\nzlkdsmkmdkmsdk",
			CodePostal: "kdskdl",
			Ville:      "ùmdslsùmd",
		},
	}
	don := recuFiscal{
		Montant: taux.Convertible(ds.NewEuros(1457.457)),
		Mode:    ds.Cheque,
		Date:    time.Now(),
	}
	fields := []champPdf{champNum}

	fields = append(fields, champsACVE...)
	fields = append(fields, champsDonateur(donateur)...)
	fields = append(fields, champsDon(don)...)
	fields = append(fields, champsTypeDon...)
	fields = append(fields, champsDateEdition()...)

	for _, field := range fields {
		fmt.Println(field.id, field.valeur)
	}
}

func TestGenerate(t *testing.T) {
	err := pdfcreator.Init(os.TempDir(), "../assets")
	tu.AssertNoErr(t, err)

	db := tu.NewTestDB(t, "../migrations/create_1_tables.sql",
		"../migrations/create_2_json_funcs.sql", "../migrations/create_3_constraints.sql",
		"../migrations/init.sql")
	defer db.Remove()

	donateur1, err := pr.Personne{Identite: pr.Identite{Nom: "Kugler", Prenom: "Benoit", Mail: "smldsmkd.flm@free.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)
	donateur2, err := pr.Personne{Identite: pr.Identite{Nom: "Kugler", Prenom: "Eudes-Jàéa", Mail: "smldsmkd.flm@free.fr"}}.Insert(db)
	tu.AssertNoErr(t, err)
	organisme, err := dons.Organisme{}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = dons.Don{IdPersonne: donateur1.Id.Opt(), Montant: ds.NewEuros(50.5), ModePaiement: ds.Virement, Date: sh.NewDateFrom(time.Now())}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = dons.Don{IdPersonne: donateur1.Id.Opt(), Montant: ds.NewEuros(50.5), ModePaiement: ds.Virement, Date: sh.NewDateFrom(time.Now().AddDate(-2, 1, 1))}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = dons.Don{IdPersonne: donateur2.Id.Opt(), Montant: ds.NewEuros(100), ModePaiement: ds.Helloasso, Date: sh.NewDateFrom(time.Now())}.Insert(db)
	tu.AssertNoErr(t, err)
	_, err = dons.Don{IdOrganisme: organisme.Id.Opt(), Montant: ds.NewEuros(100), Date: sh.NewDateFrom(time.Now())}.Insert(db)
	tu.AssertNoErr(t, err)

	archive, err := Generate(db, time.Now().Year())
	tu.AssertNoErr(t, err)
	tu.Write(t, "recus.zip", archive)
}

func BenchmarkExportRecus(b *testing.B) {
	db, err := tu.SampleDBACVE.ConnectPostgres()
	tu.AssertNoErr(b, err)
	defer db.Close()

	err = Init("templateRecuAcve.pdf")
	tu.AssertNoErr(b, err)
	pdfcreator.Init(os.TempDir(), "")
	tu.AssertNoErr(b, err)

	recus, personnes, err := loadAndSelect(db, 2025)
	tu.AssertNoErr(b, err)

	for b.Loop() {
		_, err = generate(recus, personnes)
	}
}

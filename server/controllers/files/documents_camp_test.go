package files

import (
	"fmt"
	"os"
	"testing"
	"time"

	"registro/config"
	"registro/generators/pdfcreator"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	"registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestVetements(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp1, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	respo, err := personnes.Personne{}.Insert(db)
	tu.AssertNoErr(t, err)
	dossier, err := dossiers.Dossier{IdTaux: 1, IdResponsable: respo.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	for range [50]int{} {
		pe, err := personnes.Personne{Identite: personnes.Identite{
			Nom:    utils.RandString(10, true),
			Prenom: utils.RandString(10, true),
		}}.Insert(db)
		tu.AssertNoErr(t, err)
		_, err = cps.Participant{
			IdPersonne: pe.Id, IdCamp: camp1.Id, IdDossier: dossier.Id, IdTaux: 1,
			Statut: cps.Inscrit,
		}.Insert(db)
		tu.AssertNoErr(t, err)
	}

	err = pdfcreator.Init(os.TempDir(), "../../assets")
	tu.AssertNoErr(t, err)

	liste := make([]cps.Vetement, 30)
	liste[0] = cps.Vetement{Quantite: 2, Description: "e", Important: true}
	liste[1] = cps.Vetement{Quantite: 2, Description: "e", Important: false}
	camp1.Vetements = cps.ListeVetements{
		Vetements:  liste,
		Complement: "<b>smdsd</b> <a></a>",
	}
	_, err = camp1.Update(db)
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, _, err = renderListeVetements(db.DB, config.Asso{}, camp1.Id)
	tu.AssertNoErr(t, err)
	fmt.Println("rendered in", time.Since(ti))

	ti = time.Now()
	_, _, err = renderListeParticipants(db.DB, config.Asso{}, camp1.Id)
	tu.AssertNoErr(t, err)
	fmt.Println("rendered in", time.Since(ti))
}

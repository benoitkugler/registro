package inscriptions

import (
	"testing"

	"registro/sql/camps"
	"registro/sql/dossiers"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	taux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	insc, err := Inscription{IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	err = InscriptionParticipant{IdInscription: insc.Id, IdTaux: taux.Id, IdCamp: camp.Id}.Insert(db)
	tu.AssertNoErr(t, err)
}

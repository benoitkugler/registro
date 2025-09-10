package inscriptions

import (
	"database/sql"
	"testing"

	"registro/sql/camps"
	"registro/sql/dossiers"
	"registro/sql/personnes"
	"registro/utils"
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

	part := InscriptionParticipant{
		IdInscription: insc.Id, IdTaux: taux.Id, IdCamp: camp.Id,
		Nationnalite: personnes.Nationnalite{IsSuisse: true},
	}
	err = part.Insert(db)
	tu.AssertNoErr(t, err)

	err = utils.InTx(db.DB, func(tx *sql.Tx) error {
		return InsertManyInscriptionParticipants(tx, part, part)
	})
	tu.AssertNoErr(t, err)
}

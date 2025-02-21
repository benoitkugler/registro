package inscriptions

import (
	"testing"

	"registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestSQL(t *testing.T) {
	db := tu.NewTestDB(t, "../personnes/gen_create.sql", "../dossiers/gen_create.sql", "../camps/gen_create.sql", "gen_create.sql")
	defer db.Remove()

	taux, err := dossiers.Taux{Euros: 1000}.Insert(db)
	tu.AssertNoErr(t, err)
	camp, err := camps.Camp{IdTaux: taux.Id}.Insert(db)
	tu.AssertNoErr(t, err)

	peTemp, err := pr.Personne{IsTemp: true}.Insert(db)
	tu.AssertNoErr(t, err)
	pe, err := pr.Personne{IsTemp: false}.Insert(db)
	tu.AssertNoErr(t, err)

	_, err = Inscription{IdTaux: taux.Id, ResponsablePreIdent: peTemp.Id.Opt()}.Insert(db)
	tu.AssertErr(t, err)
	insc, err := Inscription{IdTaux: taux.Id, ResponsablePreIdent: pe.Id.Opt()}.Insert(db)
	tu.AssertNoErr(t, err)

	err = InscriptionParticipant{IdInscription: insc.Id, IdTaux: taux.Id, IdCamp: camp.Id, PreIdent: peTemp.Id.Opt()}.Insert(db)
	tu.AssertErr(t, err)
	err = InscriptionParticipant{IdInscription: insc.Id, IdTaux: taux.Id, IdCamp: camp.Id, PreIdent: pe.Id.Opt()}.Insert(db)
	tu.AssertNoErr(t, err)
}

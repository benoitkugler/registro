package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"registro/config"
	api "registro/controllers/inscriptions"
	"registro/crypto"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	"registro/utils/testutils"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// This is a command helper for devs,
// to generate on demand new inscriptions
func main() {
	partsCount := flag.Int("part", 2, "number of participants to create")
	flag.Parse()

	vars, err := testutils.ReadEnvFile("env.sh")
	check(err)

	for key, val := range vars {
		err = os.Setenv(key, val)
		check(err)
	}
	defer func() {
		for key := range vars {
			os.Unsetenv(key)
		}
	}()

	cfg, err := config.NewAsso()
	check(err)
	creds, err := config.NewSMTP(false)
	check(err)

	sqlCreds, err := config.NewDB()
	check(err)
	db, err := sqlCreds.ConnectPostgres()
	check(err)

	fmt.Println("Env loaded, using DB:", sqlCreds.Name)

	ct := api.NewController(db, crypto.Encrypter{}, creds, cfg)

	// assume we already have two open camps
	camps, _, _, _, err := ct.LoadCamps()
	check(err)
	campIds := camps.IDs()
	if len(campIds) < 2 {
		panic("expected at least 2 camps open for inscription")
	}
	c1, c2 := campIds[0], campIds[1]
	parts := []api.Participant{
		{IdCamp: c1, DateNaissance: shared.NewDate(2015, 1, 1), Nom: "Muler", Prenom: "Pierre"},
		{IdCamp: c1, DateNaissance: shared.NewDate(2015, 1, 1), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
		{IdCamp: c2, DateNaissance: shared.NewDate(2015, 1, 1), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
	}
	if count := *partsCount; count < len(parts) {
		parts = parts[:count]
	}

	insc, participants, err := ct.BuildInscription(api.Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Yamina", Prenom: utils.RandString(10, false),
			DateNaissance: shared.NewDate(2000, 1, 1),
			Sexe:          pr.Man,
		},
		Participants: parts,
		Message:      utils.RandString(30, true) + "\n" + utils.RandString(10, true),
	})
	check(err)
	err = utils.InTx(db, func(tx *sql.Tx) error {
		insc, err = in.Create(tx, insc, participants)
		return err
	})
	check(err)
	_, err = api.ConfirmeInscription(db, insc.Id)
	check(err)

	fmt.Println("Added insc. with participants", *partsCount)
}

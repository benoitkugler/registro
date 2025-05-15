package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"registro/config"
	api "registro/controllers/inscriptions"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/events"
	fs "registro/sql/files"
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

// command helper for devs
func main() {
	action := flag.String("action", "", "action to perform: add_insc, add_messages, add_equipier")
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

	asso, err := config.NewAsso()
	check(err)
	smtp, err := config.NewSMTP(false)
	check(err)

	sqlCreds, err := config.NewDB()
	check(err)
	db, err := sqlCreds.ConnectPostgres()
	check(err)

	keys, err := config.NewKeys()
	check(err)
	enc := crypto.NewEncrypter(keys.EncryptKey)

	dirs, err := config.NewDirectories()
	check(err)
	fileSys := fs.NewFileSystem(dirs.Files)

	fmt.Println("Env loaded, using DB:", sqlCreds.Name)

	switch *action {
	case "add_insc":
		addInscriptions(db, smtp, asso, *partsCount)
	case "add_messages":
		addMessages(db)
	case "add_equipier":
		createEquipier(db, enc)
	case "add_structureaide":
		createStructureaide(db)
	case "open_sondages":
		openSondages(db)
	case "create_files":
		createFilesContent(db, fileSys)
	default:
		panic("invalid action")
	}
}

func createFilesContent(db *sql.DB, fileSys fs.FileSystem) {
	files, err := fs.SelectAllFiles(db)
	check(err)
	for _, file := range files {
		err = fileSys.Save(file.Id, testutils.PngData, false)
		check(err)
		err = fileSys.Save(file.Id, testutils.PngData, true)
		check(err)
	}
	fmt.Println("Written files and miniatures:", len(files))
}

func openSondages(db *sql.DB) {
	dossiers, err := ds.SelectAllDossiers(db)
	check(err)
	tmp, err := cps.SelectParticipantsByIdDossiers(db, dossiers.IDs()...)
	check(err)
	participants := tmp.ByIdDossier()
	for _, dossier := range dossiers {
		for idCamp := range participants[dossier.Id].ByIdCamp() {
			event, err := events.Event{IdDossier: dossier.Id, Kind: events.Sondage, Created: time.Now()}.Insert(db)
			check(err)
			err = events.EventSondage{IdEvent: event.Id, IdCamp: idCamp}.Insert(db)
			check(err)
		}
	}
}

func createStructureaide(db *sql.DB) {
	_, err := cps.Structureaide{Nom: "CAF Drôme"}.Insert(db)
	check(err)
	_, err = cps.Structureaide{Nom: "CAF Ardèche"}.Insert(db)
	check(err)
}

func addInscriptions(db *sql.DB, smtp config.SMTP, asso config.Asso, count int) {
	ct := api.NewController(db, crypto.Encrypter{}, smtp, asso)

	// assume we already have two open camps
	camps, _, err := ct.LoadCamps()
	check(err)
	campIds := camps.IDs()
	if len(campIds) < 2 {
		panic("expected at least 2 camps open for inscription")
	}
	c1, c2 := campIds[0], campIds[1]
	parts := []api.Participant{
		{IdCamp: c1, DateNaissance: shared.NewDate(2015, 1, 1), Nom: "Muler", Prenom: "Pierre", Sexe: pr.Man},
		{IdCamp: c1, DateNaissance: shared.NewDate(2000, 1, 1), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
		{IdCamp: c2, DateNaissance: shared.NewDate(2000, 1, 1), Nom: "Martin", Prenom: "Julie", Sexe: pr.Woman},
	}
	if count < len(parts) {
		parts = parts[:count]
	}

	insc, participants, err := ct.BuildInscription(api.Inscription{
		Responsable: in.ResponsableLegal{
			Nom: "Yamina", Prenom: utils.RandString(10, false),
			DateNaissance: shared.NewDate(2000, 1, 1),
			Sexe:          pr.Woman,
			Tels:          pr.Tels{"0684084101", "+33689755468"},
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

	fmt.Println("Added insc. with participants", count)
}

func addMessages(db *sql.DB) {
	// expect one existing dossier
	dossiers, err := ds.SelectAllDossiers(db)
	check(err)
	dossiers.RestrictByValidated(true)

	ids := dossiers.IDs()
	id := ids[rand.Intn(len(ids))]

	_, _, err = events.CreateMessage(db, id, time.Now(), utils.RandString(20, true), events.FromEspaceperso, cps.OptIdCamp{})
	check(err)

	camps, err := cps.SelectAllCamps(db)
	check(err)

	if len(camps) == 0 {
		return
	}
	camp := camps.IDs()[0]
	_, _, err = events.CreateMessage(db, id, time.Now(), utils.RandString(20, true), events.FromDirecteur, camp.Opt())
	check(err)

	fmt.Println("Added messages to dossier", id)
}

// expect at least one camp
func createEquipier(db *sql.DB, enc crypto.Encrypter) {
	camps, err := cps.SelectAllCamps(db)
	check(err)
	camp := camps[camps.IDs()[0]]

	personne, err := pr.Personne{Etatcivil: pr.Etatcivil{Nom: utils.RandString(10, false)}}.Insert(db)
	check(err)

	equipier, err := cps.Equipier{
		IdCamp: camp.Id, IdPersonne: personne.Id,
		Roles: cps.Roles{cps.Adjoint, cps.Infirmerie, cps.Cuisine},
	}.Insert(db)
	check(err)

	builtins, err := fs.LoadBuiltins(db)
	check(err)

	demandes := builtins.Defaut(equipier)
	err = utils.InTx(db, func(tx *sql.Tx) error {
		return fs.InsertManyDemandeEquipiers(tx, demandes...)
	})
	check(err)

	key := crypto.EncryptID(enc, equipier.Id)
	fmt.Printf("Created Equipier key=%s\n", key)
}

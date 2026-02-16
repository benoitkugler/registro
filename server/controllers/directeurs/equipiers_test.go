package directeurs

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"

	"registro/config"
	filesAPI "registro/controllers/files"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"
	tu "registro/utils/testutils"
)

func TestEquipiers(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	asso, smtp := loadEnv(t)
	ct, err := NewController(db.DB, "", "", fs.FileSystem{}, smtp, asso, config.Immich{})
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	pe1, err := pr.Personne{Identite: pr.Identite{
		Mail: "epondrea@free.fr",
	}}.Insert(db)
	tu.AssertNoErr(t, err)

	t.Run("create", func(t *testing.T) {
		out, err := ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: false,
			IdPersonne:     pe1.Id,
			Roles:          cps.Roles{cps.Direction},
		}, camp.Id)
		tu.AssertNoErr(t, err)
		tu.Assert(t, !out.HasBirthday)

		_, err = ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: false,
			IdPersonne:     pe1.Id,
			Roles:          cps.Roles{cps.Adjoint},
		}, camp.Id)
		tu.AssertErr(t, err) // already in

		_, err = ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: true,
			Roles:          cps.Roles{cps.Direction},
		}, camp.Id)
		tu.AssertErr(t, err) // already a direction

		eq, err := ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: true,
			Roles:          cps.Roles{cps.Animation},
		}, camp.Id)
		tu.AssertNoErr(t, err)

		err = ct.deleteEquipier(eq.Equipier.Id, camp.Id+1) // wrong camp
		tu.AssertErr(t, err)
		err = ct.deleteEquipier(eq.Equipier.Id, camp.Id)
		tu.AssertNoErr(t, err)
	})

	t.Run("invite", func(t *testing.T) {
		_, err = ct.createEquipier("", EquipiersCreateIn{
			CreatePersonne: true,
			Nom:            "Kugler", Prenom: "Benoit",
			Roles: cps.Roles{cps.Adjoint},
		}, camp.Id)
		tu.AssertNoErr(t, err)

		eq1, err := ct.createEquipier("", EquipiersCreateIn{CreatePersonne: true, Roles: cps.Roles{cps.Animation}}, camp.Id)
		tu.AssertNoErr(t, err)

		_, err = ct.createEquipier("", EquipiersCreateIn{CreatePersonne: true, Roles: cps.Roles{cps.Animation}}, camp.Id)
		tu.AssertNoErr(t, err)

		err = ct.inviteEquipiers("", EquipiersInviteIn{OnlyOne: eq1.Equipier.Id.Opt()}, eq1.Equipier.IdCamp)
		tu.AssertNoErr(t, err)
		err = ct.inviteEquipiers("", EquipiersInviteIn{}, camp.Id)
		tu.AssertNoErr(t, err)
	})
}

func uploadFile(db *sql.DB, fileSys fs.FileSystem,
	personne pr.IdPersonne, demande fs.IdDemande,
	filename string,
) error {
	_, err := filesAPI.SaveFileFor(fileSys, db, personne, demande, tu.PngData, filename)
	return err
}

func TestDemandes(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)
	for range [10]int{} {
		pe, err := pr.Personne{Identite: pr.Identite{Nom: utils.RandString(8, true)}}.Insert(db)
		tu.AssertNoErr(t, err)
		_, err = cps.Equipier{IdCamp: camp.Id, IdPersonne: pe.Id}.Insert(db)
		tu.AssertNoErr(t, err)
	}

	ct := Controller{db: db.DB, files: fs.NewFileSystem(os.TempDir())}

	l, err := ct.getDemandesEquipiers(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(l.Equipiers) == 10*len(l.Demandes)) // every entry is filled

	err = ct.setDemandeEquipier(EquipiersDemandeSetIn{}, camp.Id+1)
	tu.AssertErr(t, err)
	err = ct.setDemandeEquipier(EquipiersDemandeSetIn{DemandeKey{IdEquipier: 2, IdDemande: 3}, Obligatoire}, camp.Id)
	tu.AssertNoErr(t, err)
	err = ct.setDemandeEquipier(EquipiersDemandeSetIn{DemandeKey{IdEquipier: 3, IdDemande: 3}, Obligatoire}, camp.Id)
	tu.AssertNoErr(t, err)

	l, err = ct.getDemandesEquipiers(camp.Id)
	tu.AssertNoErr(t, err)

	err = uploadFile(ct.db, ct.files, 2, 3, "file1.png")
	tu.AssertNoErr(t, err)
	err = uploadFile(ct.db, ct.files, 3, 3, "file2.png")
	tu.AssertNoErr(t, err)

	files, err := ct.compileFilesEquipiers(camp.Id)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(files) == 2)

	buf := httptest.NewRecorder()
	err = ct.streamFilesEquipiers(camp.Id, buf)
	tu.AssertNoErr(t, err)
	tu.Write(t, "files.zip", buf.Body.Bytes())
}

// func TestEquipe(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	ti := time.Now()
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println("loading equipiers :", time.Since(ti))

// 	ti = time.Now()
// 	liste, err := rc.getEquipe("test")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println("resolving liste :", time.Since(ti))
// 	fmt.Println(liste)
// }

// func TestCreeEquipier(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	pers, err := rd.SelectAllPersonnes(ct.DB)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var idPersonne int64
// 	for _, personne := range pers {
// 		if !personne.IsTemporaire {
// 			idPersonne = personne.Id
// 			break
// 		}
// 	}
// 	equipier1, err := rc.rattacheEquipier(idPersonne, rd.RAideAnimation.AsRoles())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	equipier2, err := rc.ajouteEquipierTmp(matching.PatternsSimilarite{Nom: "KK", Prenom: "Test"}, rd.Roles{rd.RAdjoint, rd.RFactotum})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var eq EquipierDirecteur
// 	eq.Id = equipier1.Id
// 	eq.Nom = "KUGLER"
// 	eq.Prenom = "Estelle"
// 	eq.Roles = rd.RAdjoint.AsRoles()
// 	err = rc.modifieEquipier(eq)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = rc.deleteEquipier(equipier1.Id)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = rc.deleteEquipier(equipier2.Id)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestInviteEquipier(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	equipiers := rc.Camp().GetEquipe(nil)
// 	if len(equipiers) == 0 {
// 		t.Fatal("aucun équipier")
// 	}
// 	err = rc.inviteFormulaireEquipier("localhost:1323", equipiers[0].Id)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestInviteEquipiers(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(rc.Camp().Base.Equipiers) > 3 { // on se restreint à 3 équipiers
// 		ids := rc.Camp().Base.Equipiers.Ids()[0:3]
// 		equipiers := make(rd.Equipiers)
// 		for _, id := range ids {
// 			equipiers[id] = rc.Camp().Base.Equipiers[id]
// 		}
// 		rc.Camp().Base.Equipiers = equipiers
// 	}

// 	err = rc.inviteFormulairesEquipiers("localhost:1323", false)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestDocuments(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	b, err := rc.downloadDocumentsEquipiers(true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(b.Len())
// }

// func TestExportEquipiers(t *testing.T) {
// 	req := newDummyRequest(t, forceComplet)
// 	rc, err := ct.setupRequestComplet(req, loadEquipiers)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	buf, err := rc.exportListeEquipiers()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println("bytes equipiers :", len(buf.Bytes()))
// }

package mails

import (
	"fmt"
	"testing"
	"time"

	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

// func TestPrein(t *testing.T) {
// 	html, err := NewPreinscription("smsld@free.fr", []TargetRespo{
// 		{NomPrenom: "lkdkmslkd", Lien: "http://free.fr"},
// 		{NomPrenom: "sdsd"},
// 		{NomPrenom: "lkdkmssdsdlkd"},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail1_preinscription.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestRenvoieLien(t *testing.T) {
// 	html, err := NewRenvoieLienEspacePerso("smsld@free.fr", []ResumeDossier{
// 		{
// 			Responsable: rd.BasePersonne{Nom: "lkd  kmsslkd", Prenom: "ddd zz"},
// 			Lien:        "http://free.fr",
// 			CampsMap: rd.Camps{
// 				1: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 				2: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 			},
// 		},
// 		{
// 			Responsable: rd.BasePersonne{Nom: "lkdkmslkd", Prenom: "dAadd"},
// 			Lien:        "http://free.fr",
// 			CampsMap: rd.Camps{
// 				1: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 			},
// 		},
// 		{
// 			Responsable: rd.BasePersonne{Nom: "lkd-kmslkd", Prenom: "ddd"},
// 			Lien:        "http://free.fr",
// 			CampsMap: rd.Camps{
// 				1: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 				2: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 				3: {Nom: "C2", DateDebut: rd.Date(time.Now())},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail1_renvoie_lien.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestValide(t *testing.T) {
// 	html, err := NewValideMail(
// 		"http://acve.fr/inscription/valide?data:cryp4tedinscriptin",
// 		Contact{Prenom: "Claudy", Sexe: "F"})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail1_valide_inscription.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestNotifieMessage(t *testing.T) {
	asso, _ := loadEnv(t)

	html, err := NotifieMessage(asso,
		Contact{Prenom: "Claudy", Sexe: pr.Woman},
		"sdlmdmlk\nmsldsm\n\nmsldk! smdlsmdlslùd",
		"https://acve.fr/inscription/valide?data:cryp4tedinscriptin",
	)
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieMessage.html", []byte(html))
}

// func TestDebloqueFS(t *testing.T) {
// 	html, err := NewDebloqueFicheSanitaire(
// 		"http://acve.fr/inscription/valide?data:cryptedinscription",
// 		"bench26@gmail.com", "Audrey Ta utoou")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail7_debloque_fs.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestNotifications(t *testing.T) {
// 	html, err := NewNotifieEnvoiDocs(rd.Camp{
// 		DateDebut: rd.Date(time.Now()),
// 		Nom:       "C2",
// 		Envois:    rd.Envois{Locked: true, ListeVetements: true},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail_notif_doc.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestRenvoieLienJoomeo(t *testing.T) {
// 	html, err := NewRenvoieLienJoomeo("http://joomeo.com", "lekfd8 e", "87s8sd")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile("/home/benoit/Téléchargements/mail_lien_joomeo.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestRenvoieFicheSanitaire(t *testing.T) {
// 	html, err := NewRenvoieLienFicheSanitaire("http://acve.fr/espace_perso/ldmsklmds", "Azur  Evasion 2020")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile("/home/benoit/Téléchargements/mail_lien_fiche_sanitaire.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestNotifieDirecteur(t *testing.T) {
// 	html, err := NewNotifieDirecteur(
// 		Contact{Prenom: "Benoit", Sexe: "F"},
// 		nil,
// 		Responsable{},
// 		nil,
// 		"C2 2019",
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := ioutil.WriteFile(PATH+"local/mail9_directeur.html", []byte(html), 0o666); err != nil {
// 		t.Fatal(err)
// 	}
// 	ccs := []string{}
// 	err = NewMailer(logs.SmtpDev).SendMail("bench26@gmail.com", "Test", html, ccs, DefaultReplyTo)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = NewMailer(logs.SmtpDev).SendMail("bench26@gmail.com", "Test", html, ccs, CustomReplyTo("x.ben.x@free.fr"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestInviteEquipier(t *testing.T) {
	cfg, creds := loadEnv(t)

	html, err := InviteEquipier(cfg,
		fmt.Sprintf("C3 - %d", time.Now().Year()), "Vincent",
		pr.Etatcivil{Prenom: "Cl audie", Sexe: pr.Woman}, "http://test.fr")
	tu.AssertNoErr(t, err)

	tu.Write(t, "InviteEquipier.html", []byte(html))

	err = NewMailer(creds, cfg.MailsSettings).SendMail("", "Test", html, nil, nil)
	tu.AssertNoErr(t, err)
}

func TestNotificationDon(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifieDon(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, dossiers.NewEuros(45.48), "")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifieDon1.html", []byte(html))

	html, err = NotifieDon(cfg, Contact{Prenom: "Beno it", Sexe: pr.Man}, dossiers.NewEuros(45.48), "Eglise de Montmeyran")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifieDon2.html", []byte(html))
}

func TestPreinscription(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := Preinscription(cfg, "xxx.ben@free.fr", []RespoWithLink{
		{"Ben kug", "https://zmldz?454=46"},
		{"Jean claude", "https://zmldz?454=46"},
	})
	tu.AssertNoErr(t, err)
	tu.Write(t, "Preinscription.html", []byte(html))
}

func TestConfirmeInscription(t *testing.T) {
	cfg, creds := loadEnv(t)

	html, err := ConfirmeInscription(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "https://acve.fr/confirme?id='ee'")
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmeInscription.html", []byte(html))

	err = NewMailer(creds, cfg.MailsSettings).SendMail("", "Vérification de l'adresse mail", html, nil, nil)
	tu.AssertNoErr(t, err)
}

func TestNotifieFusionDossier(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifieFusionDossier(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "http://localhost/test")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifieFusionDossier.html", []byte(html))
}

func TestNotifiePlaceLiberee(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifiePlaceLiberee(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "Vive la vie - 2056", "http://localhost/test?placelib=45")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifiePlaceLiberee.html", []byte(html))
}

package mails

import (
	"fmt"
	"testing"
	"time"

	"registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"
)

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

func TestNotifieFacture(t *testing.T) {
	asso, _ := loadEnv(t)

	html, err := NotifieFacture(asso,
		Contact{Prenom: "Claudy", Sexe: pr.Woman},
		"https://acve.fr/inscription/valide?data:cryp4tedinscriptin",
	)
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieFacture.html", []byte(html))
}

func TestNotifieDocumentsCamp(t *testing.T) {
	asso, _ := loadEnv(t)

	html, err := NotifieDocumentsCamp(asso,
		Contact{Prenom: "Claudy", Sexe: pr.Woman},
		"Vive la vie 2025",
		"https://acve.fr/inscription/valide?data:cryp4tedinscriptin",
	)
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieDocumentsCamp.html", []byte(html))
}

func TestNotifieSondage(t *testing.T) {
	asso, _ := loadEnv(t)

	html, err := NotifieSondage(asso,
		Contact{Prenom: "Claudy", Sexe: pr.Woman},
		"Vive la vie 2025",
		"https://acve.fr/inscription/valide?data:cryp4tedinscriptin",
	)
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieSondage.html", []byte(html))
}

func TestInviteEquipier(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := InviteEquipier(cfg,
		fmt.Sprintf("C3 - %d", time.Now().Year()), "Vincent",
		pr.Identite{Prenom: "Cl audie", Sexe: pr.Woman}, "http://test.fr")
	tu.AssertNoErr(t, err)

	tu.Write(t, "InviteEquipier.html", []byte(html))
}

func TestNotifieModificationOptions(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifieModificationOptions(cfg,
		pr.Identite{Prenom: "Cl audie", Sexe: pr.Woman}, fmt.Sprintf("C3 - %d", time.Now().Year()), []string{
			"Vincent",
		}, "http://test.fr")
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieModificationOptions_1.html", []byte(html))

	html, err = NotifieModificationOptions(cfg,
		pr.Identite{Prenom: "Cl audie", Sexe: pr.Woman}, fmt.Sprintf("C3 - %d", time.Now().Year()), []string{
			"Vincent",
			"Beoit Kugler",
		}, "http://test.fr")
	tu.AssertNoErr(t, err)

	tu.Write(t, "NotifieModificationOptions_2.html", []byte(html))
}

func TestNotificationDon(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifieDon(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, dossiers.NewEuros(45.48))
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifieDon1.html", []byte(html))

	html, err = NotifieDon(cfg, Contact{Prenom: "Eglise de Montmeyran"}, dossiers.NewEuros(45.48))
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

func TestValidationMailInscription(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := ValidationMailInscription(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "https://acve.fr/confirme?id='ee'")
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmeInscription.html", []byte(html))
}

func TestConfirmationInscription(t *testing.T) {
	cfg, _ := loadEnv(t)
	contact, url := Contact{Prenom: "Benoit", Sexe: pr.Woman}, "https://acve.fr/confirme?id='ee'"

	html, err := ConfirmationInscription(cfg, contact, url,
		[]Participant{{"Benoit Kugler", "C2 2025"}},
		[]Participant{},
		[]Participant{},
	)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmationInscription_1.html", []byte(html))

	html, err = ConfirmationInscription(cfg, contact, url,
		[]Participant{{"Benoit Kugler", "C2 2025"}, {"Benoit Kugler", "C3 2025"}},
		[]Participant{},
		[]Participant{},
	)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmationInscription_2.html", []byte(html))

	html, err = ConfirmationInscription(cfg, contact, url,
		[]Participant{},
		[]Participant{{"Benoit Kugler", "C2 2025"}},
		[]Participant{},
	)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmationInscription_3.html", []byte(html))

	html, err = ConfirmationInscription(cfg, contact, url,
		[]Participant{},
		[]Participant{{"Benoit Kugler", "C2 2025"}, {"Benoit Kugler", "C3 2025"}},
		[]Participant{},
	)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmationInscription_4.html", []byte(html))

	html, err = ConfirmationInscription(cfg, contact, url,
		[]Participant{{"Benoit Kugler", "C2 2025"}, {"Benoit Kugler", "C3 2025"}},
		[]Participant{{"Benoit Kugler", "C2 2025"}, {"Benoit Kugler", "C3 2025"}},
		[]Participant{{"Benoit Kugler", "C2 2025"}, {"Benoit Kugler", "C3 2025"}},
	)
	tu.AssertNoErr(t, err)
	tu.Write(t, "ConfirmationInscription_5.html", []byte(html))
}

func TestNotifieFusionDossier(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifieFusionDossier(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "http://localhost/test")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifieFusionDossier.html", []byte(html))
}

func TestTransfertFicheSanitaire(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := TransfertFicheSanitaire(cfg, "http://localhost/test", "dulmmy@free.fr", "Bneoit Kugler")
	tu.AssertNoErr(t, err)
	tu.Write(t, "TransfertFicheSanitaire.html", []byte(html))
}

func TestNotifiePlaceLiberee(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := NotifiePlaceLiberee(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "Vive la vie - 2056", "http://localhost/test?placelib=45")
	tu.AssertNoErr(t, err)
	tu.Write(t, "NotifiePlaceLiberee.html", []byte(html))
}

func TestRelanceDocuments(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := RelanceDocuments(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "Vive la vie - 2056", "Julie", "http://localhost/test?placelib=45")
	tu.AssertNoErr(t, err)
	tu.Write(t, "RelanceDocuments.html", []byte(html))
}

func TestRenvoieEspacePersoURL(t *testing.T) {
	cfg, _ := loadEnv(t)
	html, err := RenvoieEspacePersoURL(cfg, "smsld@free.fr", []ResumeDossier{
		{
			Responsable: "lkd  kmsslkd" + " ddd zz",
			URL:         "http://free.fr",
			CampsMap: camps.Camps{
				1: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
				2: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
			},
		},
		{
			Responsable: "lkdkmslkd" + " dAadd",
			URL:         "http://free.fr",
			CampsMap: camps.Camps{
				1: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
			},
		},
		{
			Responsable: "lkd-kmslkd" + " ddd",
			URL:         "http://free.fr",
			CampsMap: camps.Camps{
				1: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
				2: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
				3: {Nom: "C2", DateDebut: shared.NewDateFrom(time.Now())},
			},
		},
	})
	tu.AssertNoErr(t, err)
	tu.Write(t, "RenvoieEspacePersoURL.html", []byte(html))
}

func TestSendPhotosLink(t *testing.T) {
	cfg, _ := loadEnv(t)

	html, err := SendPhotosLinkInscrits(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "Vive la vie - 2056", "http://localhost/test?placelib=45")
	tu.AssertNoErr(t, err)
	tu.Write(t, "SendPhotosLinkInscrits.html", []byte(html))

	html, err = SendPhotosLinkEquipiers(cfg, Contact{Prenom: "Benoit", Sexe: pr.Woman}, "Vive la vie - 2056", "http://localhost/test?placelib=45")
	tu.AssertNoErr(t, err)
	tu.Write(t, "SendPhotosLinkEquipiers.html", []byte(html))
}

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

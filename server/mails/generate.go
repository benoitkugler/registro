package mails

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"registro/config"
	"registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
)

//go:embed templates/*
var templates embed.FS

var (
	inviteEquipierT             *template.Template
	notifieDonT                 *template.Template
	validationMailInscriptionT  *template.Template
	preinscriptionT             *template.Template
	notifieFusionDossierT       *template.Template
	notifieMessageT             *template.Template
	notifieFactureT             *template.Template
	notifieDocumentsCampT       *template.Template
	notifieSondageT             *template.Template
	notifiePlaceLibereeT        *template.Template
	confirmationInscriptionT    *template.Template
	notifieModificationOptionsT *template.Template
	transfertFicheSanitaireT    *template.Template
	relanceDocumentsT           *template.Template
	renvoieEspacePersoURLT      *template.Template
	sendPhotosLinkInscritsT     *template.Template
	sendPhotosLinkEquipiersT    *template.Template
)

func init() {
	inviteEquipierT = parseTemplate("templates/inviteEquipier.html")
	notifieDonT = parseTemplate("templates/notifieDon.html")
	validationMailInscriptionT = parseTemplate("templates/validationMailInscription.html")
	preinscriptionT = parseTemplate("templates/preinscription.html")
	notifieFusionDossierT = parseTemplate("templates/notifieFusionDossier.html")
	notifieMessageT = parseTemplate("templates/notifieMessage.html")
	notifieFactureT = parseTemplate("templates/notifieFacture.html")
	notifieDocumentsCampT = parseTemplate("templates/notifieDocumentsCamp.html")
	notifieSondageT = parseTemplate("templates/notifieSondage.html")
	notifiePlaceLibereeT = parseTemplate("templates/notifiePlaceLiberee.html")
	confirmationInscriptionT = parseTemplate("templates/confirmationInscription.html")
	notifieModificationOptionsT = parseTemplate("templates/notifieModificationOptions.html")
	transfertFicheSanitaireT = parseTemplate("templates/transfertFicheSanitaire.html")
	relanceDocumentsT = parseTemplate("templates/relanceDocuments.html")
	renvoieEspacePersoURLT = parseTemplate("templates/renvoieEspacePersoURL.html")
	sendPhotosLinkInscritsT = parseTemplate("templates/sendPhotosLinkInscrits.html")
	sendPhotosLinkEquipiersT = parseTemplate("templates/sendPhotosLinkEquipiers.html")
}

func parseTemplate(templateFile string) *template.Template {
	main := template.Must(template.New("").ParseFS(templates, "templates/defs.html", "templates/main.html"))

	_, err := main.New("_").ParseFS(templates, templateFile)
	if err != nil {
		panic(err)
	}

	return main
}

func render(temp *template.Template, data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	err := temp.ExecuteTemplate(buf, "main.html", data)
	return buf.String(), err
}

type Contact struct {
	Prenom string
	Sexe   pr.Sexe
}

func NewContact(personne *pr.Personne) Contact {
	return Contact{Prenom: personne.FPrenom(), Sexe: personne.Sexe}
}

func (c Contact) Salutations() string {
	if c.Prenom == "" {
		return "Bonjour,"
	}
	var out string
	switch c.Sexe {
	case pr.Man:
		out = "Cher"
	case pr.Woman:
		out = "Chère"
	default:
		out = "Bonjour"
	}
	return fmt.Sprintf("%s %s,", out, c.Prenom)
}

const mailAutoSignature = template.HTML("<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>")

type Participant struct {
	Personne string
	Camp     string
}

type champsCommuns struct {
	Title       string
	Salutations string
	Signature   template.HTML

	Asso config.Asso
}

func NotifieMessage(asso config.Asso, contact Contact, contenu, lienEspacePerso string) (string, error) {
	contenu = strings.ReplaceAll(contenu, "\n", "<br/>")
	args := struct {
		champsCommuns
		Contenu                template.HTML
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Nouveau message",
			Salutations: contact.Salutations(),
			Asso:        asso,
			Signature:   mailAuto,
		},
		Contenu:                template.HTML(contenu),
		EspacePersoURL:         lienEspacePerso,
		EspacePersoButtonLabel: "MON ESPACE",
	}
	return render(notifieMessageT, args)
}

func NotifieFacture(asso config.Asso, contact Contact, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Demande de règlement",
			Salutations: contact.Salutations(),
			Asso:        asso,
			Signature:   mailAuto,
		},
		EspacePersoURL:         lienEspacePerso,
		EspacePersoButtonLabel: "MON ESPACE",
	}
	return render(notifieFactureT, args)
}

func NotifieDocumentsCamp(asso config.Asso, contact Contact, campLabel string, espacePersoURL string) (string, error) {
	args := struct {
		champsCommuns
		CampLabel              string
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Documents à remplir",
			Salutations: contact.Salutations(),
			Asso:        asso,
			Signature:   mailAuto,
		},
		CampLabel:              campLabel,
		EspacePersoURL:         espacePersoURL,
		EspacePersoButtonLabel: "COMPLÉTER LES DOCUMENTS",
	}
	return render(notifieDocumentsCampT, args)
}

func RelanceDocuments(cfg config.Asso, contact Contact, campLabel, prenom, espacePersoURL string) (string, error) {
	args := struct {
		champsCommuns
		Contact                Contact
		Camp                   string
		Prenom                 string
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Documents à remplir",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		Contact:                contact,
		Camp:                   campLabel,
		Prenom:                 prenom,
		EspacePersoURL:         espacePersoURL,
		EspacePersoButtonLabel: "COMPLÉTER LES DOCUMENTS",
	}
	return render(relanceDocumentsT, args)
}

func NotifieSondage(asso config.Asso, contact Contact, campLabel string, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		CampLabel              string
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Avis sur le séjour",
			Salutations: contact.Salutations(),
			Asso:        asso,
			Signature:   mailAuto,
		},
		CampLabel:              campLabel,
		EspacePersoURL:         lienEspacePerso,
		EspacePersoButtonLabel: "RÉPONDRE",
	}
	return render(notifieSondageT, args)
}

type RespoWithLink struct {
	NomPrenom string
	Lien      template.HTML
}

func Preinscription(asso config.Asso, mail string, responsables []RespoWithLink) (string, error) {
	args := struct {
		champsCommuns
		Mail         string
		Responsables []RespoWithLink
	}{
		champsCommuns: champsCommuns{
			Title:       "Inscription rapide",
			Salutations: Contact{}.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		Mail:         mail,
		Responsables: responsables,
	}
	return render(preinscriptionT, args)
}

func ValidationMailInscription(asso config.Asso, contact Contact, urlConfirmeInscription string) (string, error) {
	args := struct {
		champsCommuns
		MailCentre             string
		EspacePersoURL         template.HTML
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Validation de l'adresse mail",
			Salutations: contact.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		MailCentre:             asso.ContactMail,
		EspacePersoURL:         template.HTML(urlConfirmeInscription),
		EspacePersoButtonLabel: "VALIDER MON ADRESSE",
	}

	return render(validationMailInscriptionT, args)
}

func ConfirmationInscription(asso config.Asso, contact Contact, lienEspacePerso string, inscrits, attente, astatuer []Participant) (string, error) {
	args := struct {
		champsCommuns
		Inscrits               []Participant
		Attente                []Participant
		AStatuer               []Participant
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Inscription confirmée",
			Salutations: contact.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		Inscrits:               inscrits,
		Attente:                attente,
		AStatuer:               astatuer,
		EspacePersoURL:         lienEspacePerso,
		EspacePersoButtonLabel: "Mon espace",
	}

	return render(confirmationInscriptionT, args)
}

func TransfertFicheSanitaire(asso config.Asso, urlDebloqueFicheSanitaire, newMail, participant string) (string, error) {
	args := struct {
		champsCommuns
		NewMail                string
		Participant            string
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Accès à la fiche sanitaire",
			Salutations: Contact{}.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		NewMail:                newMail,
		Participant:            participant,
		EspacePersoURL:         urlDebloqueFicheSanitaire,
		EspacePersoButtonLabel: "PARTAGER",
	}
	return render(transfertFicheSanitaireT, args)
}

func InviteEquipier(cfg config.Asso, labelCamp string, directeur string, equipier pr.Identite, lienFormulaire string) (string, error) {
	s := "Cher"
	if equipier.Sexe == pr.Woman {
		s = "Chère"
	}

	args := struct {
		champsCommuns
		LabelCamp      string
		LienFormulaire string
	}{
		champsCommuns{
			Title:       "Bienvenue dans l'équipe !",
			Salutations: fmt.Sprintf("%s %s,", s, equipier.FPrenom()),
			Asso:        cfg,
			Signature:   template.HTML(directeur),
		},
		labelCamp,
		lienFormulaire,
	}
	return render(inviteEquipierT, args)
}

func NotifieModificationOptions(cfg config.Asso, directeur pr.Identite, camp string, participants []string, urlDirecteur string) (string, error) {
	s := "Cher"
	if directeur.Sexe == pr.Woman {
		s = "Chère"
	}

	args := struct {
		champsCommuns
		Camp         string
		Participants []string
		URL          string
	}{
		champsCommuns{
			Title:       "Modification des options",
			Salutations: fmt.Sprintf("%s %s,", s, directeur.FPrenom()),
			Asso:        cfg,
			Signature:   mailAutoSignature,
		},
		camp,
		participants,
		urlDirecteur,
	}
	return render(notifieModificationOptionsT, args)
}

// organisme est vide pour les dons particulier
func NotifieDon(cfg config.Asso, contact Contact, montant dossiers.Montant) (string, error) {
	args := struct {
		champsCommuns
		Montant string
	}{
		champsCommuns: champsCommuns{
			Title:       "Merci pour votre don !",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   "L'équipe " + template.HTML(cfg.Title),
		},
		Montant: montant.String(),
	}
	return render(notifieDonT, args)
}

const mailAuto template.HTML = "<i>PS: Merci de ne pas répondre directement à ce mail mais d'utiliser votre espace de suivi (ci-dessus).</i>"

func NotifieFusionDossier(cfg config.Asso, contact Contact, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		EspacePersoURL string
	}{
		champsCommuns: champsCommuns{
			Title:       "Fusion de votre dossier",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		EspacePersoURL: lienEspacePerso,
	}
	return render(notifieFusionDossierT, args)
}

func NotifiePlaceLiberee(cfg config.Asso, contact Contact, camp string, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		Contact                Contact
		Camp                   string
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Place disponible",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		Contact:                contact,
		Camp:                   camp,
		EspacePersoURL:         lienEspacePerso,
		EspacePersoButtonLabel: "MON ESPACE",
	}
	return render(notifiePlaceLibereeT, args)
}

type ResumeDossier struct {
	Responsable string
	URL         template.HTML
	CampsMap    camps.Camps
}

func (r ResumeDossier) Camps() string {
	var tmp []string
	for _, camp := range r.CampsMap {
		tmp = append(tmp, camp.Label())
	}
	return strings.Join(tmp, ", ")
}

func RenvoieEspacePersoURL(cfg config.Asso, mail string, dossiers []ResumeDossier) (string, error) {
	args := struct {
		champsCommuns
		Mail     string
		Dossiers []ResumeDossier
	}{
		champsCommuns: champsCommuns{
			Title:       "Lien de suivi",
			Salutations: Contact{}.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		Mail:     mail,
		Dossiers: dossiers,
	}
	return render(renvoieEspacePersoURLT, args)
}

func SendPhotosLinkInscrits(cfg config.Asso, contact Contact, campLabel, albumURL string) (string, error) {
	args := struct {
		champsCommuns
		Camp                   string
		Dossiers               []ResumeDossier
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Album photos",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/>",
		},
		Camp:                   campLabel,
		EspacePersoURL:         albumURL,
		EspacePersoButtonLabel: "Album photos",
	}
	return render(sendPhotosLinkInscritsT, args)
}

func SendPhotosLinkEquipiers(cfg config.Asso, contact Contact, campLabel, albumURL string) (string, error) {
	args := struct {
		champsCommuns
		Camp                   string
		Dossiers               []ResumeDossier
		EspacePersoURL         string
		EspacePersoButtonLabel string
	}{
		champsCommuns: champsCommuns{
			Title:       "Album photos",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/>",
		},
		Camp:                   campLabel,
		EspacePersoURL:         albumURL,
		EspacePersoButtonLabel: "Album photos",
	}
	return render(sendPhotosLinkEquipiersT, args)
}

// func NewRenvoieLienJoomeo(lien, login, password string) (string, error) {
// 	commun := newChampCommuns(Contact{}, "Espace photo")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsRenvoieLienJoomeo{
// 		champsCommuns: commun,
// 		Lien:          lien,
// 		Login:         login,
// 		Password:      password,
// 	}
// 	return render(templates.RenvoieLienJoomeo, "base.html", p)
// }

// func NewRenvoieLienFicheSanitaire(lien, sejour string) (string, error) {
// 	commun := newChampCommuns(Contact{}, "Fiche sanitaire")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsRenvoieLienFicheSanitaire{
// 		champsCommuns: commun,
// 		Lien:          lien,
// 		Sejour:        sejour,
// 	}
// 	return render(templates.RenvoieLienFicheSanitaire, "base.html", p)
// }

// // paramsNotifieDirecteur est à compléter
// func NewNotifieDirecteur(directeur Contact, participants []Participant, responsable Responsable,
// 	infoLines []string, labelCamp string,
// ) (string, error) {
// 	p := paramsNotifieDirecteur{
// 		champsCommuns: newChampCommuns(directeur, "Nouvelle inscription"),
// 		Directeur:     directeur,
// 		Participants:  participants,
// 		Responsable:   responsable,
// 		InfoLines:     infoLines,
// 		LabelCamp:     labelCamp,
// 	}
// 	return render(templates.NotifDirecteur, "base.html", p)
// }

// func NewNotifieEnvoiDocs(camp rd.Camp) (string, error) {
// 	p := paramsNotifEnvoisDocs{
// 		Envois:    camp.Envois,
// 		LabelCamp: camp.Label().String(),
// 	}
// 	return render(templates.NotifEnvoisDocs, "notif_envois_docs.html", p)
// }

// type paramsValideMail struct {
// 	champsCommuns
// 	UrlValideInscription string
// }

// type paramsDebloqueFicheSanitaire struct {
// 	champsCommuns
// 	NomPrenomParticipant      string
// 	NewMail                   string
// 	UrlDebloqueFicheSanitaire string
// }

// type paramsNotifieMessage struct {
// 	champsCommuns
// 	Contenu         string
// 	EspacePersoURL string
// }

// type paramsNotifFusion struct {
// 	champsCommuns
// 	EspacePersoURL string
// }

// type paramsAccuseReceptionSimple struct {
// 	Sejour string
// 	champsCommuns
// }

// type paramsNotifEnvoisDocs struct {
// 	LabelCamp string
// 	Envois    rd.Envois
// }

// type Responsable struct {
// 	Contact
// 	Mail, Tels string
// }

// type paramsNotifieDirecteur struct {
// 	champsCommuns
// 	Directeur    Contact
// 	Participants []Participant
// 	Responsable  Responsable
// 	InfoLines    []string
// 	LabelCamp    string
// }

// type paramsRenvoieLienJoomeo struct {
// 	champsCommuns
// 	Lien     string
// 	Login    string
// 	Password string
// }

// type paramsRenvoieLienFicheSanitaire struct {
// 	champsCommuns
// 	Lien   string
// 	Sejour string
// }

// func newChampCommuns(contact Contact, title string) champsCommuns {
// 	return champsCommuns{
// 		Contact:       contact,
// 		Title:         title,
// 		FooterTitle:   rd.Asso.Title,
// 		FooterInfos:   rd.Asso.Infos,
// 		SignatureMail: rd.SignatureMail,
// 	}
// }

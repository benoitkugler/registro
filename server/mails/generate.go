package mails

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"registro/config"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
)

//go:embed templates/*
var templates embed.FS

var (
	inviteEquipierT               *template.Template
	notifieDonT                   *template.Template
	confirmeInscriptionT          *template.Template
	preinscriptionT               *template.Template
	notifieFusionDossierT         *template.Template
	notifieMessageT               *template.Template
	notifiePlaceLibereeT          *template.Template
	notifieValidationInscriptionT *template.Template
)

func init() {
	inviteEquipierT = parseTemplate("templates/inviteEquipier.html")
	notifieDonT = parseTemplate("templates/notifieDon.html")
	confirmeInscriptionT = parseTemplate("templates/confirmeInscription.html")
	preinscriptionT = parseTemplate("templates/preinscription.html")
	notifieFusionDossierT = parseTemplate("templates/notifieFusionDossier.html")
	notifieMessageT = parseTemplate("templates/notifieMessage.html")
	notifiePlaceLibereeT = parseTemplate("templates/notifiePlaceLiberee.html")
	notifieValidationInscriptionT = parseTemplate("templates/notifieValidationInscription.html")
}

func parseTemplate(templateFile string) *template.Template {
	main := template.Must(template.New("").ParseFS(templates, "templates/main.html"))

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

// type DetailsParticipant struct {
// 	Participant
// 	Attente            rd.ListeAttente
// 	Groupe             string
// 	NeedFicheSanitaire bool
// }

type champsCommuns struct {
	Title       string
	Salutations string
	Signature   template.HTML

	Asso config.Asso
}

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
// 	LienEspacePerso string
// }

// type paramsNotifFusion struct {
// 	champsCommuns
// 	LienEspacePerso string
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

// type ResumeDossier struct {
// 	Responsable rd.BasePersonne
// 	Lien        string
// 	CampsMap    rd.Camps
// }

// func (r ResumeDossier) Camps() string {
// 	var tmp []string
// 	for _, camp := range r.CampsMap {
// 		tmp = append(tmp, camp.Label().String())
// 	}
// 	return strings.Join(tmp, ", ")
// }

// type paramsRenvoieLienEspacePerso struct {
// 	champsCommuns
// 	Mail     string
// 	Dossiers []ResumeDossier
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

func NotifieMessage(asso config.Asso, contact Contact, contenu, lienEspacePerso string) (string, error) {
	contenu = strings.ReplaceAll(contenu, "\n", "<br/>")
	args := struct {
		champsCommuns
		Contenu         template.HTML
		LienEspacePerso string
	}{
		champsCommuns: champsCommuns{
			Title:       "Nouveau message",
			Salutations: contact.Salutations(),
			Asso:        asso,
			Signature:   mailAuto,
		},
		Contenu:         template.HTML(contenu),
		LienEspacePerso: lienEspacePerso,
	}
	return render(notifieMessageT, args)
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

// func NewRenvoieLienEspacePerso(mail string, dossiers []ResumeDossier) (string, error) {
// 	commun := newChampCommuns(Contact{}, "Espace de suivi")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsRenvoieLienEspacePerso{
// 		champsCommuns: commun,
// 		Mail:          mail,
// 		Dossiers:      dossiers,
// 	}
// 	return render(templates.RenvoieLienEspacePerso, "base.html", p)
// }

func ConfirmeInscription(asso config.Asso, contact Contact, urlConfirmeInscription string) (string, error) {
	args := struct {
		champsCommuns
		URL template.HTML
	}{
		champsCommuns: champsCommuns{
			Title:       "Validation de l'adresse mail",
			Salutations: contact.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		URL: template.HTML(urlConfirmeInscription),
	}

	return render(confirmeInscriptionT, args)
}

func NotifieValidationInscription(asso config.Asso, contact Contact, lienEspacePerso string, inscrits, attente, astatuer []Participant) (string, error) {
	args := struct {
		champsCommuns
		Inscrits        []Participant
		Attente         []Participant
		AStatuer        []Participant
		LienEspacePerso string
	}{
		champsCommuns: champsCommuns{
			Title:       "Inscription reçue",
			Salutations: contact.Salutations(),
			Signature:   mailAutoSignature,
			Asso:        asso,
		},
		Inscrits:        inscrits,
		Attente:         attente,
		AStatuer:        astatuer,
		LienEspacePerso: lienEspacePerso,
	}

	return render(notifieValidationInscriptionT, args)
}

// func NewDebloqueFicheSanitaire(urlDebloqueFicheSanitaire, newMail, nomPrenom string) (string, error) {
// 	commun := newChampCommuns(Contact{}, "Accès à la fiche sanitaire")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsDebloqueFicheSanitaire{
// 		UrlDebloqueFicheSanitaire: urlDebloqueFicheSanitaire,
// 		NewMail:                   newMail,
// 		NomPrenomParticipant:      nomPrenom,
// 		champsCommuns:             commun,
// 	}
// 	return render(templates.DebloqueFicheSanitaire, "base.html", p)
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

func InviteEquipier(cfg config.Asso, labelCamp string, directeur string, equipier pr.Etatcivil, lienFormulaire string) (string, error) {
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
			Signature:   template.HTML(directeur) + "<br/>" + mailAutoSignature,
		},
		labelCamp,
		lienFormulaire,
	}
	return render(inviteEquipierT, args)
}

// organisme est vide pour les dons particulier
func NotifieDon(cfg config.Asso, contact Contact, montant dossiers.Montant, organisme string) (string, error) {
	args := struct {
		champsCommuns
		Montant   string
		Organisme string
	}{
		champsCommuns: champsCommuns{
			Title:       "Merci pour votre don !",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   "L'équipe " + template.HTML(cfg.Title),
		},
		Montant:   montant.String(),
		Organisme: organisme,
	}
	return render(notifieDonT, args)
}

const mailAuto template.HTML = "<i>PS: Merci de ne pas répondre directement à ce mail mais d'utiliser votre espace de suivi (ci-dessus).</i>"

func NotifieFusionDossier(cfg config.Asso, contact Contact, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		LienEspacePerso string
	}{
		champsCommuns: champsCommuns{
			Title:       "Fusion de votre dossier",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		LienEspacePerso: lienEspacePerso,
	}
	return render(notifieFusionDossierT, args)
}

func NotifiePlaceLiberee(cfg config.Asso, contact Contact, camp string, lienEspacePerso string) (string, error) {
	args := struct {
		champsCommuns
		Contact         Contact
		Camp            string
		LienEspacePerso string
	}{
		champsCommuns: champsCommuns{
			Title:       "Place disponible",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   cfg.MailsSettings.SignatureMailCentre + "<br/><br/>" + mailAuto,
		},
		Contact:         contact,
		Camp:            camp,
		LienEspacePerso: lienEspacePerso,
	}
	return render(notifiePlaceLibereeT, args)
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

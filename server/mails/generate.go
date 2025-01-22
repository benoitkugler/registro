package mails

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"registro/config"
	"registro/sql"
	pr "registro/sql/personnes"
)

//go:embed templates/*
var templates embed.FS

var (
	inviteEquipierTmpl *template.Template
	notifieDonTmpl     *template.Template
)

func init() {
	inviteEquipierTmpl = parseTemplate("templates/invite_equipier.html")
	notifieDonTmpl = parseTemplate("templates/notifie_don.html")
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

// type MailRenderer interface {
// 	Render(messagePerso string) (string, error)
// }

// type parsedTemplates struct {
// 	NotifieMessage            *template.Template
// 	Preinscription            *template.Template
// 	ValideMail                *template.Template
// 	DebloqueFicheSanitaire    *template.Template
// 	AccuseReceptionSimple     *template.Template
// 	NotifEnvoisDocs           *template.Template
// 	NotifDirecteur            *template.Template
// 	InviteEquipier            *template.Template
// 	RenvoieLienEspacePerso    *template.Template
// 	NotificationDon           *template.Template
// 	NotifFusion               *template.Template
// 	RenvoieLienJoomeo         *template.Template
// 	RenvoieLienFicheSanitaire *template.Template
// }

// // InitTemplates charge les templates depuis le dossier
// // donné par `ressourcesPath`
// func InitTemplates(ressourcesPath string) {
// 	fp := func(filename string) string {
// 		return filepath.Join(ressourcesPath, "templates_mails", filename)
// 	}
// 	templates.NotifieMessage = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("notifie_message.html")))

// 	templates.Preinscription = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("preinscription.html")))

// 	templates.ValideMail = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("valide_mail.html")))

// 	templates.DebloqueFicheSanitaire = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("debloque_fiche_sanitaire.html")))

// 	templates.AccuseReceptionSimple = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("accuse_reception_simple.html"), fp("coordonnees_centre.html")))

// 	templates.NotifEnvoisDocs = template.Must(template.New("").Funcs(FuncMap).ParseFiles(fp("notif_envois_docs.html")))

// 	templates.NotifDirecteur = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("notifie_directeur.html")))

// 	templates.InviteEquipier = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("invite_equipier.html")))

// 	templates.RenvoieLienEspacePerso = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("renvoie_lien_espace_perso.html")))

// 	templates.NotificationDon = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("notifie_don.html")))

// 	templates.NotifFusion = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("notifie_fusion_dossier.html")))

// 	templates.RenvoieLienJoomeo = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("renvoie_lien_joomeo.html")))

// 	templates.RenvoieLienFicheSanitaire = template.Must(template.New("").Funcs(FuncMap).ParseFiles(
// 		fp("base.html"), fp("renvoie_lien_fiche_sanitaire.html")))
// }

type Contact struct {
	Prenom string
	Sexe   pr.Sex
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

// type Participant struct {
// 	NomPrenom, Sexe string
// 	Prenom          string
// 	DateNaissance   rd.Date
// 	LabelCamp       string
// }

// type DetailsParticipant struct {
// 	Participant
// 	Attente            rd.ListeAttente
// 	Groupe             string
// 	NeedFicheSanitaire bool
// }

type champsCommuns struct {
	Salutations string
	Title       string
	Signature   template.HTML

	Asso config.Asso
}

// type TargetRespo struct {
// 	Lien, NomPrenom string
// }

// type paramsPreinscription struct {
// 	champsCommuns
// 	Mail         string
// 	Responsables []TargetRespo
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

// func NewAccuseReceptionSimple(camp rd.Camp, contact Contact) (string, error) {
// 	commun := newChampCommuns(contact, "Inscription reçue")
// 	p := paramsAccuseReceptionSimple{
// 		Sejour:        camp.Label().String(),
// 		champsCommuns: commun,
// 	}
// 	return render(templates.AccuseReceptionSimple, "base.html", p)
// }

// func NewNotifieMessage(contact Contact, title, contenu, lienEspacePerso string) (string, error) {
// 	p := paramsNotifieMessage{
// 		champsCommuns:   newChampCommuns(contact, title),
// 		Contenu:         contenu,
// 		LienEspacePerso: lienEspacePerso,
// 	}
// 	p.SignatureMail += "<br/><br/><i>Merci de ne pas répondre directement à ce mail mais d'utiliser votre espace de suivi (ci-dessus).</i>"
// 	return render(templates.NotifieMessage, "base.html", p)
// }

// func NewPreinscription(mail string, resp []TargetRespo) (string, error) {
// 	commun := newChampCommuns(Contact{}, "Inscription rapide")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsPreinscription{
// 		champsCommuns: commun,
// 		Mail:          mail,
// 		Responsables:  resp,
// 	}
// 	return render(templates.Preinscription, "base.html", p)
// }

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

// func NewValideMail(urlValideInscription string, contact Contact) (string, error) {
// 	commun := newChampCommuns(contact, "Confirmation de l'adresse mail")
// 	commun.SignatureMail = "<i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"
// 	p := paramsValideMail{
// 		UrlValideInscription: urlValideInscription,
// 		champsCommuns:        commun,
// 	}
// 	return render(templates.ValideMail, "base.html", p)
// }

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

func InviteEquipier(cfg config.Asso, labelCamp string, directeur string, equipier pr.Etatcivil, lienForumaire string) (string, error) {
	s := "Cher"
	if equipier.Sexe == pr.Woman {
		s = "Chère"
	}

	p := struct {
		champsCommuns
		LabelCamp      string
		LienFormulaire string
	}{
		champsCommuns{
			Title:       "Bienvenue dans l'équipe !",
			Salutations: fmt.Sprintf("%s %s,", s, equipier.FPrenom()),
			Asso:        cfg,
			Signature:   template.HTML(directeur + "<br/><i>Ps : Ceci est un mail automatique, merci de ne pas y répondre.</i>"),
		},
		labelCamp,
		lienForumaire,
	}
	return render(inviteEquipierTmpl, p)
}

// organisme est vide pour les dons particulier
func NotifieDon(cfg config.Asso, contact Contact, montant sql.Montant, organisme string) (string, error) {
	p := struct {
		champsCommuns
		Montant   string
		Organisme string
	}{
		champsCommuns: champsCommuns{
			Title:       "Merci pour votre don !",
			Salutations: contact.Salutations(),
			Asso:        cfg,
			Signature:   "L'équipe ACVE",
		},
		Montant:   montant.String(),
		Organisme: organisme,
	}
	return render(notifieDonTmpl, p)
}

// func NewNotifFusion(contact Contact, lienEspacePerso string) (string, error) {
// 	p := paramsNotifFusion{
// 		champsCommuns:   newChampCommuns(contact, "Fusion de votre dossier"),
// 		LienEspacePerso: lienEspacePerso,
// 	}
// 	p.SignatureMail += "<br/><br/><i>Merci de ne pas répondre directement à ce mail mais d'utiliser votre espace de suivi (ci-dessus).</i>"
// 	return render(templates.NotifFusion, "base.html", p)
// }

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

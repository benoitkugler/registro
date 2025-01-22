package pdfcreator

import (
	"embed"
	"html/template"

	"registro/config"
	"registro/sql/camps"
	pr "registro/sql/personnes"
)

//go:embed templates/*
var templates embed.FS

var (
	ficheSanitaireTmpl    *template.Template
	listeParticipantsTmpl *template.Template
	listeVetementsTmpl    *template.Template
)

func init() {
	ficheSanitaireTmpl = parseTemplate("templates/fiche_sanitaire.html")
	listeParticipantsTmpl = parseTemplate("templates/liste_participants.html")
	listeVetementsTmpl = parseTemplate("templates/liste_vetements.html")
}

func parseTemplate(templateFile string) *template.Template {
	main := template.Must(template.New("").ParseFS(templates, "templates/main.html"))

	_, err := main.New("_").ParseFS(templates, templateFile)
	if err != nil {
		panic(err)
	}

	return main
}

type FicheSanitaire struct {
	Personne       pr.Etatcivil
	FicheSanitaire pr.FicheSanitaire
	Responsable    pr.Etatcivil
}

type ficheSanitaireTmplData struct {
	Pages []FicheSanitaire
	Asso  config.Asso
}

// CreateFicheSanitaires returns a PDF document, one "fiche sanitaire" per page.
func CreateFicheSanitaires(cfg config.Asso, pages []FicheSanitaire) ([]byte, error) {
	return templateToPDF(ficheSanitaireTmpl, ficheSanitaireTmplData{
		Pages: pages,
		Asso:  cfg,
	})
}

type Participant struct {
	Participant string
	Responsable string
	Mail        string
	Commune     string
}

type listeParticipantsTmplData struct {
	Asso         config.Asso
	Camp         string // label
	Participants []Participant
}

// CreateListeParticipants returns a PDF document.
func CreateListeParticipants(cfg config.Asso, participants []Participant, camp string) ([]byte, error) {
	return templateToPDF(listeParticipantsTmpl, listeParticipantsTmplData{
		Asso:         cfg,
		Camp:         camp,
		Participants: participants,
	})
}

type listeVetementsTmplData struct {
	Asso       config.Asso
	Camp       string // label
	Vetements  []camps.Vetement
	Complement template.HTML
}

// CreateListeVetements returns a PDF document.
func CreateListeVetements(cfg config.Asso, liste camps.ListeVetements, camp string) ([]byte, error) {
	return templateToPDF(listeVetementsTmpl, listeVetementsTmplData{
		Asso:       cfg,
		Camp:       camp,
		Vetements:  liste.Vetements,
		Complement: template.HTML(liste.Complement),
	})
}

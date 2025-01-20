package pdfcreator

import (
	"embed"
	"html/template"

	"registro/asso"
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
	Asso  asso.AssoMeta
}

// CreateFicheSanitaires returns a PDF document, one "fiche sanitaire" per page.
func CreateFicheSanitaires(pages []FicheSanitaire) ([]byte, error) {
	return templateToPDF(ficheSanitaireTmpl, ficheSanitaireTmplData{
		Pages: pages,
		Asso:  asso.Asso, // TODO:
	})
}

type Participant struct {
	Participant string
	Responsable string
	Mail        string
	Commune     string
}

type listeParticipantsTmplData struct {
	Asso         asso.AssoMeta
	Camp         string // label
	Participants []Participant
}

// CreateListeParticipants returns a PDF document.
func CreateListeParticipants(participants []Participant, camp string) ([]byte, error) {
	return templateToPDF(listeParticipantsTmpl, listeParticipantsTmplData{
		Asso:         asso.Asso, // TODO:
		Camp:         camp,
		Participants: participants,
	})
}

type listeVetementsTmplData struct {
	Asso       asso.AssoMeta
	Camp       string // label
	Vetements  []camps.Vetement
	Complement template.HTML
}

// CreateListeVetements returns a PDF document.
func CreateListeVetements(liste camps.ListeVetements, camp string) ([]byte, error) {
	return templateToPDF(listeVetementsTmpl, listeVetementsTmplData{
		Asso:       asso.Asso, // TODO:
		Camp:       camp,
		Vetements:  liste.Vetements,
		Complement: template.HTML(liste.Complement),
	})
}

package pdfcreator

import (
	"embed"
	"html/template"
	"time"

	"registro/config"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

//go:embed templates/*
var templates embed.FS

var (
	ficheSanitaireTmpl      *template.Template
	listeParticipantsTmpl   *template.Template
	listeVetementsTmpl      *template.Template
	attestationPresenceTmpl *template.Template
)

func init() {
	ficheSanitaireTmpl = parseTemplate("templates/fiche_sanitaire.html")
	listeParticipantsTmpl = parseTemplate("templates/liste_participants.html")
	listeVetementsTmpl = parseTemplate("templates/liste_vetements.html")
	attestationPresenceTmpl = parseTemplate("templates/attestation_presence.html")
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
	FicheSanitaire pr.Fichesanitaire
	Responsable    pr.Etatcivil
}

// CreateFicheSanitaires returns a PDF document, one "fiche sanitaire" per page.
func CreateFicheSanitaires(cfg config.Asso, pages []FicheSanitaire) ([]byte, error) {
	type ficheSanitaireTmplData struct {
		Asso  config.Asso
		Pages []FicheSanitaire
	}

	return templateToPDF(ficheSanitaireTmpl, ficheSanitaireTmplData{
		Pages: pages,
		Asso:  cfg,
	})
}

type Participant struct {
	NomPrenom   string
	Responsable string
	Mail        string
	Commune     string
}

// CreateListeParticipants returns a PDF document.
func CreateListeParticipants(cfg config.Asso, participants []Participant, camp string) ([]byte, error) {
	type listeParticipantsTmplData struct {
		Asso         config.Asso
		Camp         string // label
		Participants []Participant
	}

	return templateToPDF(listeParticipantsTmpl, listeParticipantsTmplData{
		Asso:         cfg,
		Camp:         camp,
		Participants: participants,
	})
}

// CreateListeVetements returns a PDF document.
func CreateListeVetements(cfg config.Asso, liste cps.ListeVetements, camp string) ([]byte, error) {
	type listeVetementsTmplData struct {
		Asso       config.Asso
		Camp       string // label
		Vetements  []cps.Vetement
		Complement template.HTML
	}

	return templateToPDF(listeVetementsTmpl, listeVetementsTmplData{
		Asso:       cfg,
		Camp:       camp,
		Vetements:  liste.Vetements,
		Complement: template.HTML(liste.Complement),
	})
}

type Destinataire struct {
	NomPrenom  string
	Adresse    string
	CodePostal string
	Ville      string
}

// CreateAttestationPresence returns a PDF document.
func CreateAttestationPresence(cfg config.Asso, destinataire Destinataire, participants []dossiers.ParticipantExt) ([]byte, error) {
	type attestationPresenceTmplData struct {
		Asso         config.Asso
		Date         string // now
		Destinataire Destinataire
		Participants []dossiers.ParticipantExt
	}

	return templateToPDF(attestationPresenceTmpl, attestationPresenceTmplData{
		Asso:         cfg,
		Date:         shared.NewDateFrom(time.Now()).String(),
		Destinataire: destinataire,
		Participants: participants,
	})
}

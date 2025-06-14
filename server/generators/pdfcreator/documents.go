package pdfcreator

import (
	"embed"
	"html/template"
	"time"

	"registro/config"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

//go:embed templates/*
var templates embed.FS

var (
	ficheSanitaireTmpl      *template.Template
	listeParticipantsTmpl   *template.Template
	listeVetementsTmpl      *template.Template
	lettreDirecteurTmpl     *template.Template
	attestationPresenceTmpl *template.Template
	factureTmpl             *template.Template
)

func init() {
	ficheSanitaireTmpl = parseTemplate("templates/ficheSanitaire.html")
	listeParticipantsTmpl = parseTemplate("templates/listeParticipants.html")
	listeVetementsTmpl = parseTemplate("templates/listeVetements.html")
	lettreDirecteurTmpl = parseTemplate("templates/lettreDirecteur.html")
	attestationPresenceTmpl = parseTemplate("templates/attestationPresence.html")
	factureTmpl = parseTemplate("templates/facture.html")
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
func CreateAttestationPresence(cfg config.Asso, destinataire Destinataire, participants []cps.ParticipantCamp) ([]byte, error) {
	args := struct {
		Asso         config.Asso
		Date         string // now
		Destinataire Destinataire
		Participants []cps.ParticipantCamp
	}{
		Asso:         cfg,
		Date:         shared.NewDateFrom(time.Now()).String(),
		Destinataire: destinataire,
		Participants: participants,
	}

	return templateToPDF(attestationPresenceTmpl, args)
}

// CreateFacture returns a PDF document.
func CreateFacture(cfg config.Asso, destinataire Destinataire, participants []cps.ParticipantCamp, finances logic.BilanFinancesPub, paiements []ds.Paiement) ([]byte, error) {
	type participantFinance struct {
		cps.ParticipantCamp
		Finances logic.BilanParticipantPub
	}
	pList := make([]participantFinance, len(participants))
	for i, p := range participants {
		pList[i] = participantFinance{p, finances.Inscrits[p.Participant.Id]}
	}
	args := struct {
		Asso         config.Asso
		Date         string // now
		Destinataire Destinataire
		Participants []participantFinance
		Finances     logic.BilanFinancesPub
		Paiements    []ds.Paiement
		IsAcquitte   bool
	}{
		Asso:         cfg,
		Date:         shared.NewDateFrom(time.Now()).String(),
		Destinataire: destinataire,
		Participants: pList,
		Finances:     finances,
		Paiements:    paiements,
		IsAcquitte:   finances.Statut == logic.Complet,
	}

	return templateToPDF(factureTmpl, args)
}

func ensureHexColor(hexa string) string {
	if len(hexa) > 7 {
		return hexa[0:7]
	}
	return hexa
}

type directeurCoords struct {
	NomPrenom  string
	Adresse    string
	CodePostal string
	Ville      string
	Mail       string
	Tels       template.HTML
}

// CreateLettreDirecteur returns a PDF document.
func CreateLettreDirecteur(cfg config.Asso, lettre cps.Lettredirecteur, directeur pr.Etatcivil) ([]byte, error) {
	args := struct {
		Asso                config.Asso
		ExpediteurDirecteur bool
		ShowAdressePostale  bool
		ColorCoord          string
		LettreHtml          template.HTML
		Directeur           directeurCoords
	}{
		Asso:                cfg,
		ExpediteurDirecteur: !lettre.UseCoordCentre,
		ShowAdressePostale:  lettre.ShowAdressePostale,
		// weasyprint does not support AHex colors
		ColorCoord: ensureHexColor(lettre.ColorCoord),
		Directeur: directeurCoords{
			NomPrenom:  directeur.PrenomNOM(),
			Adresse:    directeur.Adresse,
			CodePostal: directeur.CodePostal,
			Ville:      directeur.Ville,
			Mail:       directeur.Mail,
			Tels:       template.HTML(directeur.Tels.StringHTML()),
		},
		LettreHtml: template.HTML(lettre.Html),
	}
	return templateToPDF(lettreDirecteurTmpl, args)
}

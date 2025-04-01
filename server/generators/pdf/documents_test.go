package pdfcreator

import (
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"registro/config"
	"registro/sql/camps"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
	tu "registro/utils/testutils"
)

var cfg config.Asso

func init() {
	err := Init(os.TempDir(), "../..")
	if err != nil {
		panic(err)
	}

	os.Setenv("ASSO", "acve")
	os.Setenv("ASSO_BANK_IBAN", "iban1,iban2")
	cfg, err = config.NewAsso()
	if err != nil {
		panic(err)
	}
}

func randBool() bool {
	i := rand.Intn(2)
	return i == 0
}

func randFicheSanitaire() FicheSanitaire {
	al := pr.Allergies{
		Alimentaires:    randBool(),
		Asthme:          randBool(),
		Medicamenteuses: randBool(),
		Autres:          utils.RandString(200, true),
		ConduiteATenir:  "mzkmk \n slkdj lksjd lsdààéàéàsdmskd \n sdsdlk",
	}
	ml := pr.Maladies{
		Angine:     randBool(),
		Coqueluche: randBool(),
		Oreillons:  randBool(),
		Otite:      randBool(),
		Rhumatisme: randBool(),
		Rougeole:   randBool(),
	}
	fs := pr.Fichesanitaire{
		TraitementMedical: randBool(),
		DifficultesSante:  utils.RandString(100, true),
		Handicap:          randBool(),
		Recommandations:   utils.RandString(300, true),
		Tel:               "546565654646",
		Medecin: pr.Medecin{
			Nom: utils.RandString(30, true),
			Tel: pr.Tel(utils.RandString(30, true)),
		},
	}
	if randBool() {
		fs.Allergies = al
	}
	if randBool() {
		fs.Maladies = ml
	}
	pers := pr.Etatcivil{
		Nom:           utils.RandString(15, true),
		Prenom:        "zkle é@dzkmk è",
		Sexe:          pr.Woman,
		DateNaissance: shared.NewDateFrom(time.Now()),
	}
	resp := pr.Etatcivil{
		Nom:             utils.RandString(25, true),
		Prenom:          utils.RandString(25, true),
		SecuriteSociale: utils.RandString(25, true),
		Adresse:         "lskkd \n lsmdksmd smdl",
		CodePostal:      utils.RandString(5, true),
		Ville:           utils.RandString(15, true),
		Pays:            pr.Pays(utils.RandString(2, false)),
		Tels:            []string{"7987987979", "897779897998789"},
	}
	return FicheSanitaire{pers, fs, resp}
}

func randFicheSanitaires() []FicheSanitaire {
	out := make([]FicheSanitaire, 2)
	for i := range out {
		out[i] = randFicheSanitaire()
	}
	return out
}

func TestFicheSanitaire(t *testing.T) {
	data := randFicheSanitaires()

	ti := time.Now()
	content, err := CreateFicheSanitaires(cfg, data)
	fmt.Println(time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "FicheSanitaires.pdf", content)
}

func BenchmarkFS(b *testing.B) {
	pages := randFicheSanitaires()
	for i := 0; i < b.N; i++ {
		_, _ = CreateFicheSanitaires(cfg, pages)
	}
}

func randParticipants() []Participant {
	out := make([]Participant, 50)
	for i := range out {
		out[i] = Participant{
			utils.RandString(50, true),
			utils.RandString(50, true),
			utils.RandString(50, true),
			utils.RandString(50, true),
		}
	}
	return out
}

func TestListeParticipants(t *testing.T) {
	data := randParticipants()

	ti := time.Now()
	content, err := CreateListeParticipants(cfg, data, "VIvie la vie 2024")
	fmt.Println(time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeParticipants.pdf", content)
}

func randVetements() []camps.Vetement {
	out := make([]camps.Vetement, 50)
	for i := range out {
		out[i] = camps.Vetement{
			Quantite:    rand.Int(),
			Description: utils.RandString(50, true),
			Important:   randBool(),
		}
	}
	return out
}

func TestListeVetements(t *testing.T) {
	data := randVetements()

	ti := time.Now()
	content, err := CreateListeVetements(cfg, camps.ListeVetements{
		Vetements:  data,
		Complement: "Il n'ya <b> pas</b> de service de lingerie.",
	}, "VIvie la vie 2024")
	fmt.Println(time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "ListeVetements.pdf", content)
}

func TestAttestationPresence(t *testing.T) {
	camp := camps.Camp{
		Nom:       "C2",
		DateDebut: shared.NewDate(2022, 5, 13),
		Duree:     30,
		Agrement:  "5465sd6s64s6d4",
	}
	personne := pr.Etatcivil{
		Nom: "Kugler", Prenom: "Benoit",
		Sexe: pr.Woman, DateNaissance: shared.NewDate(1999, 1, 3),
	}

	ti := time.Now()
	content, err := CreateAttestationPresence(cfg, Destinataire{
		NomPrenom:  "Kugler benoit",
		Adresse:    "200, Route de Dieulefit",
		CodePostal: "07568",
		Ville:      "Montélimar",
	}, []camps.ParticipantCamp{
		{Camp: camp, ParticipantPersonne: camps.ParticipantPersonne{Personne: pr.Personne{Etatcivil: personne}}},
		{Camp: camp, ParticipantPersonne: camps.ParticipantPersonne{Personne: pr.Personne{Etatcivil: personne}}},
		{Camp: camp, ParticipantPersonne: camps.ParticipantPersonne{Personne: pr.Personne{Etatcivil: personne}}},
		{Camp: camp, ParticipantPersonne: camps.ParticipantPersonne{Personne: pr.Personne{Etatcivil: personne}}},
	})
	fmt.Println(time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "AttestationPresence.pdf", content)
}

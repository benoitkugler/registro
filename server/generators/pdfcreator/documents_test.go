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
	err := Init(os.TempDir(), "../../assets")
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

func TestEmptyFicheSanitair(t *testing.T) {
	data := []FicheSanitaire{
		{},
	}
	ti := time.Now()
	content, err := CreateFicheSanitaires(cfg, data)
	fmt.Println(time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "FicheSanitairesEmpty.pdf", content)
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

const lettre1 = `
<p>Chers parents,</p>
<p> </p>
<p>voici quelques informations complémentaires.</p>
<p>Le camp aura lieu chez André et Myriam BARBE, 25 rue de la Pierre Merlière à la Motte d&#39;Aveillans. Ce village est situé à 900 mètres d&#39;altitude...nous serons donc au frais !</p>
<p>Nous logerons sous tente et vivrons quasiment tout le temps en plein air. Merci donc de prévoir des vêtements en conséquence !</p>
<p>Outre les activités grand jeu et veillées habituels, nous avons prévu 2 jours de rando avec nuit en refuge ou à la belle étoile, ainsi qu&#39;une activité catamaran et peut-être kayak. Pour ces deux activités, nous avons impérativement besoin du teste d&#39;aisance aquatique que vous trouverez en pièce jointe. Si vous l&#39;avez déjà fourni les années précédentes pour un séjour ACVE, inutile de le renvoyer.</p>
<p>Pensez également à remplir la fiche sanitaire...</p>
<p>Lors de transports en bus et des sorties extérieures, des masques grand public seront fournis aux ados. inutile de venir avec (sauf si votre ado voyage en train pour venir au camp...).</p>
<p>Vous trouverez également en pièce jointe une liste de vêtements et de matériel. Cette liste est non exhaustive mais la place dans les tentes sera réduite (car les bagages devront être rangés dans la tente où dormiront les ados afin d&#39;éviter tout échange de vêtement ou de matériel...) donc ne pas prendre de trop gros bagages... Aucun service de lingerie ne fonctionnera mais les ados pourront laver leurs vêtements à la main si nécessaire...</p>
<p>Nous allons créer un groupe WhatsApp pour donner facilement des nouvelles quotidiennes durant le camp. merci de m&#39;indiquer le numéro de téléphone à utiliser.</p>
<p>Les téléphones des ados seront mis en consigne au début du séjour et ne seront donnés qu&#39;à certains moments : merci de bien prévenir vos ados de ce point !</p>
<p>Je reste disponible en cas de questions !</p>
<p>A très bientôt !</p>
<p>Karine</p>
<p> </p>
<p><span style="font-size: 14pt; color: #3598db">ESPACE PERSONNEL</span>  </p>
<p>Lors de l&#39;inscription de votre enfant, un espace personnel <em>Parents</em> vous a été attribué et un lien vers celui-ci envoyé dans le mail de confirmation (<span style="background-color: #ffffff">Mon</span> <span style="background-color: #ffffff">Dossier</span>). Dans cet espace vous trouverez :</p>
<ul>
<li>le <strong>suivi financier</strong> : vous pourrez alors joindre en ligne les aides auxquelles vous avez le droit (bons CAF, Comité d&#39;entreprise...), afin que le centre d&#39;inscription puisse les déduire de la facture finale.</li>
<li>les <strong>documents liés au séjour</strong> : liste de vêtement, lettre aux parents, plan d’accès au site ....</li>
<li>les <strong>documents à compléter</strong> ou joindre en ligne : test d&#39;aisance aquatique si besoin ...</li>
<li>l’accès à l&#39;<strong>album photo</strong> du séjour</li>
<li>la <strong>fiche sanitaire</strong> à compléter en ligne avec les allergies alimentaires.</li>
</ul>
<p><em>TOUTES LES INFOS ET DOCUMENTS DU SÉJOUR SE TROUVENT DANS VOTRE ESPACE DÉDIÉ !</em></p>`

func TestLettreDirecteur(t *testing.T) {
	ti := time.Now()
	content, err := CreateLettreDirecteur(cfg, camps.Lettredirecteur{
		// UseCoordCentre: ,
		Html:               lettre1,
		ShowAdressePostale: true,
		ColorCoord:         "#FF12A1",
	}, pr.Etatcivil{
		Nom:        "Kugler",
		Prenom:     " benoit",
		Mail:       "ummy@free.fr",
		Adresse:    "200, Route de Dieulefit",
		CodePostal: "07568",
		Ville:      "Montélimar",
	})
	fmt.Println("Generated in", time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "LettreDirecteur_1.pdf", content)

	ti = time.Now()
	content, err = CreateLettreDirecteur(cfg, camps.Lettredirecteur{
		UseCoordCentre:     true,
		Html:               lettre1,
		ShowAdressePostale: true,
	}, pr.Etatcivil{})
	fmt.Println("Generated in", time.Since(ti))
	tu.AssertNoErr(t, err)
	tu.Write(t, "LettreDirecteur_2.pdf", content)
}

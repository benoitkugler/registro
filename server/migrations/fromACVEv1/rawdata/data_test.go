package rawdata

import (
	"math/rand"
	"time"

	"github.com/lib/pq"
)

func randint64() int64 {
	return int64(rand.Intn(1000000))
}

func randfloat64() float64 {
	return rand.Float64() * float64(rand.Int31())
}

func randEuros() Euros {
	return Euros(randfloat64())
}

func randbool() bool {
	i := rand.Int31n(2)
	return i == 1
}

func randBool() Bool {
	return Bool(randbool())
}

func randint() int {
	return int(rand.Intn(1000000))
}

func randInt() Int {
	return Int(randint())
}

func randAide() Aide {
	return Aide{
		Id:              randint64(),
		IdStructureaide: randint64(),
		IdParticipant:   randint64(),
		Valeur:          randEuros(),
		Valide:          randBool(),
		ParJour:         randBool(),
		NbJoursMax:      randInt(),
	}
}

var letterRunes2 = []rune("azertyuiopqsdfghjklmwxcvbn123456789é@!?&èïab ")

func randstring() string {
	b := make([]rune, 50)
	maxLength := len(letterRunes2)
	for i := range b {
		b[i] = letterRunes2[rand.Intn(maxLength)]
	}
	return string(b)
}

func randString() String {
	return String(randstring())
}

func randTrajetBus() TrajetBus {
	return TrajetBus{
		RendezVous: randString(),
		Prix:       randEuros(),
	}
}

func randBusCamp() BusCamp {
	return BusCamp{
		Actif:       randbool(),
		Commentaire: randString(),
		Aller:       randTrajetBus(),
		Retour:      randTrajetBus(),
	}
}

func randMaterielSkiCamp() MaterielSkiCamp {
	return MaterielSkiCamp{
		Actif:      randbool(),
		PrixAcve:   randEuros(),
		PrixLoueur: randEuros(),
	}
}

func randOptionsCamp() OptionsCamp {
	return OptionsCamp{
		Bus:         randBusCamp(),
		MaterielSki: randMaterielSkiCamp(),
	}
}

func randtTime() time.Time {
	return time.Unix(int64(rand.Int31()), 5)
}

func randDate() Date {
	return Date(randtTime())
}

func randVetement() Vetement {
	return Vetement{
		Quantite:    randint(),
		Description: randstring(),
		Obligatoire: randbool(),
	}
}

func randSliceVetement() []Vetement {
	l := rand.Intn(10)
	out := make([]Vetement, l)
	for i := range out {
		out[i] = randVetement()
	}
	return out
}

func randListeVetements() ListeVetements {
	return ListeVetements{
		Liste:      randSliceVetement(),
		Complement: randstring(),
	}
}

func randSchemaPaiement() SchemaPaiement {
	choix := [...]SchemaPaiement{SPAcompte, SPTotal}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randEnvois() Envois {
	return Envois{
		Locked:            randbool(),
		LettreDirecteur:   randbool(),
		ListeVetements:    randbool(),
		ListeParticipants: randbool(),
	}
}

func randPlage() Plage {
	return Plage{
		From: randDate(),
		To:   randDate(),
	}
}

func randOptionSemaineCamp() OptionSemaineCamp {
	return OptionSemaineCamp{
		Plage1: randPlage(),
		Plage2: randPlage(),
		Prix1:  randEuros(),
		Prix2:  randEuros(),
	}
}

func randPrixParStatut() PrixParStatut {
	return PrixParStatut{
		Id:          randint64(),
		Prix:        randEuros(),
		Statut:      randString(),
		Description: randString(),
	}
}

func randSlicePrixParStatut() []PrixParStatut {
	l := rand.Intn(10)
	out := make([]PrixParStatut, l)
	for i := range out {
		out[i] = randPrixParStatut()
	}
	return out
}

func randSliceEuros() []Euros {
	l := rand.Intn(10)
	out := make([]Euros, l)
	for i := range out {
		out[i] = randEuros()
	}
	return out
}

func randOptionPrixCamp() OptionPrixCamp {
	return OptionPrixCamp{
		Active:  randstring(),
		Semaine: randOptionSemaineCamp(),
		Statut:  randSlicePrixParStatut(),
		Jour:    randSliceEuros(),
	}
}

func randArray4float64() [4]float64 {
	var out [4]float64
	for i := range out {
		out[i] = randfloat64()
	}
	return out
}

func randOptionQuotientFamilial() OptionQuotientFamilial {
	return OptionQuotientFamilial(randArray4float64())
}

func randCamp() Camp {
	return Camp{
		Id:                randint64(),
		Lieu:              randString(),
		Nom:               randString(),
		Prix:              randEuros(),
		NbPlaces:          randInt(),
		Password:          randString(),
		Ouvert:            randBool(),
		NbPlacesReservees: randInt(),
		NumeroJS:          randString(),
		NeedEquilibreGf:   randBool(),
		AgeMin:            randInt(),
		AgeMax:            randInt(),
		Options:           randOptionsCamp(),
		DateDebut:         randDate(),
		DateFin:           randDate(),
		ListeVetements:    randListeVetements(),
		SchemaPaiement:    randSchemaPaiement(),
		JoomeoAlbumId:     randString(),
		Envois:            randEnvois(),
		LienCompta:        randString(),
		OptionPrix:        randOptionPrixCamp(),
		InscriptionSimple: randBool(),
		Infos:             randString(),
		QuotientFamilial:  randOptionQuotientFamilial(),
	}
}

func randCampContrainte() CampContrainte {
	return CampContrainte{
		IdCamp:       randint64(),
		IdContrainte: randint64(),
	}
}

func randuint8() uint8 {
	return uint8(rand.Intn(1000000))
}

func randSliceuint8() []byte {
	l := rand.Intn(10)
	out := make([]byte, l)
	for i := range out {
		out[i] = randuint8()
	}
	return out
}

func randContenuDocument() ContenuDocument {
	return ContenuDocument{
		IdDocument: randint64(),
		Contenu:    randSliceuint8(),
		Miniature:  randSliceuint8(),
	}
}

func randOptionnalId() OptionnalId {
	return OptionnalId{
		Int64: randint64(),
		Valid: randbool(),
	}
}

func randBuiltinContrainte() BuiltinContrainte {
	choix := [...]BuiltinContrainte{CAutre, CBafa, CBafaEquiv, CBafd, CBafdEquiv, CCarteId, CCarteVitale, CCertMedCuisine, CHaccp, CInvalide, CPermis, CSb, CScolarite, CSecour, CTestNautique, CVaccin}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randContrainte() Contrainte {
	return Contrainte{
		Id:          randint64(),
		IdPersonne:  randOptionnalId(),
		IdDocument:  randOptionnalId(),
		Builtin:     randBuiltinContrainte(),
		Nom:         randString(),
		Description: randString(),
		MaxDocs:     randint(),
		JoursValide: randint(),
	}
}

func randTaille() Taille {
	return Taille(randint())
}

func randTime() Time {
	return Time(randtTime())
}

func randDocument() Document {
	return Document{
		Id:             randint64(),
		Taille:         randTaille(),
		NomClient:      randString(),
		Description:    randString(),
		DateHeureModif: randTime(),
	}
}

func randDocumentAide() DocumentAide {
	return DocumentAide{
		IdDocument: randint64(),
		IdAide:     randint64(),
	}
}

func randDocumentCamp() DocumentCamp {
	return DocumentCamp{
		IdDocument: randint64(),
		IdCamp:     randint64(),
		IsLettre:   randbool(),
	}
}

func randDocumentPersonne() DocumentPersonne {
	return DocumentPersonne{
		IdDocument:   randint64(),
		IdPersonne:   randint64(),
		IdContrainte: randint64(),
	}
}

func randModePaiment() ModePaiment {
	choix := [...]ModePaiment{MPAncv, MPAucun, MPCarte, MPCheque, MPEspece, MPHelloasso, MPVirement}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randInfoDon() InfoDon {
	return InfoDon{
		IdPaiementHelloAsso: randstring(),
	}
}

func randDon() Don {
	return Don{
		Id:            randint64(),
		Valeur:        randEuros(),
		ModePaiement:  randModePaiment(),
		DateReception: randDate(),
		RecuEmis:      randDate(),
		Infos:         randInfoDon(),
		Remercie:      randBool(),
		Details:       randString(),
		Affectation:   randString(),
	}
}

func randDonDonateur() DonDonateur {
	return DonDonateur{
		IdDon:       randint64(),
		IdPersonne:  randOptionnalId(),
		IdOrganisme: randOptionnalId(),
	}
}

func randRole() Role {
	choix := [...]Role{RAdjoint, RAideAnimation, RAnimation, RAutre, RBabysiter, RChauffeur, RCuis, RDirecteur, RFactotum, RInfirm, RIntend, RLing, RMen}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randSliceRole() []Role {
	l := rand.Intn(10)
	out := make([]Role, l)
	for i := range out {
		out[i] = randRole()
	}
	return out
}

func randRoles() Roles {
	return Roles(randSliceRole())
}

func randDiplome() Diplome {
	choix := [...]Diplome{DAgreg, DAssSociale, DAucun, DBafa, DBafaStag, DBafd, DBafdStag, DBapaat, DBeatep, DBjeps, DCap, DDeug, DDut, DEducSpe, DEje, DInstit, DMonEduc, DProf, DStaps, DZzautre}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randApprofondissement() Approfondissement {
	choix := [...]Approfondissement{AAucun, AAutre, ACanoe, AMoto, ASb, AVoile}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randOptionnalPlage() OptionnalPlage {
	return OptionnalPlage{
		Plage:  randPlage(),
		Active: randbool(),
	}
}

func randInvitationEquipier() InvitationEquipier {
	return InvitationEquipier(randint())
}

func randOptionnalBool() OptionnalBool {
	choix := [...]OptionnalBool{OBNon, OBOui, OBPeutEtre}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randEquipier() Equipier {
	return Equipier{
		Id:                 randint64(),
		IdCamp:             randint64(),
		IdPersonne:         randint64(),
		Roles:              randRoles(),
		Diplome:            randDiplome(),
		Appro:              randApprofondissement(),
		Presence:           randOptionnalPlage(),
		InvitationEquipier: randInvitationEquipier(),
		Charte:             randOptionnalBool(),
	}
}

func randEquipierContrainte() EquipierContrainte {
	return EquipierContrainte{
		IdEquipier:   randint64(),
		IdContrainte: randint64(),
		Optionnel:    randbool(),
	}
}

func randSexe() Sexe {
	choix := [...]Sexe{SAucun, SFemme, SHomme}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randDestinataire() Destinataire {
	return Destinataire{
		NomPrenom:  randString(),
		Sexe:       randSexe(),
		Adresse:    randString(),
		CodePostal: randString(),
		Ville:      randString(),
	}
}

func randSliceDestinataire() []Destinataire {
	l := rand.Intn(10)
	out := make([]Destinataire, l)
	for i := range out {
		out[i] = randDestinataire()
	}
	return out
}

func randDestinatairesOptionnels() DestinatairesOptionnels {
	return DestinatairesOptionnels(randSliceDestinataire())
}

func randSlicestring() []string {
	l := rand.Intn(10)
	out := make([]string, l)
	for i := range out {
		out[i] = randstring()
	}
	return out
}

func randStringArray() pq.StringArray {
	return pq.StringArray(randSlicestring())
}

func randFacture() Facture {
	return Facture{
		Id:                      randint64(),
		IdPersonne:              randint64(),
		DestinatairesOptionnels: randDestinatairesOptionnels(),
		Key:                     randString(),
		CopiesMails:             randStringArray(),
		LastConnection:          randtTime(),
		IsConfirmed:             randbool(),
		IsValidated:             randbool(),
		PartageAdressesOK:       randbool(),
	}
}

func randGroupe() Groupe {
	return Groupe{
		Id:       randint64(),
		IdCamp:   randint64(),
		Nom:      randString(),
		Plage:    randPlage(),
		Couleur:  randstring(),
		isSimple: randbool(),
	}
}

func randGroupeContrainte() GroupeContrainte {
	return GroupeContrainte{
		IdGroupe:     randint64(),
		IdContrainte: randint64(),
	}
}

func randGroupeEquipier() GroupeEquipier {
	return GroupeEquipier{
		IdGroupe:   randint64(),
		IdEquipier: randint64(),
		IdCamp:     randint64(),
	}
}

func randGroupeParticipant() GroupeParticipant {
	return GroupeParticipant{
		IdParticipant: randint64(),
		IdGroupe:      randint64(),
		IdCamp:        randint64(),
		Manuel:        randbool(),
	}
}

func randImageuploaded() Imageuploaded {
	return Imageuploaded{
		IdCamp:   randint64(),
		Filename: randstring(),
		Lien:     randstring(),
		Content:  randSliceuint8(),
	}
}

func randIdentificationId() IdentificationId {
	return IdentificationId{
		Valid:   randbool(),
		Id:      randint64(),
		Crypted: randstring(),
	}
}

func randTels() Tels {
	return Tels(randSlicestring())
}

func randPays() Pays {
	return Pays(randstring())
}

func randResponsableLegal() ResponsableLegal {
	return ResponsableLegal{
		Lienid:        randIdentificationId(),
		Nom:           randString(),
		Prenom:        randString(),
		Sexe:          randSexe(),
		Mail:          randString(),
		Adresse:       randString(),
		CodePostal:    randString(),
		Ville:         randString(),
		Tels:          randTels(),
		DateNaissance: randDate(),
		Pays:          randPays(),
	}
}

func randBus() Bus {
	choix := [...]Bus{BAller, BAllerRetour, BAucun, BRetour}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randMaterielSki() MaterielSki {
	return MaterielSki{
		Need:     randstring(),
		Mode:     randstring(),
		Casque:   randbool(),
		Poids:    randint(),
		Taille:   randint(),
		Pointure: randint(),
	}
}

func randOptionsParticipant() OptionsParticipant {
	return OptionsParticipant{
		Bus:         randBus(),
		MaterielSki: randMaterielSki(),
	}
}

func randSemaine() Semaine {
	choix := [...]Semaine{SComplet, SSe1, SSe2}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randSliceint() []int {
	l := rand.Intn(10)
	out := make([]int, l)
	for i := range out {
		out[i] = randint()
	}
	return out
}

func randJours() Jours {
	return Jours(randSliceint())
}

func randOptionPrixParticipant() OptionPrixParticipant {
	return OptionPrixParticipant{
		Semaine: randSemaine(),
		Statut:  randint64(),
		Jour:    randJours(),
	}
}

func randParticipantInscription() ParticipantInscription {
	return ParticipantInscription{
		Lienid:           randIdentificationId(),
		Nom:              randString(),
		Prenom:           randString(),
		DateNaissance:    randDate(),
		Sexe:             randSexe(),
		IdCamp:           randint64(),
		Options:          randOptionsParticipant(),
		OptionsPrix:      randOptionPrixParticipant(),
		QuotientFamilial: randInt(),
	}
}

func randSliceParticipantInscription() []ParticipantInscription {
	l := rand.Intn(10)
	out := make([]ParticipantInscription, l)
	for i := range out {
		out[i] = randParticipantInscription()
	}
	return out
}

func randParticipantInscriptions() ParticipantInscriptions {
	return ParticipantInscriptions(randSliceParticipantInscription())
}

func randInscription() Inscription {
	return Inscription{
		Id:                randint64(),
		Info:              randString(),
		DateHeure:         randTime(),
		CopiesMails:       randStringArray(),
		Responsable:       randResponsableLegal(),
		Participants:      randParticipantInscriptions(),
		PartageAdressesOK: randbool(),
	}
}

func randLettredirecteur() Lettredirecteur {
	return Lettredirecteur{
		IdCamp:             randint64(),
		Html:               randstring(),
		UseCoordCentre:     randbool(),
		ShowAdressePostale: randbool(),
		ColorCoord:         randstring(),
	}
}

func randMessageKind() MessageKind {
	choix := [...]MessageKind{MAccuseReception, MAttestationPresence, MCentre, MDocuments, MFacture, MFactureAcquittee, MInscription, MPaiement, MPlaceLiberee, MResponsable, MSondage, MSupprime}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randMessage() Message {
	return Message{
		Id:        randint64(),
		IdFacture: randint64(),
		Kind:      randMessageKind(),
		Created:   randTime(),
		Modified:  randTime(),
		Vu:        randbool(),
	}
}

func randDistribution() Distribution {
	choix := [...]Distribution{DEspacePerso, DMail, DMailAndDownload}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randMessageAttestation() MessageAttestation {
	return MessageAttestation{
		IdMessage:    randint64(),
		Distribution: randDistribution(),
		GuardKind:    randMessageKind(),
	}
}

func randMessageDocument() MessageDocument {
	return MessageDocument{
		IdMessage: randint64(),
		IdCamp:    randint64(),
		guardKind: randMessageKind(),
	}
}

func randMessageMessage() MessageMessage {
	return MessageMessage{
		IdMessage: randint64(),
		Contenu:   randString(),
		GuardKind: randMessageKind(),
	}
}

func randMessagePlacelibere() MessagePlacelibere {
	return MessagePlacelibere{
		IdMessage:     randint64(),
		IdParticipant: randint64(),
		guardKind:     randMessageKind(),
	}
}

func randMessageSondage() MessageSondage {
	return MessageSondage{
		IdMessage: randint64(),
		IdCamp:    randint64(),
		guardKind: randMessageKind(),
		isSimple:  randbool(),
	}
}

func randMessageView() MessageView {
	return MessageView{
		IdMessage: randint64(),
		IdCamp:    randint64(),
		Vu:        randbool(),
		guardKind: randMessageKind(),
	}
}

func randCoordonnees() Coordonnees {
	return Coordonnees{
		Tels:       randTels(),
		Mail:       randString(),
		Adresse:    randString(),
		CodePostal: randString(),
		Ville:      randString(),
		Pays:       randPays(),
	}
}

func randExemplaires() Exemplaires {
	return Exemplaires{
		PubEte:     randint(),
		PubHiver:   randint(),
		EchoRocher: randint(),
		EOnews:     randint(),
	}
}

func randOrganisme() Organisme {
	return Organisme{
		Id:            randint64(),
		Nom:           randString(),
		ContactPropre: randBool(),
		Contact:       randCoordonnees(),
		IdContact:     randOptionnalId(),
		IdContactDon:  randOptionnalId(),
		Exemplaires:   randExemplaires(),
	}
}

func randPaiement() Paiement {
	return Paiement{
		Id:              randint64(),
		IdFacture:       randint64(),
		IsAcompte:       randBool(),
		IsRemboursement: randBool(),
		InBordereau:     randTime(),
		LabelPayeur:     randString(),
		NomBanque:       randString(),
		ModePaiement:    randModePaiment(),
		Numero:          randString(),
		Valeur:          randEuros(),
		IsInvalide:      randBool(),
		DateReglement:   randTime(),
		Details:         randString(),
	}
}

func randStatutAttente() StatutAttente {
	choix := [...]StatutAttente{Attente, AttenteReponse, Inscrit, Refuse}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randListeAttente() ListeAttente {
	return ListeAttente{
		Statut: randStatutAttente(),
		Raison: randstring(),
	}
}

func randPourcent() Pourcent {
	return Pourcent(randint())
}

func randRemises() Remises {
	return Remises{
		ReducEquipiers: randPourcent(),
		ReducEnfants:   randPourcent(),
		ReducSpeciale:  randEuros(),
	}
}

func randParticipant() Participant {
	return Participant{
		Id:               randint64(),
		IdCamp:           randint64(),
		IdPersonne:       randint64(),
		IdFacture:        randOptionnalId(),
		ListeAttente:     randListeAttente(),
		Remises:          randRemises(),
		OptionPrix:       randOptionPrixParticipant(),
		Options:          randOptionsParticipant(),
		DateHeure:        randTime(),
		isSimple:         randbool(),
		QuotientFamilial: randInt(),
	}
}

func randParticipantEquipier() ParticipantEquipier {
	return ParticipantEquipier{
		IdParticipant: randint64(),
		IdEquipier:    randint64(),
		IdGroupe:      randint64(),
	}
}

func randParticipantsimple() Participantsimple {
	return Participantsimple{
		Id:         randint64(),
		IdPersonne: randint64(),
		IdCamp:     randint64(),
		DateHeure:  randTime(),
		Info:       randString(),
		isSimple:   randbool(),
	}
}

func randDepartement() Departement {
	return Departement(randstring())
}

func randBasePersonne() BasePersonne {
	return BasePersonne{
		Nom:                  randString(),
		NomJeuneFille:        randString(),
		Prenom:               randString(),
		DateNaissance:        randDate(),
		VilleNaissance:       randString(),
		DepartementNaissance: randDepartement(),
		Sexe:                 randSexe(),
		Tels:                 randTels(),
		Mail:                 randString(),
		Adresse:              randString(),
		CodePostal:           randString(),
		Ville:                randString(),
		Pays:                 randPays(),
		SecuriteSociale:      randString(),
		Profession:           randString(),
		Etudiant:             randBool(),
		Fonctionnaire:        randBool(),
	}
}

func randRangMembreAsso() RangMembreAsso {
	choix := [...]RangMembreAsso{RMABureau, RMACA, RMAMembre, RMANonMembre}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randSliceint64() []int64 {
	l := rand.Intn(10)
	out := make([]int64, l)
	for i := range out {
		out[i] = randint64()
	}
	return out
}

func randCotisation() Cotisation {
	return Cotisation(randSliceint64())
}

func randMaladies() Maladies {
	return Maladies{
		Rubeole:    randbool(),
		Varicelle:  randbool(),
		Angine:     randbool(),
		Oreillons:  randbool(),
		Scarlatine: randbool(),
		Coqueluche: randbool(),
		Otite:      randbool(),
		Rougeole:   randbool(),
		Rhumatisme: randbool(),
	}
}

func randAllergies() Allergies {
	return Allergies{
		Asthme:          randbool(),
		Alimentaires:    randbool(),
		Medicamenteuses: randbool(),
		Autres:          randstring(),
		ConduiteATenir:  randstring(),
	}
}

func randMedecin() Medecin {
	return Medecin{
		Nom: randstring(),
		Tel: randstring(),
	}
}

func randFicheSanitaire() FicheSanitaire {
	return FicheSanitaire{
		TraitementMedical: randbool(),
		Maladies:          randMaladies(),
		Allergies:         randAllergies(),
		DifficultesSante:  randstring(),
		Recommandations:   randstring(),
		Handicap:          randbool(),
		Tel:               randstring(),
		Medecin:           randMedecin(),
		LastModif:         randTime(),
		Mails:             randSlicestring(),
	}
}

func randPersonne() Personne {
	return Personne{
		Id:               randint64(),
		BasePersonne:     randBasePersonne(),
		VersionPapier:    randBool(),
		PubHiver:         randBool(),
		PubEte:           randBool(),
		EchoRocher:       randBool(),
		RangMembreAsso:   randRangMembreAsso(),
		QuotientFamilial: randInt(),
		Cotisation:       randCotisation(),
		Eonews:           randBool(),
		FicheSanitaire:   randFicheSanitaire(),
		IsTemporaire:     randBool(),
	}
}

func randSatisfaction() Satisfaction {
	choix := [...]Satisfaction{SDecevant, SMoyen, SSatisfaisant, STressatisfaisant, SVide}
	i := rand.Intn(len(choix))
	return choix[i]
}

func randRepSondage() RepSondage {
	return RepSondage{
		InfosAvantSejour:   randSatisfaction(),
		InfosPendantSejour: randSatisfaction(),
		Hebergement:        randSatisfaction(),
		Activites:          randSatisfaction(),
		Theme:              randSatisfaction(),
		Nourriture:         randSatisfaction(),
		Hygiene:            randSatisfaction(),
		Ambiance:           randSatisfaction(),
		Ressenti:           randSatisfaction(),
		MessageEnfant:      randString(),
		MessageResponsable: randString(),
	}
}

func randSondage() Sondage {
	return Sondage{
		Id:         randint64(),
		IdCamp:     randint64(),
		IdFacture:  randint64(),
		Modified:   randtTime(),
		RepSondage: randRepSondage(),
	}
}

func randStructureaide() Structureaide {
	return Structureaide{
		Id:              randint64(),
		Nom:             randString(),
		Immatriculation: randString(),
		Adresse:         randString(),
		CodePostal:      randString(),
		Ville:           randString(),
		Telephone:       randString(),
		Info:            randString(),
	}
}

func randModules() Modules {
	return Modules{
		Personnes:     randint(),
		Camps:         randint(),
		Inscriptions:  randint(),
		SuiviCamps:    randint(),
		SuiviDossiers: randint(),
		Paiements:     randint(),
		Aides:         randint(),
		Equipiers:     randint(),
		Dons:          randint(),
	}
}

func randUser() User {
	return User{
		Id:      randint64(),
		Label:   randString(),
		Mdp:     randString(),
		IsAdmin: randBool(),
		Modules: randModules(),
	}
}

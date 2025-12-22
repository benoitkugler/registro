package rawdata

// ce fichier défini des enums et les labels associés
// il sert de référence (unique) pour générer des types .ts et des maps .go

const (
	SAucun Sexe = ""  // -
	SHomme Sexe = "M" // Homme
	SFemme Sexe = "F" // Femme
)

const (
	RMANonMembre RangMembreAsso = ""  // Non membre
	RMAMembre    RangMembreAsso = "1" // Membre
	RMACA        RangMembreAsso = "2" // Membre du C.A.
	RMABureau    RangMembreAsso = "3" // Membre du bureau
)

const (
	MPAucun     ModePaiment = ""          // -
	MPVirement  ModePaiment = "vir"       // Virement
	MPCheque    ModePaiment = "cheque"    // Chèque
	MPEspece    ModePaiment = "esp"       // Espèces
	MPCarte     ModePaiment = "cb"        // Carte bancaire
	MPAncv      ModePaiment = "ancv"      // ANCV
	MPHelloasso ModePaiment = "helloasso" // Hello Asso
)

const (
	RDirecteur     Role = "_dir"       // Direction
	RAdjoint       Role = "_adjoint"   // Adjoint
	RAnimation     Role = "_anim"      // Animation
	RAideAnimation Role = "_aideanim"  // Aide-animateur
	RChauffeur     Role = "_chauffeur" // Chauffeur
	RIntend        Role = "_intend"    // Intendance
	RBabysiter     Role = "_babysiter" // Baby-sitter
	RMen           Role = "_men"       // Ménage
	RFactotum      Role = "_factotum"  // Factotum
	RInfirm        Role = "_infirm"    // Assistant sanitaire
	RCuis          Role = "_cuis"      // Cuisine
	RLing          Role = "_ling"      // Lingerie
	RAutre         Role = "_autre"     // Autre
)

const (
	DAucun      Diplome = ""            // Aucun
	DBafa       Diplome = "bafa"        // BAFA Titulaire
	DBafaStag   Diplome = "bafa_stag"   // BAFA Stagiaire
	DBafd       Diplome = "bafd"        // BAFD titulaire
	DBafdStag   Diplome = "bafd_stag"   // BAFD stagiaire
	DCap        Diplome = "cap"         // CAP petit enfance
	DAssSociale Diplome = "ass_sociale" // Assitante Sociale
	DEducSpe    Diplome = "educ_spe"    // Educ. spé.
	DMonEduc    Diplome = "mon_educ"    // Moniteur educateur
	DInstit     Diplome = "instit"      // Professeur des écoles
	DProf       Diplome = "prof"        // Enseignant du secondaire
	DAgreg      Diplome = "agreg"       // Agrégé
	DBjeps      Diplome = "bjeps"       // BPJEPS
	DDut        Diplome = "dut"         // DUT carrière sociale
	DEje        Diplome = "eje"         // EJE
	DDeug       Diplome = "deug"        // DEUG
	DStaps      Diplome = "staps"       // STAPS
	DBapaat     Diplome = "bapaat"      // BAPAAT
	DBeatep     Diplome = "beatep"      // BEATEP
	DZzautre    Diplome = "zzautre"     // AUTRE
)

const (
	AAucun Approfondissement = ""      // Non effectué
	AAutre Approfondissement = "autre" // Approfondissement
	ASb    Approfondissement = "sb"    // Surveillant de baignade
	ACanoe Approfondissement = "canoe" // Canoë - Kayak
	AVoile Approfondissement = "voile" // Voile
	AMoto  Approfondissement = "moto"  // Loisirs motocyclistes
)

const (
	SComplet Semaine = ""  // Camp complet
	SSe1     Semaine = "1" // Semaine 1
	SSe2     Semaine = "2" // Semaine 2
)

const (
	SPAcompte SchemaPaiement = "acompte" // Avec acompte
	SPTotal   SchemaPaiement = "total"   // Paiement direct (sans acompte)
)

const (
	OBPeutEtre OptionnalBool = 0  // Peut-être
	OBOui      OptionnalBool = 1  // Oui
	OBNon      OptionnalBool = -1 // Non
)

const (
	BAucun       Bus = ""             // -
	BAller       Bus = "aller"        // Aller
	BRetour      Bus = "retour"       // Retour
	BAllerRetour Bus = "aller_retour" // Aller-Retour
)

const (
	MSupprime MessageKind = iota // Message supprimé

	// expediteur : responsable
	MResponsable // Message

	// expediteur : centre d'inscription
	MCentre // Message du centre

	MAccuseReception     // Inscription validée
	MFacture             // Facture
	MDocuments           // Document des séjours
	MFactureAcquittee    // Facture acquittée
	MAttestationPresence // Attestation de présence
	MSondage             // Avis sur le séjour

	// enregistre le moment d'inscription
	MInscription // Moment d'inscription

	MPlaceLiberee // Place libérée

	// n'est jamais utilisé dans la base mais simplifie le frontend
	MPaiement //
)

// du plus favorable au moins favorable. L'ordre compte
// dans la fonction `HintsAttente.Hint`
const (
	Inscrit StatutAttente = iota // Inscrit
	Attente                      // Liste d'attente
	// une place s'est libérée et on attend une confirmation
	AttenteReponse // Attente de confirmation
	// On est (quasi) certain que le participant ne sera
	// pas pris, mais on préfère ne pas le supprimer du dossier
	Refuse // Refusé
)

const (
	Invalide     Completion = iota // -
	NonCommencee                   // En attente
	EnCours                        // En cours
	Complete                       // Complet
)

const (
	DEspacePerso     Distribution = iota // Téléchargée depuis l'espace de suivi
	DMail                                // Notifiée par courriel
	DMailAndDownload                     // Téléchargée après notification
)

// Attention, la valeur compte pour la présentation
// sur le frontend comme "form-rating"
const (
	SVide             Satisfaction = iota // -
	SDecevant                             // Décevant
	SMoyen                                // Moyen
	SSatisfaisant                         // Satisfaisant
	STressatisfaisant                     // Très satisfaisant
)

const (
	CInvalide       BuiltinContrainte = ""                 // -
	CBafa           BuiltinContrainte = "bafa"             // BAFA
	CBafd           BuiltinContrainte = "bafd"             // BAFD
	CCarteId        BuiltinContrainte = "carte_id"         // Carte d''identité/Passeport
	CPermis         BuiltinContrainte = "permis"           // Permis de conduire
	CSb             BuiltinContrainte = "sb"               // Surveillant de baignade
	CSecour         BuiltinContrainte = "secour"           // Secourisme (PSC1 - AFPS)
	CBafdEquiv      BuiltinContrainte = "bafd_equiv"       // Equivalent BAFD
	CBafaEquiv      BuiltinContrainte = "bafa_equiv"       // Equivalent BAFA
	CCarteVitale    BuiltinContrainte = "carte_vitale"     // Carte Vitale
	CHaccp          BuiltinContrainte = "haccp"            // Cuisine (HACCP)
	CCertMedCuisine BuiltinContrainte = "cert_med_cuisine" // Certificat médical Cuisine
	CScolarite      BuiltinContrainte = "scolarite"        // Certificat de scolarité
	CAutre          BuiltinContrainte = "autre"            // Autre
	CVaccin         BuiltinContrainte = "vaccin"           // Vaccin

	CTestNautique BuiltinContrainte = "test_nautique" // Test nautique
)

// Ce package implémentes les fondements de la gestion des données
// ACVE. Il est destiné à être utilisé sur le serveur et sur le client.
// Il définit les types de donneés utilisés
// et permet de changer d'une représentation à une autre (SQL/JSON/mémoire vive)
package rawdata

import (
	"time"

	"github.com/lib/pq"
)

// ----------------------------------- Personne ----------------------------------- //

// Personne représente les attributs d'une personne
type Personne struct {
	Id int64 `json:"id"`

	BasePersonne

	VersionPapier    Bool           `json:"version_papier"`
	PubHiver         Bool           `json:"pub_hiver"`
	PubEte           Bool           `json:"pub_ete"`
	EchoRocher       Bool           `json:"echo_rocher"`
	RangMembreAsso   RangMembreAsso `json:"rang_membre_asso"`
	QuotientFamilial Int            `json:"quotient_familial"`
	Cotisation       Cotisation     `json:"cotisation"` // dernière année
	Eonews           Bool           `json:"eonews"`
	FicheSanitaire   FicheSanitaire `json:"fiche_sanitaire"`

	// Si `true` indique un profil non vérifié. Les personnes temporaires
	// peuvent être supprimées à tout moment (nécessitant une redirection des tables
	// documents, participants,...)
	IsTemporaire Bool `json:"is_temporaire"`
}

// ----------------------------------- Camp ----------------------------------- //

// Camp représente un séjour proposé par l'ACVE
// requis participantsimples
// sql:ADD UNIQUE(id,inscription_simple)
type Camp struct {
	Id                int64       `json:"id"`
	Lieu              String      `json:"lieu"`
	Nom               String      `json:"nom"`
	Prix              Euros       `json:"prix"`
	NbPlaces          Int         `json:"nb_places"`
	Password          String      `json:"password"`
	Ouvert            Bool        `json:"ouvert"`
	NbPlacesReservees Int         `json:"nb_places_reservees"`
	NumeroJS          String      `json:"numero_js"`
	NeedEquilibreGf   Bool        `json:"need_equilibre_gf"`
	AgeMin            Int         `json:"age_min"`
	AgeMax            Int         `json:"age_max"`
	Options           OptionsCamp `json:"options"`
	DateDebut         Date        `json:"date_debut"`
	DateFin           Date        `json:"date_fin"`

	ListeVetements ListeVetements `json:"liste_vetements"`
	SchemaPaiement SchemaPaiement `json:"schema_paiement"`
	JoomeoAlbumId  String         `json:"joomeo_album_id"`
	Envois         Envois         `json:"envois"`
	LienCompta     String         `json:"lien_compta"`
	OptionPrix     OptionPrixCamp `json:"option_prix"`

	// Si oui, la procédure d'inscription est simplifiée.
	InscriptionSimple Bool `json:"inscription_simple"`

	Infos            String                 `json:"infos"`             // info additionnelle à afficher sur le formulaire d'inscription
	QuotientFamilial OptionQuotientFamilial `json:"quotient_familial"` // optionel
}

// LettreDirecteur conserve le html utilisé pour générer la lettre.
// En revanche, c'est bien le document stocké dans la base de données
// qui est utilisé ensuie.
// Le contenu html de la lettre peut contenir des images
// et alourdir significativement la table Camps.
// sql:ADD UNIQUE(id_camp)
type Lettredirecteur struct {
	IdCamp             int64  `json:"id_camp" sql_on_delete:"CASCADE"`
	Html               string `json:"html"`
	UseCoordCentre     bool   `json:"use_coord_centre"`
	ShowAdressePostale bool   `json:"show_adresse_postale"`
	ColorCoord         string `json:"color_coord"`
}

// Images contenues dans les lettres aux parents
// sql: ADD UNIQUE(id_camp, lien)
type Imageuploaded struct {
	IdCamp   int64  `json:"id_camp" sql_on_delete:"CASCADE"`
	Filename string `json:"filename"`
	Lien     string `json:"lien"`
	Content  []byte `json:"content"`
}

// Groupe représente un groupe de participants
// Un séjour peut définir (ou non) une liste de groupes
// sql:ADD UNIQUE(id_camp, nom)
// sql:ADD UNIQUE(id, id_camp)
// assure qu'un groupe est lié à un camp complet
// sql:ALTER COLUMN isSimple SET DEFAULT false
// sql:ADD CHECK(isSimple = false)
// sql:ADD FOREIGN KEY (id_camp, isSimple) REFERENCES camps(id,inscription_simple)
type Groupe struct {
	Id     int64 `json:"id"`
	IdCamp int64 `json:"id_camp" sql_on_delete:"CASCADE"`

	// un nom vide indique un groupe par défaut
	Nom String `json:"nom"`
	// indication: ignorée forcément pour un groupe par défaut
	Plage Plage `json:"plage"`
	// Hex color, optionnelle
	Couleur string `json:"couleur"`

	isSimple bool
}

// Sondage enregistre les retours sur un séjour
// sql:ADD UNIQUE(id_camp, id_facture)
type Sondage struct {
	Id        int64     `json:"id"`
	IdCamp    int64     `json:"id_camp" sql_on_delete:"CASCADE"`
	IdFacture int64     `json:"id_facture" sql_on_delete:"CASCADE"`
	Modified  time.Time `json:"modified"`

	RepSondage
}

// ------------------------------ Inscriptions ------------------------------

// Inscription enregistre l'inscription faite via le formulaire publique.
// L'inscription publique est transformée en facture dès réception,
// cette table ne sert donc qu'à garder une trace en cas de problème.
// Entre autre, l'intégrité des camps n'est pas assurée
type Inscription struct {
	Id          int64          `json:"id"`
	Info        String         `json:"info"`
	DateHeure   Time           `json:"date_heure"`
	CopiesMails pq.StringArray `json:"copies_mails"`

	Responsable  ResponsableLegal        `json:"responsable"`
	Participants ParticipantInscriptions `json:"participants"`

	PartageAdressesOK bool `json:"partage_adresses_ok"`
}

// -------------------------- Participants --------------------------

// Participant représente un inscrit sur un séjour complet (éventuellement sur liste d'attente)
// sql:ADD UNIQUE(id_personne, id_camp)
// nécessaire pour être référencé
// sql:ADD UNIQUE(id, id_camp)
// assure qu'un participant est lié à un camp complet
// sql:ALTER COLUMN isSimple SET DEFAULT false
// sql:ADD CHECK(isSimple = false)
// sql:ADD FOREIGN KEY (id_camp, isSimple) REFERENCES camps(id,inscription_simple)
type Participant struct {
	Id         int64       `json:"id"`
	IdCamp     int64       `json:"id_camp"`
	IdPersonne int64       `json:"id_personne" sql_on_delete:"CASCADE"`
	IdFacture  OptionnalId `json:"id_facture" sql_on_delete:"SET NULL"`

	ListeAttente ListeAttente          `json:"liste_attente"`
	Remises      Remises               `json:"remises"`
	OptionPrix   OptionPrixParticipant `json:"option_prix"`
	Options      OptionsParticipant    `json:"options"`

	// Moment d'inscription
	DateHeure Time `json:"date_heure"`

	isSimple bool

	QuotientFamilial Int `json:"quotient_familial"`
}

// GroupeParticipant défini le contenu des groupes
// sql:ADD UNIQUE (id_participant)
// sql:ADD UNIQUE (id_participant, id_groupe)
// sql:ADD FOREIGN KEY (id_participant, id_camp) REFERENCES participants(id,id_camp)
// sql:ADD FOREIGN KEY (id_groupe, id_camp) REFERENCES groupes(id,id_camp)
type GroupeParticipant struct {
	IdParticipant int64 `json:"id_participant" sql_on_delete:"CASCADE"`
	IdGroupe      int64 `json:"id_groupe" sql_on_delete:"CASCADE"`
	// redondance pour assurer l'intégrité
	IdCamp int64 `json:"id_camp"`
	// indique si l'attribuation a été faite
	// en modifiant directement la fiche du participant ou
	// en fonction de l'âge
	Manuel bool `json:"manuel"`
}

// Equipier représente un participant dans l'équipe d'un séjour
// sql:ADD UNIQUE(id_personne, id_camp)
// requise par la contrainte ParticipantEquipier
// sql:ADD UNIQUE(id, id_camp)
type Equipier struct {
	Id         int64 `json:"id"`
	IdCamp     int64 `json:"id_camp"`
	IdPersonne int64 `json:"id_personne" sql_on_delete:"CASCADE"`

	Roles              Roles              `json:"roles"`
	Diplome            Diplome            `json:"diplome"`
	Appro              Approfondissement  `json:"appro"`
	Presence           OptionnalPlage     `json:"presence"`
	InvitationEquipier InvitationEquipier `json:"invitation_equipier"`

	// validation de la charte ACVE
	Charte OptionnalBool `json:"charte"`
}

// GroupeEquipiers définit les animateurs d'un groupe
// sql:ADD UNIQUE(id_groupe, id_equipier)
// sql:ADD FOREIGN KEY (id_equipier, id_camp) REFERENCES equipiers(id,id_camp)
// sql:ADD FOREIGN KEY (id_groupe, id_camp) REFERENCES groupes(id,id_camp)
type GroupeEquipier struct {
	IdGroupe   int64 `json:"id_groupe" sql_on_delete:"CASCADE"`
	IdEquipier int64 `json:"id_equipier" sql_on_delete:"CASCADE"`
	// redondance pour assurer l'intégrité
	IdCamp int64 `json:"id_camp" sql_on_delete:"CASCADE"`
}

// ParticipantEquipier associe un animateur de référence à un inscrit
// sql:ADD UNIQUE (id_participant)
// sql:ADD FOREIGN KEY (id_participant, id_groupe) REFERENCES groupe_participants(id_participant, id_groupe)
// sql:ADD FOREIGN KEY (id_equipier, id_groupe) REFERENCES groupe_equipiers(id_equipier, id_groupe) ON DELETE CASCADE
type ParticipantEquipier struct {
	IdParticipant int64 `json:"id_participant" sql_on_delete:"CASCADE"`
	IdEquipier    int64 `json:"id_equipier" sql_on_delete:"CASCADE"`
	// redondance pour assurer l'intégrité
	IdGroupe int64 `json:"id_groupe" sql_on_delete:"CASCADE"`
}

// Participantsimple représente un inscrit sur un séjour simplifié,
// sans dossier ni groupe associé.
// sql:ADD UNIQUE(id_personne, id_camp)
// assure qu'un participant simple appartient à un camp simple
// sql:ALTER COLUMN isSimple SET DEFAULT TRUE
// sql:ADD CHECK(isSimple = TRUE)
// sql:ADD FOREIGN KEY (id_camp, isSimple) REFERENCES camps(id,inscription_simple)
type Participantsimple struct {
	Id         int64 `json:"id"`
	IdPersonne int64 `json:"id_personne" sql_on_delete:"CASCADE"`
	IdCamp     int64 `json:"id_camp"` // devrait être simple

	// Moment d'inscription
	DateHeure Time   `json:"date_heure"`
	Info      String `json:"info"`

	isSimple bool
}

type Structureaide struct {
	Id              int64  `json:"id"`
	Nom             String `json:"nom"`
	Immatriculation String `json:"immatriculation"`
	Adresse         String `json:"adresse"`
	CodePostal      String `json:"code_postal"`
	Ville           String `json:"ville"`
	Telephone       String `json:"telephone"`
	Info            String `json:"info"`
}

type Aide struct {
	Id              int64 `json:"id"`
	IdStructureaide int64 `json:"id_structureaide"`
	IdParticipant   int64 `json:"id_participant" sql_on_delete:"CASCADE"`
	Valeur          Euros `json:"valeur"`
	Valide          Bool  `json:"valide"`
	ParJour         Bool  `json:"par_jour"`
	NbJoursMax      Int   `json:"nb_jours_max"`
}

// Organisme désigne un groupe de personne (typiquement une église)
// Un organisme a soit des coordonnées propre, soit une personne de contact
// En plus, un contact pour les dons peut être ajouté.
// sql:ADD CHECK(contact_propre <> (id_contact IS NOT NULL))
type Organisme struct {
	Id  int64  `json:"id"`
	Nom String `json:"nom"`

	ContactPropre Bool        `json:"contact_propre"`
	Contact       Coordonnees `json:"contact"`
	IdContact     OptionnalId `json:"id_contact" sql_foreign_key:"personne"`

	IdContactDon OptionnalId `json:"id_contact_don" sql_foreign_key:"personne"`
	Exemplaires  Exemplaires `json:"exemplaires"` // pour les communications
}

type Don struct {
	Id            int64       `json:"id"`
	Valeur        Euros       `json:"valeur"`
	ModePaiement  ModePaiment `json:"mode_paiement"`
	DateReception Date        `json:"date_reception"`
	RecuEmis      Date        `json:"recu_emis"`
	Infos         InfoDon     `json:"infos"`
	Remercie      Bool        `json:"remercie"`    // `true` si le remerciement a été envoyé
	Details       String      `json:"details"`     // détails additionels
	Affectation   String      `json:"affectation"` // indicatif
}

// DonDonateur est une table de lien indiquant l'origine d'un don,
// groupe ou personne
// sql:ADD CHECK(id_personne <> null OR id_organisme <> null)
// sql:ADD CHECK(id_personne = null OR id_organisme = null)
// sql:ADD UNIQUE(id_don)
type DonDonateur struct {
	IdDon       int64       `json:"id_don" sql_on_delete:"CASCADE"`
	IdPersonne  OptionnalId `json:"id_personne"`
	IdOrganisme OptionnalId `json:"id_organisme"`
}

// ---------------------------- Documents et contraintes ----------------------------

// Contrainte encode la catégorie d'un document à fournir
// L'attribut 'builtin' permet d'identifier des contraintes universelles
//
// sql:ADD UNIQUE(nom, builtin, id_personne)
// sql:ADD CHECK(builtin <> ” OR id_personne IS NOT NULL)
type Contrainte struct {
	Id         int64       `json:"id"`
	IdPersonne OptionnalId `json:"id_personne" sql_on_delete:"CASCADE"` // proprietaire de la contrainte
	IdDocument OptionnalId `json:"id_document"`                         // document à remplir

	Builtin     BuiltinContrainte `json:"builtin"`
	Nom         String            `json:"nom"`
	Description String            `json:"description"`
	// nombre max de documents qui peuvent satisfaire la contrainte
	MaxDocs int `json:"max_docs"`
	// si > 0 indique un document temporaire :
	// les documents liés seront supprimés JoursValide jours
	// après leur dernière modification
	JoursValide int `json:"jours_valide"`
}

// sql:ADD UNIQUE(id_equipier, id_contrainte)
type EquipierContrainte struct {
	IdEquipier   int64 `json:"id_equipier" sql_on_delete:"CASCADE"`
	IdContrainte int64 `json:"id_contrainte"`
	Optionnel    bool  `json:"optionnel"`
}

// CampContrainte représente une contrainte demandée
// à tous les participants
// sql:ADD UNIQUE(id_camp, id_contrainte)
type CampContrainte struct {
	IdCamp       int64 `json:"id_camp" sql_on_delete:"CASCADE"`
	IdContrainte int64 `json:"id_contrainte"`
}

// GroupeContrainte représente une contrainte
// attribuée à un groupe spécifique (de façon cumulative
// avec les contraintes camps)
// sql:ADD UNIQUE(id_groupe, id_contrainte)
type GroupeContrainte struct {
	IdGroupe     int64 `json:"id_groupe" sql_on_delete:"CASCADE"`
	IdContrainte int64 `json:"id_contrainte"`
}

// Document représente les méta données d'un document stocké sur le serveur
// Son contenu et son utilisation sont définis par les objets `TargetDocument` et `ContenuDocument`
type Document struct {
	Id int64 `json:"id"`

	// En bytes
	Taille         Taille `json:"taille"`
	NomClient      String `json:"nom_client"`
	Description    String `json:"description"`
	DateHeureModif Time   `json:"date_heure_modif"`
}

// ContenuDocument stocke le contenu effectif du document, compressé
// ainsi qu'une miniature
// sql:ADD UNIQUE(id_document)
type ContenuDocument struct {
	IdDocument int64 `json:"id_document" sql_on_delete:"CASCADE"`
	// Deprecated: empty on the database
	Contenu   []byte `json:"contenu"`
	Miniature []byte `json:"miniature"`
}

// table de lien pour les lettres des séjours et les documents additionnels
// sql:	ADD UNIQUE(id_document)
type DocumentCamp struct {
	IdDocument int64 `json:"id_document" sql_on_delete:"CASCADE"`
	IdCamp     int64 `json:"id_camp"`
	IsLettre   bool  `json:"is_lettre"` // sinon, document additionnel
}

// table de lien pour les documents liés aux personnes
// un document peut remplir une contrainte personnalisée
// sql:ADD UNIQUE(id_document)
type DocumentPersonne struct {
	IdDocument   int64 `json:"id_document" sql_on_delete:"CASCADE"`
	IdPersonne   int64 `json:"id_personne"`
	IdContrainte int64 `json:"id_contrainte"`
}

// table de lien pour les justificatifs des aides
// sql:ADD UNIQUE(id_document)
type DocumentAide struct {
	IdDocument int64 `json:"id_document" sql_on_delete:"CASCADE"`
	IdAide     int64 `json:"id_aide"`
}

// --------------------------------- ---------------------------------
type Paiement struct {
	Id        int64 `json:"id"`
	IdFacture int64 `json:"id_facture"`

	IsAcompte       Bool        `json:"is_acompte"`
	IsRemboursement Bool        `json:"is_remboursement"`
	InBordereau     Time        `json:"in_bordereau"`
	LabelPayeur     String      `json:"label_payeur"`
	NomBanque       String      `json:"nom_banque"`
	ModePaiement    ModePaiment `json:"mode_paiement"`
	Numero          String      `json:"numero"`
	Valeur          Euros       `json:"valeur"`
	IsInvalide      Bool        `json:"is_invalide"`
	DateReglement   Time        `json:"date_reglement"`
	Details         String      `json:"details"`
}

type Facture struct {
	Id                      int64                   `json:"id"`
	IdPersonne              int64                   `json:"id_personne"`
	DestinatairesOptionnels DestinatairesOptionnels `json:"destinataires_optionnels"`
	Key                     String                  `json:"key"`
	CopiesMails             pq.StringArray          `json:"copies_mails"` // liste d'adresse en copies des mails envoyés

	LastConnection time.Time `json:"last_connection"` // connection sur l'espace de suivi

	IsConfirmed bool `json:"is_confirmed"` // mail confirmé par les parents
	IsValidated bool `json:"is_validated"` // validée par le centre

	// Autorisation de partage des adresses
	PartageAdressesOK bool `json:"partage_adresses_ok"`
}

// Message encode un échange entre le centre d'inscription
// et le responsable d'un dossier
// sql:ADD UNIQUE(id, kind)
// Les définitions suivantes sont temporaires et permettent de
// solidifer la migration
// noTableSql:CREATE FUNCTION m_responsable() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MResponsable';
// noTableSql:CREATE FUNCTION m_centre() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MCentre';
// noTableSql:CREATE FUNCTION m_accuse_reception() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MAccuseReception';
// noTableSql:CREATE FUNCTION m_facture() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MFacture';
// noTableSql:CREATE FUNCTION m_documents() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MDocuments';
// noTableSql:CREATE FUNCTION m_facture_acquittee() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MFactureAcquittee';
// noTableSql:CREATE FUNCTION m_attestation_presence() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MAttestationPresence';
// noTableSql:CREATE FUNCTION m_sondage() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MSondage';
// noTableSql:CREATE FUNCTION m_inscription() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MInscription';
// noTableSql:CREATE FUNCTION m_place_liberee() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #MessageKind.MPlaceLiberee';
type Message struct {
	Id        int64       `json:"id"`
	IdFacture int64       `json:"id_facture" sql_on_delete:"CASCADE"`
	Kind      MessageKind `json:"kind"`
	Created   Time        `json:"created"`
	Modified  Time        `json:"modified"`
	Vu        bool        `json:"vu"` // indique si le message a été vu (dépend du contexte)
}

// Définit les tables complétant un message (à garder synchronisé)
var ComplementMessagesTables = [...]string{
	"message_documents",
	"message_sondages",
	"message_placeliberes",
	"message_attestations",
	"message_messages",
	"message_views",
}

// MessageDocument complete un message 'documents'
// en donnant le camp concerné.
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guardKind = #MessageKind.MDocuments)
// sql:ALTER COLUMN guardKind SET DEFAULT #MessageKind.MDocuments
// sql:ADD FOREIGN KEY (id_message, guardKind) REFERENCES messages(id,kind)
type MessageDocument struct {
	IdMessage int64 `json:"id_message" sql_on_delete:"CASCADE"`
	IdCamp    int64 `json:"id_camp"`

	guardKind MessageKind
}

// MessageSondage complete un message 'sondage'
// en donnant le camp concerné.
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guardKind = #MessageKind.MSondage)
// sql:ALTER COLUMN guardKind SET DEFAULT #MessageKind.MSondage
// sql:ADD FOREIGN KEY (id_message, guardKind) REFERENCES messages(id,kind)
// sql:ALTER COLUMN isSimple SET DEFAULT FALSE
// sql:ADD CHECK(isSimple = FALSE)
// sql:ADD FOREIGN KEY (id_camp, isSimple) REFERENCES camps(id,inscription_simple)
type MessageSondage struct {
	IdMessage int64 `json:"id_message" sql_on_delete:"CASCADE"`
	IdCamp    int64 `json:"id_camp"`

	guardKind MessageKind
	isSimple  bool
}

// MessagePlacelibere complète une notification de place libérée
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guardKind = #MessageKind.MPlaceLiberee)
// sql:ALTER COLUMN guardKind SET DEFAULT #MessageKind.MPlaceLiberee
// sql:ADD FOREIGN KEY (id_message, guardKind) REFERENCES messages(id,kind)
type MessagePlacelibere struct {
	IdMessage     int64 `json:"id_message" sql_on_delete:"CASCADE"`
	IdParticipant int64 `json:"id_participant"`

	guardKind MessageKind
}

// MessageAttestation complète l'accès
// à une facture acquittée/attestation de présence
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guard_kind = #MessageKind.MFactureAcquittee OR guard_kind = #MessageKind.MAttestationPresence)
// sql:ADD FOREIGN KEY (id_message, guard_kind) REFERENCES messages(id,kind)
// pour la migration
// noTableSql:CREATE FUNCTION d_mail() RETURNS int LANGUAGE sql IMMUTABLE PARALLEL SAFE AS 'SELECT #Distribution.DMail';
type MessageAttestation struct {
	IdMessage    int64        `json:"id_message" sql_on_delete:"CASCADE"`
	Distribution Distribution `json:"distribution"`

	GuardKind MessageKind `json:"guard_kind"`
}

// MessageMessage complète l'accès à un message libre
// sql:ADD UNIQUE(id_message)
// contraintes d'intégrité :
// sql:ADD CHECK(guard_kind = #MessageKind.MResponsable OR guard_kind = #MessageKind.MCentre)
// sql:ADD FOREIGN KEY (id_message, guard_kind) REFERENCES messages(id,kind)
type MessageMessage struct {
	IdMessage int64  `json:"id_message" sql_on_delete:"CASCADE"`
	Contenu   String `json:"contenu"`

	GuardKind MessageKind `json:"guard_kind"`
}

// MessageView indique si une demande particulière a été lue par le directeur.
// sql:ADD CHECK(guardKind = #MessageKind.MResponsable)
// sql:ALTER COLUMN guardKind SET DEFAULT #MessageKind.MResponsable
// sql:ADD FOREIGN KEY (id_message, guardKind) REFERENCES messages(id,kind)
// sql:ADD UNIQUE(id_message, id_camp)
type MessageView struct {
	IdMessage int64 `json:"id_message" sql_on_delete:"CASCADE"`
	IdCamp    int64 `json:"id_camp" sql_on_delete:"CASCADE"`
	Vu        bool  `json:"vu"`
	guardKind MessageKind
}

// ---------------------------- Users ----------------------------

// User représente un utilisateur du client
type User struct {
	Id      int64   `json:"id"`
	Label   String  `json:"label"`
	Mdp     String  `json:"mdp"`
	IsAdmin Bool    `json:"is_admin"`
	Modules Modules `json:"modules"`
}

package rawdata

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AnneesCotisations représentes les années où
// une cotisation est demandée, triée du plus récent au plus ancien.
var AnneesCotisations []int

const (
	NonInvite InvitationEquipier = iota
	Invite
	Verifie
)

func init() {
	current_year := time.Now().Year()
	for i := current_year; i >= origineCotisation; i-- {
		AnneesCotisations = append(AnneesCotisations, i)
	}
}

// types utilisés. Permet une implémentation personnalisée
// de rendering, search, etc....

// Int is integer
type Int int

type OptionnalId sql.NullInt64

// NewOptionnalId renvoie un id valid.
func NewOptionnalId(id int64) OptionnalId {
	return OptionnalId{Valid: true, Int64: id}
}

func NullId() OptionnalId {
	return OptionnalId{}
}

func FromIId(id IId) OptionnalId {
	if id == nil {
		return NullId()
	}
	return NewOptionnalId(id.Int64())
}

// AsIId renvoie nil ou un Id
func (o OptionnalId) AsIId() IId {
	if o.IsNil() {
		return nil
	}
	return Id(o.Int64)
}

func (o OptionnalId) IsNil() bool {
	return !o.IsNotNil()
}

func (o OptionnalId) IsNotNil() bool {
	return o.Valid
}

// Is renvoie `true` si l'id est valide et vaut `id`
func (o OptionnalId) Is(id int64) bool {
	return o.Valid && o.Int64 == id
}

type IdentificationId struct {
	// `true` pour une identification
	Valid bool  `json:"valid,omitempty"`
	Id    int64 `json:"id,omitempty"`

	// Usefull to send to the browser. Useless internally or in the client
	Crypted string `json:"crypted,omitempty"`
}

// String is string
type String string

// Bool is bool
type Bool bool

// OptionnalBool représente une quantité à 3 états : Oui, Non, Indifférent
type OptionnalBool int8

// Bool renvoi `true` uniquement si `b` vaut Oui
func (b OptionnalBool) Bool() bool {
	return b == OBOui
}

func (b OptionnalBool) Check() error {
	switch b {
	case 0, OBOui, OBNon:
		return nil
	default:
		return fmt.Errorf("invalid value for optionnal bool : %d", b)
	}
}

// Float is 64bit float
type Float float64

// Time is time.Time
type Time time.Time

// Date représente une date
type Date time.Time

// Euros représente un prix en euros
type Euros float64

// Round arrondi à deux chiffres après la virgule
func (e Euros) Round() float64 {
	centimes := int(math.Round(float64(e) * 100))
	return float64(centimes) / 100
}

// Centimes renvoi le nombres de centimes
func (e Euros) Centimes() int64 {
	return int64(math.Round(float64(e) * 100))
}

// IsLess renvoie `true` si le prix est inférieur ou égal à `other`
// au sens de 2 décimales
func (e Euros) IsLess(other Euros) bool {
	diff := e.Round() - other.Round()
	return diff <= 0.01
}

// IsEqual renvoie `true` si le prix est égal à `other`
// au sens de 2 décimales
func (e Euros) IsEqual(other Euros) bool {
	diff := e.Round() - other.Round()
	return diff == 0
}

// Remise applique une remise en pourcentage (renvoie au minimum 0)
func (e Euros) Remise(rem Pourcent) Euros {
	p := e * Euros(100-rem) / 100
	if p < 0 {
		p = 0
	}
	return p
}

// Pourcent est entre 0 et 100
type Pourcent int

type Taille int

// Montant modélise une valeur pour être (ou non) relative
type Montant struct {
	parJour bool
	valeur  Euros
}

func (m Montant) ValeurEffective(nbJours int) Euros {
	if m.parJour {
		return Euros(nbJours) * m.valeur
	}
	return m.valeur
}

// Time convertit la date en l'objet standard.
func (d Date) Time() time.Time {
	d2 := time.Time(d)
	return d2.Truncate(24 * time.Hour)
}

// Time convertit l'instant en l'objet standard.
func (d Time) Time() time.Time {
	d2 := time.Time(d)
	return d2.Truncate(time.Second)
}

func (d Date) Equals(other Date) bool {
	dt, ot := d.Time(), other.Time()
	return dt.Year() == ot.Year() && dt.Month() == ot.Month() && dt.Day() == ot.Day()
}

type Plage struct {
	From Date `json:"from,omitempty"`
	To   Date `json:"to,omitempty"`
}

func (plage Plage) ExpandMonths() []Date {
	to := plage.To.Time()
	to = time.Date(to.Year(), to.Month(), plage.From.Time().Day()+1, 0, 0, 0, 0, time.UTC)
	var out []Date
	for plage.From.Time().Before(to) {
		out = append(out, plage.From)
		plage.From = Date(plage.From.Time().AddDate(0, 1, 0))
	}
	return out
}

// NbJours renvoie la durée de la plage, arrivée et départ COMPRIS.
func (plage Plage) NbJours() int {
	day := 24 * time.Hour
	debut := plage.From.Time().Truncate(day) // on enlève les heures, minutes, secondes...
	fin := plage.To.Time().Truncate(day)     // idem
	diff := (fin.Sub(debut) / day)
	return int(diff) + 1 // pour prendre en compte le dernier jour
}

// Contains indique si `date` est contenue dans la plage.
// Le critère est lâche : si `plage` est nulle, la `date` est acceptée.
func (plage Plage) Contains(date time.Time) bool {
	date = date.Truncate(24 * time.Hour)
	isDebutOK := plage.From.Equals(Date(date)) || plage.From.Time().Before(date)
	isFinOK := plage.To.Time().IsZero() || plage.To.Equals(Date(date)) || plage.To.Time().After(date)
	return isDebutOK && isFinOK
}

type OptionnalPlage struct {
	Plage
	Active bool `json:"active,omitempty"`
}

// ------------------- Personne ------------------- //

// "M" ou "F"
type Sexe string

// Tels contient plusieurs numéros
type Tels []string

// RangMembreAsso indique la position dans l'association
type RangMembreAsso string

func (r RangMembreAsso) toInt() int {
	if r == RMANonMembre {
		r = "0"
	}
	out, _ := strconv.Atoi(string(r))
	return out // en cas d'erreur renvoie le catégorie 0
}

// AtLeast renvoie `true` si `r` fait partie (au sens large)
// de la catégorie `other`
func (r RangMembreAsso) AtLeast(other RangMembreAsso) bool {
	return r.toInt() >= other.toInt()
}

// Pays est le code ISO 3166 d'un pays
type Pays string

// Departement code le numéro d'un département,
// ou le nom du département lui même (pour l'étranger par exemple).
type Departement string

// Cotisation donne les années de cotisation
type Cotisation []int64

func (c Cotisation) Map() map[int]bool {
	out := make(map[int]bool, len(c))
	for _, annee := range c {
		out[int(annee)] = true
	}
	return out
}

// Bilan vérifie si les cotisations ont été payées
func (c Cotisation) Bilan() bool {
	m := c.Map()
	for _, annee := range AnneesCotisations {
		if !m[annee] {
			return false
		}
	}
	return true
}

// ------------------- Camp ------------------- //

// OptionPrixCamp stocke une option sur le prix d'un camp. Une seule est effective,
// déterminée par Active
type OptionPrixCamp struct {
	Active  string            `json:"active,omitempty"`
	Semaine OptionSemaineCamp `json:"semaine,omitempty"`
	Statut  []PrixParStatut   `json:"statut,omitempty"`
	// Prix de chaque jour du camp (souvent constant)
	// Le champ Prix du séjour peut être inférieur à la somme
	// pour une remise.
	Jour []Euros `json:"jour,omitempty"`
}

// SchemaPaiement peut valoir "acompte" ou "total"
// et détermine si on demande un acompte pour le camp donné.
type SchemaPaiement string

// OptionQuotientFamilial applique un pourcentage sur prix de base,
// pour les categories définie par `QuotientFamilial`
// Par cohérence avec le prix de base, la dernière valeur vaut toujours 100
// (sauf pour les entrées vides).
type OptionQuotientFamilial [4]float64

// GetRatio renvoie le pourcentage appliqué au prix de base pour le quotient
// familial donné.
func (q OptionQuotientFamilial) GetRatio(qf Int) float64 {
	n := len(QuotientFamilial)
	for i := 0; i < n-1; i++ {
		q1, q2 := QuotientFamilial[i], QuotientFamilial[i+1]
		if q1 < qf && qf <= q2 {
			return q[i]
		}
	}
	return 100
}

type BuiltinContrainte string

// ------------------- Paiement ------------------- //

// Parmis "cb" "cheque" "vir" "espece" "ancv" "helloasso"
type ModePaiment string

// ------------------- Participant ------------------- //

// Bus précise l'utilisation d'une navette
type Bus string

// Match renvoie `true` si `b` est à prendre en compte
// pour le critère `critère`
func (b Bus) Match(critere Bus) bool {
	if critere == "" { // pas de critère accepte tout
		return true
	}
	if critere == BAller || critere == BRetour {
		// on sélectionne aussi aller_retour
		return b == critere || b == BAllerRetour
	}
	return b == critere
}

// définit la place dans la liste d'attente
type StatutAttente uint

// ListeAttente indique si le participant est en liste d'attente
// Le zéro signifie inscrit.
type ListeAttente struct {
	Statut StatutAttente `json:"statut"`
	Raison string        `json:"raison"`
}

func (l ListeAttente) IsInscrit() bool {
	return l.Statut == Inscrit
}

func (l ListeAttente) String() String {
	if l.IsInscrit() {
		return "Campeur"
	}
	return String(fmt.Sprintf("Liste d'attente (%s)", l.Raison))
}

// HintsAttente expose une série de critère
// qui guide le choix de mise en liste d'attente
type HintsAttente struct {
	AgeMin, AgeMax, EquilibreGF, Place StatutAttente
}

func (h HintsAttente) Causes() []string {
	var causes []string
	if h.AgeMin != Inscrit {
		causes = append(causes, "Age trop faible")
	}
	if h.AgeMax != Inscrit {
		causes = append(causes, "Age trop élevé")
	}
	if h.EquilibreGF != Inscrit {
		causes = append(causes, "Equilibre Garçons/Filles menacé")
	}
	if h.Place != Inscrit {
		causes = append(causes, "Nombre de places insuffisant")
	}
	return causes
}

// String résume les indications. Renvoie une chaine vide
// si tous les critères sont remplis
func (h HintsAttente) String() string {
	causes := h.Causes()
	return strings.Join(causes, ", ")
}

func (h HintsAttente) Sortable() string {
	return h.String()
}

// Hint renvoie le choix "logique" correspondant
// aux critères.
// Par design, ce n'est pas forcément le statut actuel du participant
func (h HintsAttente) Hint() StatutAttente {
	// on prend le "pire" des critères
	// comme ils sont classés, cela revient à prendre le max
	var pire StatutAttente
	for _, s := range [4]StatutAttente{h.AgeMax, h.AgeMin, h.EquilibreGF, h.Place} {
		if s > pire {
			pire = s
		}
	}
	return pire
}

// -------------------------- Equipiers --------------------------

// InvitationEquipier représente l'utilisation de la page équipier :
type InvitationEquipier int

// Role dans un camp
type Role string

func (r Role) Check() error {
	_, in := RoleLabels[r]
	if !in {
		return fmt.Errorf("rôle inconnu : %s", r)
	}
	return nil
}

// IsAuPair renvoie `true` si le rôle est considéré
// comme au pair.
func (r Role) IsAuPair() bool {
	switch r {
	case RDirecteur, RAdjoint, RAnimation, RAideAnimation:
		return true
	default:
		return false
	}
}

func (r Role) AsRoles() Roles {
	return Roles{r}
}

type Roles []Role

func (rs Roles) Check() error {
	if len(rs) == 0 {
		return errors.New("rôle manquant")
	}
	for _, r := range rs {
		if err := r.Check(); err != nil {
			return err
		}
	}
	return nil
}

func (rs Roles) String() string {
	chuncks := make([]string, len(rs))
	for i, v := range rs {
		chuncks[i] = v.String()
	}
	return strings.Join(chuncks, " ; ")
}

func (rs Roles) Sortable() string {
	chuncks := make([]string, len(rs))
	for i, v := range rs {
		chuncks[i] = v.Sortable()
	}
	sort.Strings(chuncks)
	return strings.Join(chuncks, "-")
}

// IsAuPair vérifie si au moins un des rôles est
// considéré comme au pair
func (rs Roles) IsAuPair() bool {
	for _, r := range rs {
		if r.IsAuPair() {
			return true
		}
	}
	return false
}

// Is vérifie si `r` est présent
func (rs Roles) Is(r Role) bool {
	for _, v := range rs {
		if v == r {
			return true
		}
	}
	return false
}

// Diplome est l'identifiant du type de diplome
type Diplome string

// Approfondissement est l'identifiant du type d'appofondissement
type Approfondissement string

// OptionsParticipant donne les champs optionnels pour un inscrit
type OptionsParticipant struct {
	Bus         Bus         `json:"bus"`
	MaterielSki MaterielSki `json:"materiel_ski"`
}

// Semaine précise le choix d'une seule semaine de camp
type Semaine string

// Jours stocke les indexes (0-based) des jours de présence
// d'un participant à un séjour
// Une liste vide indique la présence sur TOUT le séjour.
type Jours []int

// Set renvoie un crible des indexes de présence
func (js Jours) Set() map[int]bool {
	set := map[int]bool{}
	for _, index := range js {
		set[index] = true
	}
	return set
}

// Sanitize vérifie que les jours sont valides
func (js Jours) Sanitize(duree int) error {
	if len(js.Set()) != len(js) {
		return fmt.Errorf("jours de présences invalides: %v", js)
	}
	for _, j := range js {
		if j < 0 || j >= duree {
			return fmt.Errorf("jour de présence invalide (%d)", j)
		}
	}
	return nil
}

// NbJours renvoie le nombre de jours de présence
// en évitant un éventuel doublon
func (js Jours) NbJours(datesCamp Plage) int {
	if len(js) == 0 {
		return datesCamp.NbJours()
	}
	return len(js.Set())
}

// ClosestPlage renvoie la plage englobant les jours de présence
func (js Jours) ClosestPlage(datesCamp Plage) Plage {
	const jour = 24 * time.Hour
	if len(js) == 0 { // zero value : tout le séjour
		return datesCamp
	}
	sort.Ints(js)
	indexMin, indexMax := js[0], js[len(js)-1]
	from := datesCamp.From.Time().Add(jour * time.Duration(indexMin))
	to := datesCamp.From.Time().Add(jour * time.Duration(indexMax))
	return Plage{From: Date(from), To: Date(to)}
}

// CalculePrix somme les prix des journées de présence
func (js Jours) CalculePrix(prix []Euros) Euros {
	var total Euros
	for i := range js.Set() {
		if i >= len(prix) { // ne devrait pas arriver
			continue
		}
		total += prix[i]
	}
	return total
}

// Description renvoie les jours de présence au camp
func (js Jours) Description(camp Camp) string {
	jsSet := js.Set()
	if len(jsSet) == 0 || len(jsSet) == camp.Duree() {
		return "Tout le séjour"
	}
	var sortedKeys []int
	for k := range jsSet {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Ints(sortedKeys)
	var days []string
	for _, index := range sortedKeys { // 0 based
		day := camp.DateDebut.Time().Add(time.Duration(index) * 24 * time.Hour)
		days = append(days, Date(day).shortDesc())
	}
	return strings.Join(days, "; ")
}

// OptionPrixParticipant répond à OptionPrixCamp. L'option est active si :
//   - elle est active dans le camp
//   - elle est non nulle dans le participant
type OptionPrixParticipant struct {
	Semaine Semaine `json:"semaine"`
	Statut  int64   `json:"statut"`
	Jour    Jours   `json:"jour"`
}

// IsNonNil renvoie `true` si une option est active
// pour la catégorie demandée.
func (op OptionPrixParticipant) IsNonNil(categorie string) bool {
	switch categorie {
	case OptionsPrix.SEMAINE:
		return op.Semaine != ""
	case OptionsPrix.STATUT:
		return op.Statut != 0
	case OptionsPrix.JOUR:
		return len(op.Jour) > 0
	default:
		return false
	}
}

// EtatNegociaton indique les avanvements du dossier
// (dates d'envois des mails)
// Les documents sont envoyés camp par camp, d'où un map camp -> moment d'envoi
type EtatNegociation struct {
	AccuseReception     Time                `json:"accuse_reception,omitempty"`
	Facture             Time                `json:"facture,omitempty"`
	Documents           map[int64]time.Time `json:"documents,omitempty"`
	FactureAcquittee    Time                `json:"facture_acquittee,omitempty"`
	AttestationPresence Time                `json:"attestation_presence,omitempty"`
}

// ListeVetements associe à un groupe une liste
// d'objet et un complement
type ListeVetements struct {
	Liste      []Vetement `json:"liste,omitempty"`
	Complement string     `json:"complement,omitempty"`
}

type TrajetBus struct {
	RendezVous String `json:"rendez_vous"`
	Prix       Euros  `json:"prix"`
}

type BusCamp struct {
	Actif       bool      `json:"actif"`
	Commentaire String    `json:"commentaire"`
	Aller       TrajetBus `json:"aller"`
	Retour      TrajetBus `json:"retour"`
}

type OptionsCamp struct {
	Bus         BusCamp         `json:"bus,omitempty"`
	MaterielSki MaterielSkiCamp `json:"materiel_ski,omitempty"`
}

type Exemplaires struct {
	PubEte     int `json:"pub_ete,omitempty"`
	PubHiver   int `json:"pub_hiver,omitempty"`
	EchoRocher int `json:"echo_rocher,omitempty"`
	EOnews     int `json:"e_onews,omitempty"`
}

// MaterielSki indique le matériel souhaité par un participant.
type MaterielSki struct {
	Need     string `json:"need,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Casque   bool   `json:"casque,omitempty"`
	Poids    int    `json:"poids,omitempty"`
	Taille   int    `json:"taille,omitempty"`
	Pointure int    `json:"pointure,omitempty"`
}

func (s *MaterielSki) IsZero() bool {
	return s.Need == "" && s.Mode == "" && !s.Casque && s.Poids == 0 && s.Taille == 0 && s.Pointure == 0
}

//--------------------------------------------------------------------
//------------------------ Fiche Sanitaire ---------------------------
//--------------------------------------------------------------------

type Maladies struct {
	Rubeole    bool `json:"rubeole,omitempty"`
	Varicelle  bool `json:"varicelle,omitempty"`
	Angine     bool `json:"angine,omitempty"`
	Oreillons  bool `json:"oreillons,omitempty"`
	Scarlatine bool `json:"scarlatine,omitempty"`
	Coqueluche bool `json:"coqueluche,omitempty"`
	Otite      bool `json:"otite,omitempty"`
	Rougeole   bool `json:"rougeole,omitempty"`
	Rhumatisme bool `json:"rhumatisme,omitempty"`
}

// List retourne les maladies présentes.
func (m Maladies) List() []string {
	var out []string
	if m.Rubeole {
		out = append(out, "Rubéole")
	}
	if m.Varicelle {
		out = append(out, "Varicelle")
	}
	if m.Angine {
		out = append(out, "Angine")
	}
	if m.Oreillons {
		out = append(out, "Oreillons")
	}
	if m.Scarlatine {
		out = append(out, "Scarlatine")
	}
	if m.Coqueluche {
		out = append(out, "Coqueluche")
	}
	if m.Otite {
		out = append(out, "Otite")
	}
	if m.Rougeole {
		out = append(out, "Rougeole")
	}
	if m.Rhumatisme {
		out = append(out, "Rhumatisme articulaire aigü")
	}
	return out
}

type Allergies struct {
	Asthme          bool   `json:"asthme,omitempty"`
	Alimentaires    bool   `json:"alimentaires,omitempty"`
	Medicamenteuses bool   `json:"medicamenteuses,omitempty"`
	Autres          string `json:"autres,omitempty"`
	ConduiteATenir  string `json:"conduite_a_tenir,omitempty"`
}

func (a Allergies) List() []string {
	var out []string
	if a.Asthme {
		out = append(out, "Asthme")
	}
	if a.Alimentaires {
		out = append(out, "Alimentaires")
	}
	if a.Medicamenteuses {
		out = append(out, "Médicamenteuses")
	}
	if a.Autres != "" {
		out = append(out, a.Autres)
	}
	return out
}

type Medecin struct {
	Nom string `json:"nom,omitempty"`
	Tel string `json:"tel,omitempty"`
}

// FicheSanitaire stocke les informations remplies sur l'espace perso.
// A coordonner avec le contact du participant (nom, Prenom, Adresse, CodePostal, Ville, Tels, Sécurité Sociale)
type FicheSanitaire struct {
	TraitementMedical bool      `json:"traitement_medical,omitempty"`
	Maladies          Maladies  `json:"maladies,omitempty"`
	Allergies         Allergies `json:"allergies,omitempty"`
	DifficultesSante  string    `json:"difficultes_sante,omitempty"`
	Recommandations   string    `json:"recommandations,omitempty"`
	Handicap          bool      `json:"handicap,omitempty"`
	Tel               string    `json:"tel,omitempty"` // en plus des numéros d'un éventuel contact
	Medecin           Medecin   `json:"medecin,omitempty"`

	LastModif Time     `json:"last_modif,omitempty"` // dernière modification
	Mails     []string `json:"mails,omitempty"`      // mail des propiétaires, pour sécurité.
}

// IsNone renvoie `true` si la fiche n'a jamais été modifiée.
func (f FicheSanitaire) IsNone() bool {
	return f.LastModif.Time().IsZero()
}

// Remises donne les champs optionnels concernant le suivi financier d'un participant
type Remises struct {
	ReducEquipiers Pourcent `json:"reduc_equipiers,omitempty"`
	ReducEnfants   Pourcent `json:"reduc_enfants,omitempty"`
	ReducSpeciale  Euros    `json:"reduc_speciale,omitempty"`
}

func (r Remises) Pourcent() Pourcent {
	return r.ReducEquipiers + r.ReducEnfants
}

// IsActive renvoie `true` si au moins une remise est active
func (r Remises) IsActive() bool {
	return r.ReducSpeciale > 0 || r.ReducEquipiers > 0 || r.ReducEnfants > 0
}

type Envois struct {
	Locked            bool `json:"__locked__,omitempty"`
	LettreDirecteur   bool `json:"lettre_directeur,omitempty"`
	ListeVetements    bool `json:"liste_vetements,omitempty"`
	ListeParticipants bool `json:"liste_participants,omitempty"`
}

type Vetement struct {
	Quantite    int    `json:"quantite,omitempty"`
	Description string `json:"description,omitempty"`
	Obligatoire bool   `json:"obligatoire,omitempty"`
}

type OptionSemaineCamp struct {
	Plage1 Plage `json:"plage_1"`
	Plage2 Plage `json:"plage_2"`
	Prix1  Euros `json:"prix_1,omitempty"`
	Prix2  Euros `json:"prix_2,omitempty"`
}

type PrixParStatut struct {
	Id          int64  `json:"id,omitempty"`
	Prix        Euros  `json:"prix,omitempty"`
	Statut      String `json:"statut,omitempty"`
	Description String `json:"description,omitempty"`
}

type MaterielSkiCamp struct {
	Actif      bool  `json:"actif,omitempty"`
	PrixAcve   Euros `json:"prix_acve,omitempty"`
	PrixLoueur Euros `json:"prix_loueur,omitempty"`
}

type Destinataire struct {
	NomPrenom  String `json:"nom_prenom,omitempty"`
	Sexe       Sexe   `json:"sexe,omitempty"`
	Adresse    String `json:"adresse,omitempty"`
	CodePostal String `json:"code_postal,omitempty"`
	Ville      String `json:"ville,omitempty"`
}

type DestinatairesOptionnels []Destinataire

type Completion uint

type Distribution uint8

// ClientUser représent un utilisateur, vu par le client
type ClientUser struct {
	Id      int64
	Label   string
	IsAdmin bool
}

// InfoDon est un champ caché qui stocke des infos
// additionnelles
type InfoDon struct {
	IdPaiementHelloAsso string `json:"id_paiement_hello_asso,omitempty"`
}

// Satisfaction est une énumération indiquant le
// niveau de satisfaction sur le sondage de fin de séjour
type Satisfaction uint8

type RepSondage struct {
	InfosAvantSejour   Satisfaction `json:"infos_avant_sejour"`
	InfosPendantSejour Satisfaction `json:"infos_pendant_sejour"`
	Hebergement        Satisfaction `json:"hebergement"`
	Activites          Satisfaction `json:"activites"`
	Theme              Satisfaction `json:"theme"`
	Nourriture         Satisfaction `json:"nourriture"`
	Hygiene            Satisfaction `json:"hygiene"`
	Ambiance           Satisfaction `json:"ambiance"`
	Ressenti           Satisfaction `json:"ressenti"`
	MessageEnfant      String       `json:"message_enfant"`
	MessageResponsable String       `json:"message_responsable"`
}

type Coordonnees struct {
	Tels       Tels   `json:"tels"`
	Mail       String `json:"mail"`
	Adresse    String `json:"adresse"`
	CodePostal String `json:"code_postal"`
	Ville      String `json:"ville"`
	Pays       Pays   `json:"pays"`
}

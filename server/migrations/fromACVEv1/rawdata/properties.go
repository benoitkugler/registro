package rawdata

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const CampDeltaTerminated = 45 // jours

const (
	mcMail Field = iota
	mcNom
	mcPrenom
)

var HeadersMails = []Header{
	{Field: mcMail, Label: "Mail"},
	{Field: mcNom, Label: "Nom"},
	{Field: mcPrenom, Label: "Prénom"},
}

// Avant renvoie true si le jour et le mois de date1 sont avant (au sens large) le jour et le mois de date2.
// Ignore les années. Prend en compte le cas du 29 fév.
func Avant(d1 Date, d2 Date) bool {
	t1, t2 := time.Time(d1), time.Time(d2)
	if t1.Month() < t2.Month() {
		return true
	}
	if t1.Month() > t2.Month() {
		return false
	}
	return t1.Day() <= t2.Day()
}

// Pour trier par âge manière naturelle, l'âge ne suffit pour distinguer les participants
// de même âge.
type AgeDate struct {
	age  int
	date Date
}

func (a AgeDate) String() string {
	return Int(a.age).String()
}

func (a AgeDate) Sortable() string {
	// en fait, il suffit de retourner la date de naissance ici
	return a.date.Sortable()
}

func (a AgeDate) Age() int {
	return a.age
}

// CalculeAge renvoie l'age qu'aura une personne née le `dateNaissance`
// au jour `now`
func CalculeAge(dateNaissance Date, now time.Time) AgeDate {
	if dateNaissance.Time().IsZero() {
		return AgeDate{}
	}

	if now.IsZero() {
		now = time.Now()
	}

	years := now.Year() - dateNaissance.Time().Year()
	if Avant(dateNaissance, Date(now)) {
		return AgeDate{age: years, date: dateNaissance}
	}
	return AgeDate{age: years - 1, date: dateNaissance}
}

// ----------------------------------- Personne ----------------------------------- //
// BasePersonne stocke les champs comparés lors du rapprochement
// des inscriptions, des équipiers ou des donateurs
type BasePersonne struct {
	Nom                  String      `json:"nom"`
	NomJeuneFille        String      `json:"nom_jeune_fille"`
	Prenom               String      `json:"prenom"`
	DateNaissance        Date        `json:"date_naissance"`
	VilleNaissance       String      `json:"ville_naissance"`
	DepartementNaissance Departement `json:"departement_naissance"`
	Sexe                 Sexe        `json:"sexe"`
	Tels                 Tels        `json:"tels"`
	Mail                 String      `json:"mail"`
	Adresse              String      `json:"adresse"`
	CodePostal           String      `json:"code_postal"`
	Ville                String      `json:"ville"`
	Pays                 Pays        `json:"pays"`
	SecuriteSociale      String      `json:"securite_sociale"`
	Profession           String      `json:"profession"`
	Etudiant             Bool        `json:"etudiant"`
	Fonctionnaire        Bool        `json:"fonctionnaire"`
}

func (r BasePersonne) FPrenom() string {
	return FormatPrenom(r.Prenom)
}

func (r BasePersonne) FNom() string {
	return r.Nom.ToUpper()
}

// NomPrenom renvoie NOM Prenom
func (r BasePersonne) NomPrenom() String {
	return String(r.FNom() + " " + r.FPrenom())
}

func (r BasePersonne) MailItem() Item {
	return Item{Fields: F{
		mcNom:    String(r.FNom()),
		mcPrenom: String(r.FPrenom()),
		mcMail:   r.Mail,
	}}
}

func (r BasePersonne) AgeDate() AgeDate {
	dn := r.DateNaissance
	if dn.Time().IsZero() {
		return AgeDate{}
	}
	return CalculeAge(dn, time.Time{})
}

func (r BasePersonne) Age() int {
	return r.AgeDate().Age()
}

func (r BasePersonne) ToDestinataire() Destinataire {
	return Destinataire{
		NomPrenom:  r.NomPrenom(),
		Adresse:    r.Adresse,
		CodePostal: r.CodePostal,
		Ville:      r.Ville,
		Sexe:       r.Sexe,
	}
}

// Coordonnees selectionne les champs des coordonnées.
func (r BasePersonne) Coordonnees() Coordonnees {
	return Coordonnees{
		Tels:       r.Tels,
		Mail:       r.Mail,
		Adresse:    r.Adresse,
		CodePostal: r.CodePostal,
		Ville:      r.Ville,
		Pays:       r.Pays,
	}
}

// ----------------------------------- Camp ----------------------------------- //

// Label renvoie une description courte
func (c Camp) Label() String {
	return String(fmt.Sprintf("%s %d", c.Nom, time.Time(c.DateDebut).Year()))
}

// Label renvoie une description courte contenant le mois
func (c Camp) LabelWithMonth() String {
	date := c.DateDebut.Time()
	return String(fmt.Sprintf("%s %02d/%d", c.Nom, date.Month(), date.Year()))
}

// Description renvoi une description détaillée,
// au format HTML
func (c Camp) Description() string {
	return fmt.Sprintf(`Lieu : <b>%s</b> ; Age: de <b>%d à %d</b> ans`, c.Lieu, c.AgeMin, c.AgeMax)
}

// Periode renvoie la période du camp (cf `PERIODES`)
func (c Camp) Periode() String {
	month := time.Time(c.DateDebut).Month()
	switch month {
	case 7, 8:
		return PERIODES.ETE
	case 9, 10, 11:
		return PERIODES.AUTOMNE
	case 12, 1, 2, 3:
		return PERIODES.HIVER
	case 4, 5, 6:
		return PERIODES.PRINTEMPS
	default:
		return ""
	}
}

func (c Camp) ColorPeriode() Color {
	switch c.Periode() {
	case PERIODES.ETE:
		return RGBA{R: 45, G: 185, B: 187, A: 200}
	case PERIODES.PRINTEMPS:
		return RGBA{R: 170, G: 228, B: 62, A: 200}
	case PERIODES.AUTOMNE:
		return RGBA{R: 173, G: 116, B: 30, A: 200}
	case PERIODES.HIVER:
		return RGBA{R: 203, G: 199, B: 193, A: 200}
	default:
		return nil
	}
}

// Duree renvoi le nombre de jours du camp, ou 0 si les dates sont manquantes
func (c Camp) Duree() int {
	if c.DateDebut.Time().IsZero() || c.DateFin.Time().IsZero() {
		return 0
	}
	return Plage{From: c.DateDebut, To: c.DateFin}.NbJours()
}

func (c Camp) Annee() Int {
	return Int(c.DateDebut.Time().Year())
}

// IsTerminated renvoie `true` si le camp est
// passé d'au moins `CampDeltaTerminated`.
// Un camp sans date de fin est considéré comme terminé.
func (c Camp) IsTerminated() bool {
	dateFin := time.Time(c.DateFin)
	if dateFin.IsZero() {
		return true
	}
	return time.Now().After(dateFin.AddDate(0, 0, CampDeltaTerminated))
}

// CheckEnvoisLock renvoie une erreur si l'envoi des documents est encore verouillé
func (c Camp) CheckEnvoisLock() error {
	if c.Envois.Locked {
		return fmt.Errorf("L'envoi des documents du séjour %s est encore verrouillé.", c.Label())
	}
	return nil
}

// AgeDebutCamp renvoie l'âge qu'aura `personne` au premier jour
// du séjour.
func (c Camp) AgeDebutCamp(personne BasePersonne) AgeDate {
	dateArrivee := c.DateDebut
	if dateArrivee.Time().IsZero() {
		return personne.AgeDate()
	}
	return CalculeAge(personne.DateNaissance, dateArrivee.Time())
}

// AgeFinCamp renvoie l'âge qu'aura `personne` au dernier jour
// du séjour.
func (c Camp) AgeFinCamp(personne BasePersonne) AgeDate {
	dateDepart := c.DateFin
	if dateDepart.Time().IsZero() {
		return personne.AgeDate()
	}
	return CalculeAge(personne.DateNaissance, dateDepart.Time())
}

// HasAgeValide renvoie le statut correspondant aux âges min et max du séjour
// Seul le camp DateNaissance de `personne` est utilisé.
func (c Camp) HasAgeValide(personne BasePersonne) (min StatutAttente, max StatutAttente) {
	ageDebut, ageFin := Int(c.AgeDebutCamp(personne).Age()), Int(c.AgeFinCamp(personne).Age())
	invalideDataDebut := personne.DateNaissance.Time().IsZero() || c.DateDebut.Time().IsZero()
	invalideDataFin := personne.DateNaissance.Time().IsZero() || c.DateFin.Time().IsZero()

	isMaxValide := func() StatutAttente {
		if c.AgeMax == 0 { // Pas de critère > pas de dates.
			return Inscrit
		}
		if invalideDataDebut { // impossible de calculer l'age
			return Attente
		}
		if ageDebut > c.AgeMax {
			return Attente
		}
		return Inscrit
	}

	isMinValide := func() StatutAttente {
		if c.AgeMin == 0 { // Pas de critère > pas de dates.
			return Inscrit
		}
		if invalideDataFin { // impossible de calculer l'age
			return Attente
		}
		if ageFin < c.AgeMin {
			return Attente
		}
		return Inscrit
	}
	min, max = isMinValide(), isMaxValide()
	// cas particulier pour l'âge de 6 ans :
	// pour respecter la loi on est obligé de refuser
	if c.AgeMin == 6 && ageFin < c.AgeMin {
		min = Refuse
	}
	return min, max
}

// CheckDureeOptionJour vérifie que l'option jour spécifie un nombre
// de jours correct, si elle est active
func (c *Camp) CheckDureeOptionJour() error {
	opt := c.OptionPrix
	if opt.Active != OptionsPrix.JOUR {
		return nil // rien à vérifier
	}
	if c.Duree() != len(opt.Jour) {
		return fmt.Errorf("Le nombre de jour de l'option <i>Prix à la journée</i> ne correspond pas à la durée du séjour.")
	}
	return nil
}

func (r Aide) Montant() Montant {
	return Montant{
		parJour: bool(r.ParJour),
		valeur:  r.Valeur,
	}
}

func (r Don) Label() string {
	return fmt.Sprintf("montant <b>%s</b>, reçu le <i>%s</i>", r.Valeur.String(), r.DateReception.String())
}

func (f Facture) UrlEspacePerso(host string) string {
	if f.Key == "" {
		return ""
	}
	return path.Join(host, string(f.Key))
}

// Description renvoie une description et le montant, au format HTML
func (r Paiement) Description() (string, string) {
	var payeur string
	if r.IsAcompte {
		payeur = fmt.Sprintf("Acompte de <i>%s</i> au %s", r.LabelPayeur, r.DateReglement.String())
	} else if r.IsRemboursement {
		payeur = fmt.Sprintf("Remboursement au %s", r.DateReglement.String())
	} else {
		payeur = fmt.Sprintf("Paiement de <i>%s</i> au %s", r.LabelPayeur, r.DateReglement.String())
	}
	montant := fmt.Sprintf("<i>%s</i>", r.Valeur.String())
	if r.IsRemboursement {
		montant = "<b>-</b> " + montant
	}
	prefix := ""
	if r.IsInvalide {
		prefix = "<b>Invalide</b> "
		montant = fmt.Sprintf("(%s)", montant)
	}
	return prefix + payeur, montant
}

// Permissions représente les droites sur chaque module,
// qui interprète l'entier fourni. 0 signifie non utilisation.
type Modules struct {
	Personnes     int `json:"personnes,omitempty"`
	Camps         int `json:"camps,omitempty"`
	Inscriptions  int `json:"inscriptions,omitempty"`
	SuiviCamps    int `json:"suivi_camps,omitempty"`
	SuiviDossiers int `json:"suivi_dossiers,omitempty"`
	Paiements     int `json:"paiements,omitempty"`
	Aides         int `json:"aides,omitempty"`
	Equipiers     int `json:"equipiers,omitempty"`
	Dons          int `json:"dons,omitempty"`
}

func (m Modules) ToReadOnly() Modules {
	var out Modules
	if m.Personnes > 0 {
		out.Personnes = 1
	}
	if m.Camps > 0 {
		out.Camps = 1
	}
	if m.Inscriptions > 0 {
		out.Inscriptions = 1
	}
	if m.SuiviCamps > 0 {
		out.SuiviCamps = 1
	}
	if m.SuiviDossiers > 0 {
		out.SuiviDossiers = 1
	}
	if m.Paiements > 0 {
		out.Paiements = 1
	}
	if m.Aides > 0 {
		out.Aides = 1
	}
	if m.Equipiers > 0 {
		out.Equipiers = 1
	}
	if m.Dons > 0 {
		out.Dons = 1
	}
	return out
}

func (c Contraintes) FindBuiltin(categorie BuiltinContrainte) Contrainte {
	for _, ct := range c {
		if ct.Builtin == categorie {
			return ct
		}
	}
	return Contrainte{}
}

type Links struct {
	EquipierContraintes map[int64]EquipierContraintes // id equipier ->
	ParticipantGroupe   map[int64]GroupeParticipant   // idParticipant -> groupe (optionnel)
	ParticipantEquipier map[int64]ParticipantEquipier // idParticipant -> idAnimateur (optionnel)
	GroupeContraintes   map[int64]GroupeContraintes   // idGroupe -> ids contraintes
	CampContraintes     map[int64]CampContraintes     // idCamp -> ids contraintes
	DocumentPersonnes   map[int64]DocumentPersonne    // idDocument->
	DocumentAides       map[int64]DocumentAide        // idDocument ->
	DocumentCamps       map[int64]DocumentCamp        // idDocument ->
	MessageDocuments    map[int64]MessageDocument     // idMessage ->
	MessageSondages     map[int64]MessageSondage      // idMessage ->
	MessagePlaceliberes map[int64]MessagePlacelibere  // idMessage ->
	MessageAttestations map[int64]MessageAttestation  // idMessage ->
	MessageMessages     map[int64]MessageMessage      // idMessage ->
	MessageViews        map[int64]MessageViews        // idMessage ->
	DonDonateurs        map[int64]DonDonateur         // idDon ->
}

type TargetDocument interface {
	UpdateLinks(links *Links)
}

func (dp DocumentPersonne) UpdateLinks(links *Links) {
	links.DocumentPersonnes[dp.IdDocument] = dp
}

func (da DocumentAide) UpdateLinks(links *Links) {
	links.DocumentAides[da.IdDocument] = da
}

// ParProprietaire attribue chaque document à son propriétaire.
// Seuls les documents associés à des personnes sont considérés.
func (d Documents) ParProprietaire(targets DocumentPersonnes) map[int64][]Document {
	mapDocs := map[int64][]Document{}
	for _, target := range targets {
		mapDocs[target.IdPersonne] = append(mapDocs[target.IdPersonne], d[target.IdDocument])
	}
	return mapDocs
}

// AsIds renvoie les ids des contraintes
func (eqcts EquipierContraintes) AsIds() Ids {
	var out Ids
	for _, eqc := range eqcts {
		out = append(out, eqc.IdContrainte)
	}
	return out
}

type (
	LiensCampParticipants    map[int64][]int64
	LiensFactureParticipants map[int64][]int64
)

// Resoud parcourt la table des participants et renvois
// les associations facture -> participants, camp -> inscrits;
func (pts Participants) Resoud() (LiensFactureParticipants, LiensCampParticipants) {
	factureToParticipants := make(LiensFactureParticipants)
	campsToParticipants := make(LiensCampParticipants)
	for id, participant := range pts {
		idFac := participant.IdFacture
		if idFac.Valid {
			factureToParticipants[idFac.Int64] = append(factureToParticipants[idFac.Int64], id)
		}
		idCamp := participant.IdCamp
		campsToParticipants[idCamp] = append(campsToParticipants[idCamp], id)
	}
	return factureToParticipants, campsToParticipants
}

func (dest DestinatairesOptionnels) Index(i int) (Destinataire, error) {
	if i >= len(dest) {
		return Destinataire{}, fmt.Errorf("Index de destinataire %d invalide", i)
	}
	return dest[i], nil
}

func (eqs Equipiers) FindDirecteur() (Equipier, bool) {
	for _, part := range eqs {
		if part.Roles.Is(RDirecteur) {
			return part, true
		}
	}
	return Equipier{}, false
}

func (eqs Equipiers) FindAjoints() (adjoints []Equipier) {
	for _, eq := range eqs {
		if eq.Roles.Is(RAdjoint) {
			adjoints = append(adjoints, eq)
		}
	}
	return adjoints
}

func (eqs Equipiers) FindAuPairs() (animateurs []Equipier) {
	for _, eq := range eqs {
		if eq.Roles.IsAuPair() {
			animateurs = append(animateurs, eq)
		}
	}
	return animateurs
}

// FindLettre renvoie le premier lien vers une lettre, s'il existe
func (docs DocumentCamps) FindLettre() (DocumentCamp, bool) {
	for _, doc := range docs {
		if doc.IsLettre {
			return doc, true
		}
	}
	return DocumentCamp{}, false
}

// IId renvois un IdPersonne ou un IdOrganisme ou nil
func (d DonDonateur) IId() IId {
	if d.IdPersonne.Valid {
		return IdPersonne(d.IdPersonne.Int64)
	} else if d.IdOrganisme.Valid {
		return IdOrganisme(d.IdOrganisme.Int64)
	} else {
		return nil
	}
}

// Filepath renvoie le chemin du contenu du document.
func (doc Document) Filepath(dir string) string {
	ext := strings.ToLower(filepath.Ext(string(doc.NomClient)))
	name := fmt.Sprintf("doc_%d%s", doc.Id, ext)
	return filepath.Join(dir, name)
}

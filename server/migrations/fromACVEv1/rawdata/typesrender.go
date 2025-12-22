package rawdata

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	reSepTel    = regexp.MustCompile("[ -/;\t]")
	reSepPrenom = regexp.MustCompile("[ -.]")

	stringToMonth = map[string]time.Month{}
)

const (
	origineCotisation = 2015
)

const (
	VQuantite Field = iota
	VDescription
	VArrivee
	VDepart
)

const (
	PPSPrix Field = iota
	PPSStatut
	PPSDescription
)

const (
	DNomPrenom Field = iota
)

func init() {
	for m, s := range Months {
		stringToMonth[s] = time.Month(m + 1)
	}
}

func (s String) ToLower() string {
	return strings.ToLower(string(s))
}

func (s String) ToUpper() string {
	return strings.ToUpper(string(s))
}

func (s String) ToTitle() string {
	return strings.Title(string(s))
}

func (s String) TrimSpace() String {
	return String(strings.TrimSpace(s.String()))
}

func (s String) String() string {
	return string(s)
}

func (s String) Sortable() string {
	return strings.TrimSpace(s.ToLower())
}

func (f Float) Sortable() string {
	return fmt.Sprintf(". %30.5f", f)
}

func (d Date) String() string {
	da := d.Time()
	if da.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d", da.Day(), da.Month(), da.Year())
}

func (d Date) Sortable() string {
	return d.Time().String()
}

var weekDays = [7]string{
	"Dim",
	"Lun",
	"Mar",
	"Mer",
	"Jeu",
	"Ven",
	"Sam",
}

// shortDesc renvoie la date au format mer 2
func (d Date) shortDesc() string {
	t := d.Time()
	return fmt.Sprintf("%s %d", weekDays[t.Weekday()], t.Day())
}

func (d Time) String() string {
	da := d.Time().Local()
	if da.IsZero() {
		return ""
	}
	month := Months[int(da.Month()-1)]
	return fmt.Sprintf("%d %s %04d à %02dh%02d", da.Day(), month, da.Year(), da.Hour(), da.Minute())
}

func (d Time) Sortable() string {
	return d.Time().String()
}

func (p Plage) String() string {
	return fmt.Sprintf("du %s au %s", p.From, p.To)
}

func (p Plage) Sortable() string {
	return p.From.Sortable() + p.To.Sortable()
}

func (i Int) String() string {
	return strconv.Itoa(int(i))
}

func (i Int) Sortable() string {
	return fmt.Sprintf("%020d", i)
}

func (b Bool) String() String {
	if b {
		return "Oui"
	} else {
		return "Non"
	}
}

func (s Sexe) String() String {
	return String(SexeLabels[s])
}

func (s Sexe) Accord() string {
	if s == "F" {
		return "e"
	}
	return ""
}

func FormatPrenom(s String) string {
	parts := reSepPrenom.Split(string(s), -1)
	var tmp []string
	for _, p := range parts {
		if p != "" {
			tmp = append(tmp, strings.Title(strings.ToLower(p)))
		}
	}
	return strings.Join(tmp, "-")
}

func FormatSecuriteSocial(secu String) String {
	t := string(secu)
	t = strings.Replace(t, " ", "", -1)
	if len(t) < 15 {
		t = t + strings.Join(make([]string, 15+1-len(t)), " ")
	}
	t = t[0:1] + " " + t[1:3] + " " + t[3:5] + " " + t[5:7] + " " + t[7:10] + " " + t[10:13] + " " + t[13:15]
	return String(strings.TrimSpace(t))
}

// CondenseTel renvoie le numéro sans espaces ou séparteurs
func CondenseTel(t string) string {
	return reSepTel.ReplaceAllString(t, "")
}

func FormatTel(t string) string {
	return FormatTelSep(t, " ")
}

func FormatTelSep(t, separator string) string {
	t = CondenseTel(t)
	if len(t) < 8 {
		return t
	}
	start := len(t) - 8
	chunks := []string{t[:start]}
	for i := 0; i < 4; i++ {
		chunks = append(chunks, t[start+2*i:start+2*i+2])
	}
	return strings.Join(chunks, separator)
}

func renderTels(t Tels, sepInner, sepOuter string) String {
	fmted := make([]string, len(t))
	for index, tel := range t {
		fmted[index] = FormatTelSep(tel, sepInner)
	}
	return String(strings.Join(fmted, sepOuter))
}

func (t Tels) Sortable() string {
	reduced := make([]string, len(t))
	for index, tel := range t {
		reduced[index] = CondenseTel(tel)
	}
	sort.Strings(reduced)
	return strings.Join(reduced, "")
}

func (t Tels) String() string {
	return renderTels(t, "-", ";").String()
}

// StringLines renvoie une chaine sur plusieurs lignes
func (t Tels) StringLines() String {
	return renderTels(t, " ", "\n")
}

// StringLines renvoie une chaine sur plusieurs lignes, au format HTML
func (t Tels) StringHTML() String {
	return renderTels(t, " ", "<br/>")
}

func (r RangMembreAsso) String() string {
	return RangMembreAssoLabels[r]
}

func (r RangMembreAsso) Sortable() string {
	return string(r)
}

func (p Pays) String() string {
	if s, ok := PaysMap[p]; ok {
		return s
	}
	return string(p)
}

func (p Pays) Sortable() string { return string(p) }

func (r Departement) String() string {
	return string(r)
}

func (r Departement) Sortable() string {
	return string(r)
}

func (p Euros) String() string {
	return strings.Replace(fmt.Sprintf("%.2f €", p.Round()), ".", ",", -1)
}

func (p Euros) Sortable() string {
	return Float(p).Sortable()
}

func (p Pourcent) String() string {
	return fmt.Sprintf("%d %%", p)
}

func (p Pourcent) Sortable() string {
	return Int(p).Sortable()
}

func (t Taille) String() string {
	if t > 10000000 {
		return fmt.Sprintf("%s MB", Int(t/1000000).String())
	}
	return fmt.Sprintf("%s KB", Int(t/1000).String())
}

func (t Taille) Sortable() string {
	return Int(t).Sortable()
}

func (m Montant) String() string {
	var pJour string
	if m.parJour {
		pJour = " /jour"
	}
	return m.valeur.String() + pJour
}

func (m Montant) Sortable() string {
	return fmt.Sprintf("%v %s", m.parJour, m.valeur.Sortable())
}

// Description renvoie une liste de chaine au format HTML.
func (rem Remises) Description(sepTab string) []string {
	var s []string
	if rem.ReducEnfants > 0 {
		s = append(s, fmt.Sprintf("<i>Remise nombre d'enfants :</i> %s %s", sepTab,
			rem.ReducEnfants))
	}
	if rem.ReducEquipiers > 0 {
		s = append(s, fmt.Sprintf("<i>Remise équipiers :</i> %s %s", sepTab,
			rem.ReducEquipiers))
	}
	if rem.ReducSpeciale > 0 {
		s = append(s, fmt.Sprintf("<i>Remise spéciale :</i> %s %s", sepTab,
			rem.ReducSpeciale))
	}
	return s
}

func (m ModePaiment) String() String {
	return String(ModePaimentLabels[m])
}

func (o OptionPrixCamp) String() string {
	switch o.Active {
	case OptionsPrix.JOUR:
		return "Prix au nombre de jours"
	case OptionsPrix.SEMAINE:
		return "Prix à la semaine"
	case OptionsPrix.STATUT:
		return fmt.Sprintf("Prix par statut (%d)", len(o.Statut))
	}
	return ""
}

func (e Envois) String() string {
	var docs []string
	if e.LettreDirecteur {
		docs = append(docs, "lettre aux parents")
	}
	if e.ListeParticipants {
		docs = append(docs, "liste des participants")
	}
	if e.ListeVetements {
		docs = append(docs, "liste de vêtements")
	}
	return strings.Join(docs, ", ")
}

// Description renvoi du html
func (e Envois) Description() string {
	s := e.String()
	if e.Locked {
		return fmt.Sprintf("<b>Verrouillé</b> (%s)", s)
	}
	return s
}

type InfoEnvoi struct {
	Label string
	Time  Time
}

// Description renvoie la liste des mails envoyés, triée par date.
func (e EtatNegociation) Description(camps Camps) []InfoEnvoi {
	tmp := []InfoEnvoi{
		{"Accusé de réception", e.AccuseReception},
		{"Facture", e.Facture},
		{"Facture acquittée", e.FactureAcquittee},
		{"Attestation de présence", e.AttestationPresence},
	}
	for idCamp, t := range e.Documents {
		camp := camps[idCamp]
		tmp = append(tmp, InfoEnvoi{fmt.Sprintf("Documents du %s", camp.Label()), Time(t)})
	}
	sort.Slice(tmp, func(i, j int) bool { return tmp[j].Time.Time().Before(tmp[i].Time.Time()) })
	return tmp
}

func (r Role) String() string {
	return RoleLabels[r]
}

func (r Role) Sortable() string {
	s := r.String()
	if r == RDirecteur { // on s'assure que le directeur soit affiché en premier
		s = "AA-" + s
	}
	return s
}

func (d Diplome) String() String {
	return String(DiplomeLabels[d])
}

func (a Approfondissement) String() String {
	return String(ApprofondissementLabels[a])
}

func (b Bus) String() String {
	return String(BusLabels[b])
}

func (m MaterielSki) String() String {
	if m.Need == "" {
		return MaterielSkiNeed[m.Need]
	}
	var casque string
	if m.Casque {
		casque = "casque"
	}

	return String(fmt.Sprintf("%s - %s - %d kg - %d cm - %d (%s)",
		MaterielSkiModes[m.Mode],
		casque, m.Poids, m.Taille, m.Pointure,
		MaterielSkiNeed[m.Need],
	))
}

func (v Vetement) AsItem() Item {
	desc := v.Description
	if v.Obligatoire {
		desc = "(Important) " + desc
	}
	fields := F{
		VQuantite:    Int(v.Quantite),
		VDescription: String(desc),
	}
	isBold := v.Obligatoire
	bolds := B{VQuantite: isBold, VDescription: isBold}
	return Item{Fields: fields, Bolds: bolds}
}

func (c Cotisation) sorted() []int {
	sorted := make(sort.IntSlice, len(c))
	for i, an := range c {
		sorted[i] = int(an)
	}
	sort.Sort(sort.Reverse(sorted))
	return sorted
}

// String décrit l'état de la cotisation.
func (c Cotisation) String() string {
	ok := c.Bilan()
	if ok {
		return "Ok"
	}
	sorted := c.sorted()
	chunks := make([]string, len(sorted))
	for i, an := range sorted {
		chunks[i] = strconv.Itoa(an)
	}
	return strings.Join(chunks, ", ")
}

func (c Cotisation) Sortable() string {
	return fmt.Sprintf("%v", c.sorted())
}

func (p PrixParStatut) AsItem() Item {
	fields := F{
		PPSPrix:        p.Prix,
		PPSStatut:      p.Statut,
		PPSDescription: p.Description,
	}
	return Item{Id: Id(p.Id), Fields: fields, Bolds: B{PPSPrix: true}}
}

func (c Completion) String() string {
	return CompletionLabels[c]
}

func (c Completion) Sortable() string {
	return strconv.Itoa(int(c))
}

func (c Completion) Color() HexColor {
	switch c {
	case Complete:
		return "#00BB00"
	case EnCours:
		return "#eb7a34"
	case NonCommencee:
		return "#ff1e00"
	default:
		return "#55999999"
	}
}

func (p Destinataire) AsItem() Item {
	fields := F{
		DNomPrenom: p.NomPrenom,
	}
	return Item{Fields: fields}
}

func (sa StatutAttente) String() string {
	return StatutAttenteLabels[sa]
}

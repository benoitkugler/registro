package rawdata

// Data permet d'unifier toutes les données de sorties (listes, fichiers, etc...)
// Une valeur `nil` est tout à fait  possible et doit être gérée par le 'consomateur'
type Data interface {
	String() string

	// Sortable renvoie une chaîne qui préserve l'ordre
	Sortable() string
}

// Colors indique la couleur de chaque champ
type Colors interface {
	byField(field Field) Color
}

type ConstantColor struct {
	Color
}

func (c ConstantColor) byField(field Field) Color {
	return c.Color
}

type MapColors map[Field]Color

func (m MapColors) byField(field Field) Color {
	return m[field]
}

type Field uint8

type Header struct {
	Field Field
	Label string
}

type B = map[Field]bool

// Fields protège l'accès aux champs vides
type F map[Field]Data

// Data renvoie la donnée, qui est assurée de ne pas être `nil`
func (fs F) Data(field Field) Data {
	d := fs[field]
	if d == nil {
		return String("")
	}
	return d
}

// IId permet de différencier des éléments de plusieurs table
type IId interface {
	Int64() int64
}

// Id implémente IId, et suffit dans la plupart des cas.
type Id int64

func (i Id) Int64() int64 { return int64(i) }

// IdFacture représente un élement de la table Facture
type IdFacture int64

func (i IdFacture) Int64() int64 { return int64(i) }

// IdParticipant représente un élement de la table Participant
type IdParticipant int64

func (i IdParticipant) Int64() int64 { return int64(i) }

// IdParticipantsimple représente un élement de la table Participantsimple
type IdParticipantsimple int64

func (i IdParticipantsimple) Int64() int64 { return int64(i) }

// IdPersonne représente un élement de la table Personne
type IdPersonne int64

func (i IdPersonne) Int64() int64 { return int64(i) }

// IdOrganisme représente un élement de la table Organisme
type IdOrganisme int64

func (i IdOrganisme) Int64() int64 { return int64(i) }

// Item représente un élément générique, identifiable
// et statique, qui définit sa mise en forme
type Item struct {
	Id               IId
	Fields           F
	Bolds            B
	TextColors       Colors
	BackgroundColors Colors
}

// TextColor renvoie la couleur du texte, qui est assurée de ne pas être `nil`
func (i Item) TextColor(field Field) Color {
	var color Color
	if i.TextColors != nil {
		color = i.TextColors.byField(field)
	}
	return defaultC(color)
}

// BackgroundColor renvoie la couleur d'arrière plan, qui est assurée de ne pas être `nil`
func (i Item) BackgroundColor(field Field) Color {
	var color Color
	if i.BackgroundColors != nil {
		color = i.BackgroundColors.byField(field)
	}
	return defaultC(color)
}

type Table = []Item

// ItemChilds est un arbre à un niveau
type ItemChilds struct {
	Item
	Childs []Item
}

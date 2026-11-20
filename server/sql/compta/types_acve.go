package compta

// AccountACVE maps to one of the available account names
type AccountACVE uint8

const (
	Empty AccountACVE = iota // Aucun

	MaisonFamiliale // Maison Familiale
	Clairiere       // La Clairiere
	Rocher          // Le Rocher
	Montagne        // La Montagne
	Bergerie        // La Bergerie
	Ruisseau        // Le Ruisseau
	Glandasse       // Le Glandasse
	Vercors         // Le Vercors
	Source          // La Source

	Principal      // Compte principal
	CentreVacances // Centre de vacances
	Aurore         // Aurore
)

// FraisMDRACVE records the indirect expenses made
// by using the material from MdR
type FraisMDRACVE struct {
	Bus []Bus
	// PhotocopiesStart, PhotocopiesEnd int
}

type Bus struct {
	Name    string
	Kind    CarKind
	KmStart int
	KmEnd   int
}

type AccountKind uint8

const (
	AccountCamp AccountKind = iota // Compte dédié au séjour
	AccountAsso                    // Compte de l'asso.
)

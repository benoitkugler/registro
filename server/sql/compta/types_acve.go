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

var accountLabels = [Aurore + 1]string{
	Empty: "Aucun",

	MaisonFamiliale: "Maison Familiale",
	Clairiere:       "La Clairiere",
	Rocher:          "Le Rocher",
	Montagne:        "La Montagne",
	Bergerie:        "La Bergerie",
	Ruisseau:        "Le Ruisseau",
	Glandasse:       "Le Glandasse",
	Vercors:         "Le Vercors",
	Source:          "La Source",

	Principal:      "Compte principal",
	CentreVacances: "Centre de vacances",
	Aurore:         "Aurore",
}

var accountColors = [Aurore + 1]string{
	Empty:           "",
	MaisonFamiliale: "#c5e0b4",
	Clairiere:       "#ff6d6d",
	Rocher:          "#e575ff",
	Montagne:        "#ffcc00",
	Bergerie:        "#ffff00",
	Ruisseau:        "",
	Glandasse:       "#ffff99",
	Vercors:         "#66ff66",
	Source:          "#00ffff",
	Principal:       "",
	CentreVacances:  "",
	Aurore:          "",
}

type CarKind uint8

const (
	CarBus     CarKind = iota // Bus
	CarMinibus                // Minibus
	CarCar                    // Véhicule léger
)

const nbCarKind = CarCar + 1

// CarsCosts stores the kilometric cost for each kind of cars,
// as Euros cents
type CarsCosts [nbCarKind]int

type CarsKms [nbCarKind]int32

// FraismdrACVE records the indirect expenses made
// by using the material from MdR
type FraismdrACVE struct {
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
	AccountCamp AccountKind = iota // the account chosen for the camp (see AccountACVE)
	AccountAsso                    // the main asso account
)

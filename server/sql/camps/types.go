package camps

type Vetement struct {
	Quantite    int
	Description string
	Important   bool
}

type ListeVetements struct {
	Vetements []Vetement
	// HTML code inserted at the end of the list
	Complement string
}

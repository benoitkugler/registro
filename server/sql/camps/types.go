package camps

import "registro/sql/personnes"

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

// Satisfaction est une énumération indiquant le
// niveau de satisfaction sur le sondage de fin de séjour
type Satisfaction uint8

type ReponseSondage struct {
	InfosAvantSejour   Satisfaction 
	InfosPendantSejour Satisfaction 
	Hebergement        Satisfaction 
	Activites          Satisfaction 
	Theme              Satisfaction 
	Nourriture         Satisfaction 
	Hygiene            Satisfaction 
	Ambiance           Satisfaction 
	Ressenti           Satisfaction 
	MessageEnfant      string     
	MessageResponsable string       
}

type ParticipantExt struct {
	Participant
	Camp
	personnes.Personne
}

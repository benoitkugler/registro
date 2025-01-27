package dossier 

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

// Mode de paiement
type ModePaiement uint8 

const (
	_  ModePaiement = iota 
	EnLigne // (carte bancaire, en ligne)
	Cheque
	Virement 
	Especes 
	Ancv // TODO: utile ?
	Helloasso // 
)

type ParticipantExt struct {
	Participant
	Camp
	personnes.Personne
}

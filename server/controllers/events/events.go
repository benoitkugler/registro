package events

type Event struct {
	Created time.Time 
	Content EventContent
}

// EventContent exposes on event on the dossier track: either a dossier.Paiement or an event.Event
type EventContent interface{
	isEventContent()
}

type Supprime struct{}
type Inscription struct{}
type AccuseReception struct{}
type Message struct{}
type Facture  struct{}      
type CampDocs   struct{}     
type Attestation   struct{} 
type PlaceLiberee struct{}
type Sondage    struct{}
type Paiement struct{}


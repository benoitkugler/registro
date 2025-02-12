package events

// Event exposes on event on the dossier track
type Event struct {
	Event events.Event
	Content Content
}

// Event exposes on event on the dossier track
type Content interface{
	isContent()
}

func (Supprime) isContent() {}
func (Inscription) isContent() {}
func (AccuseReception) isContent() {}
func (Message) isContent() {}
func (Facture) isContent() {}
func (CampDocs) isContent() {}
func (PlaceLiberee) isContent() {}
func (Attestation) isContent() {}
func (Sondage) isContent() {}


type Supprime struct{}
type Inscription struct{}
type AccuseReception struct{}

type Message struct{
	Message events.EventMessage
	OrigineCampLabel string // optionnel 
	VuParCamps []string // labels
} 

type Facture  struct{}      

type CampDocs   struct{
	CampLabel string
}     

type PlaceLiberee struct{
	ParticpantLabel string 
	CampLabel string 
}

type Attestation struct{
	Distribution events.Distribution
	// IsPresence is true for 'Attestation de présence',
	// false for 'Facture acquittée'.
	IsPresence bool
} 

type Sondage struct{
	CampLabel string
}


func LoadEvents(db events.DB, idDossier ds.IdDossier) ([]Event, error) {
	raws, err := events.SelectEventsByIdDossiers(db, idDossier)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	ids := raw.IDs()

	tmp1, err := events.SelectEventMessagesByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	messages := tmp1.ByIdEvent()
	tmp1bis, err := events.SelectEventMessageVusByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	vupars := tmp1bis.ByIdEvent()
	tmp2, err := events.SelectEventCampDocsByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	campDocs := tmp2.ByIdEvent()
	tmp3, err := events.SelectEventPlaceLibereesByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	placeLiberees := tmp3.ByIdEvent()
	tmp4, err := events.SelectEventAttestationsByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	attestations := tmp4.ByIdEvent()
	tmp5, err := events.SelectEventSondagesByIdEvents(db, ids)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	sondages := tmp5.ByIdEvent()

	idCamps := append(append(tmp1bis.IdCamps(), tmp5.IdCamps()...), tmp1.IdCamps()...)
	camps, err := cps.SelectCamps(db, idCamps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	events := make([]Event, len(raw))
	for i, event := range raws {
		switch event.Kind {
		case Supprime:
			raws[i] = Event{event, Supprime{}}
		case Inscription:
			raws[i] = Event{event, Inscription{}}
		case AccuseReception:
			raws[i] = Event{event, AccuseReception{}}
		case Message:
			raws[i] = Event{event, Message{
				Message: messages[event.Id],
				OrigineCampLabel string // optionnel 
				VuParCamps []string // labels
			}}
		case Facture:
			raws[i] = Event{event, Facture{}}
		case CampDocs:
			raws[i] = Event{event, CampDocs{

			}}
		case PlaceLiberee:
			raws[i] = Event{event, PlaceLiberee{

			}}
		case Attestation:
			raws[i] = Event{event, Attestation{

			}}
		case Sondage:
			raws[i] = Event{event, Sondage{

			}}
		}
	}
}
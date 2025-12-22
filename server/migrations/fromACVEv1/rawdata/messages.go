package rawdata

// MessageKind détermine la catégorie de l'échange
type MessageKind uint8

var colors = [...]HexColor{
	MSupprime:            "#99808080",
	MResponsable:         "#993cff3c",
	MCentre:              "#9926c626",
	MAccuseReception:     "#994848c1",
	MFacture:             "#99ea8a1f",
	MDocuments:           "#9933def5",
	MFactureAcquittee:    "#99ead34c",
	MAttestationPresence: "#99ead34c",
	MSondage:             "#99dc24ca",
	MInscription:         "#996565ee",
	MPlaceLiberee:        "#99f9ff1c",
	MPaiement:            "#99ad9726",
}

func (m MessageKind) String() string {
	if int(m) < len(MessageKindLabels) {
		return MessageKindLabels[m]
	}
	return "Catégorie inconnue"
}

func (m MessageKind) Color() HexColor {
	if int(m) < len(colors) {
		return colors[m]
	}
	return ""
}

func (m MessageKind) MailTitle() string {
	switch m {
	case MCentre, MAccuseReception, MFacture, MDocuments, MFactureAcquittee, MAttestationPresence, MSondage, MPlaceLiberee:
		return m.String()
	default:
		return ""
	}
}

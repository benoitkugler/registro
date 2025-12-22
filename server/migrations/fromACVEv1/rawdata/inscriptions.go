package rawdata

const (
	IDateHeure Field = iota
	IResponsable
	INbParticipants
)

type ResponsableLegal struct {
	Lienid        IdentificationId `json:"lienid,omitempty"`
	Nom           String           `json:"nom,omitempty"`
	Prenom        String           `json:"prenom,omitempty"`
	Sexe          Sexe             `json:"sexe,omitempty"`
	Mail          String           `json:"mail,omitempty"`
	Adresse       String           `json:"adresse,omitempty"`
	CodePostal    String           `json:"code_postal,omitempty"`
	Ville         String           `json:"ville,omitempty"`
	Tels          Tels             `json:"tels,omitempty"`
	DateNaissance Date             `json:"date_naissance,omitempty"`
	Pays          Pays             `json:"pays,omitempty"`
}

type ParticipantInscription struct {
	Lienid           IdentificationId      `json:"lienid"`
	Nom              String                `json:"nom"`
	Prenom           String                `json:"prenom"`
	DateNaissance    Date                  `json:"date_naissance"`
	Sexe             Sexe                  `json:"sexe"`
	IdCamp           int64                 `json:"id_camp"`
	Options          OptionsParticipant    `json:"options"`
	OptionsPrix      OptionPrixParticipant `json:"options_prix"`
	QuotientFamilial Int                   `json:"quotient_familial"`
}

type ParticipantInscriptions []ParticipantInscription

func (r ResponsableLegal) ToPersonne() Personne {
	return Personne{BasePersonne: BasePersonne{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		Sexe:          r.Sexe,
		Mail:          r.Mail,
		Adresse:       r.Adresse,
		CodePostal:    r.CodePostal,
		Ville:         r.Ville,
		Tels:          r.Tels,
		DateNaissance: r.DateNaissance,
		Pays:          r.Pays,
	}}
}

// ToInscription renvoie les champs de la personne
// vus comme le responsable d'une inscription
func (r Personne) ToInscription() ResponsableLegal {
	return ResponsableLegal{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		Sexe:          r.Sexe,
		Mail:          r.Mail,
		Adresse:       r.Adresse,
		CodePostal:    r.CodePostal,
		Ville:         r.Ville,
		Tels:          r.Tels,
		DateNaissance: r.DateNaissance,
		Pays:          r.Pays,
	}
}

func (r ParticipantInscription) ToPersonne() Personne {
	return Personne{BasePersonne: BasePersonne{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		Sexe:          r.Sexe,
		DateNaissance: r.DateNaissance,
	}}
}

// ToParticipantInscription renvoie la personne comme un
// participant d'une inscription
func (r Personne) ToParticipantInscription() ParticipantInscription {
	return ParticipantInscription{
		Nom:           r.Nom,
		Prenom:        r.Prenom,
		Sexe:          r.Sexe,
		DateNaissance: r.DateNaissance,
	}
}

func (r Inscription) AsItem() Item {
	fields := F{IDateHeure: r.DateHeure, IResponsable: r.Responsable.ToPersonne().NomPrenom(), INbParticipants: Int(len(r.Participants))}
	return Item{Id: Id(r.Id), Fields: fields}
}

package camps

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"database/sql"
	"time"

	"registro/sql/dossiers"
	pr "registro/sql/personnes"
	sh "registro/sql/shared"
)

type (
	IdCamp          int64
	IdLettreImage   int64
	IdGroupe        int64
	IdParticipant   int64
	IdSondage       int64
	IdEquipier      int64
	IdStructureaide int64
	IdAide          int64
)

// Camp
//
// Requise par la contrainte Participant
// gomacro:SQL ADD UNIQUE(Id, IdTaux)
type Camp struct {
	Id IdCamp

	IdTaux dossiers.IdTaux

	Nom         string
	DateDebut   sh.Date
	Duree       int // nombre de jours date et fin inclus
	Lieu        string
	Agrement    string
	Description string // Description est affichée sur le formulaire d'inscription
	Navette     OptionNavette

	// Places est le nombre de places prévues pour le séjour
	// Les participants sont initialement passés en liste d'attente
	// si ce seuil est dépassé.
	Places          int
	AgeMin          int  // inclusif
	AgeMax          int  // inclusif
	NeedEquilibreGF bool // si 'true', prend en compte le nombre de garçons et filles

	Ouvert bool // ouvert aux inscriptions ou non

	Prix                   Montant
	OptionPrix             OptionPrixCamp
	OptionQuotientFamilial PrixQuotientFamilial

	Password string
}

// Lettredirecteur conserve le html utilisé pour générer la lettre.
// En revanche, c'est bien le document PDF généré et enregistré dans la
// table Files qui est envoyé aux parents.
//
// gomacro:SQL ADD UNIQUE(IdCamp)
type Lettredirecteur struct {
	IdCamp             IdCamp `gomacro-sql-on-delete:"CASCADE"`
	Html               string
	UseCoordCentre     bool
	ShowAdressePostale bool
	ColorCoord         string
}

// LettreImage stockes les images contenues dans les lettres aux parents,
// accessibles via un lien crypté
type LettreImage struct {
	Id       IdLettreImage
	IdCamp   int64  `gomacro-sql-on-delete:"CASCADE"`
	Filename string // as uploaded
	Content  []byte
}

// Participant
//
// gomacro:SQL ADD FOREIGN KEY (IdCamp, IdTaux) REFERENCES Camp (Id,IdTaux)
// gomacro:SQL ADD FOREIGN KEY (IdDossier, IdTaux) REFERENCES Dossier (Id,IdTaux) ON DELETE CASCADE
//
// Une même personne ne peut être présent qu'une seule fois dans un séjour
// gomacro:SQL ADD UNIQUE(IdCamp, IdPersonne)
//
// Requise par la contrainte GroupeParticipant
// gomacro:SQL ADD UNIQUE(Id, IdCamp)
//
// gomacro:QUERY SwitchParticipantPersonne UPDATE Participant SET IdPersonne = $target$ WHERE IdPersonne = $temporaire$;
// gomacro:QUERY SwitchParticipantDossier UPDATE Participant SET IdDossier = $to$ WHERE IdDossier = $from$;
type Participant struct {
	Id         IdParticipant
	IdCamp     IdCamp
	IdPersonne pr.IdPersonne
	IdDossier  dossiers.IdDossier `gomacro-sql-on-delete:"CASCADE"`

	// IdTaux is used for consistency,
	// so that a [Dossier] has only one taux
	IdTaux dossiers.IdTaux

	Statut           StatutParticipant
	Remises          Remises
	QuotientFamilial int // optional, 0 for inactive

	OptionPrix OptionPrixParticipant

	Details string  // rempli sur l'espace de suivi
	Navette Navette // rempli sur l'espace de suivi
}

// Groupe représente un groupe de participants
// Un séjour peut définir (ou non) une liste de groupes
//
// gomacro:SQL ADD UNIQUE(IdCamp, Nom)
// Requise par la contrainte GroupeParticipant
// gomacro:SQL ADD UNIQUE(Id, IdCamp)
type Groupe struct {
	Id     IdGroupe
	IdCamp IdCamp `gomacro-sql-on-delete:"CASCADE"`

	// TODO: check that
	// un nom vide indique un groupe par défaut
	Nom string
	// indication: ignorée forcément pour un groupe par défaut
	Plage sh.Plage
	// Hex color, optionnelle
	Couleur string
}

// GroupeParticipant défini le contenu des groupes
// gomacro:SQL ADD UNIQUE (IdParticipant)
// gomacro:SQL ADD UNIQUE (IdParticipant, IdCamp)
// gomacro:SQL ADD FOREIGN KEY (IdParticipant, IdCamp) REFERENCES Participant (Id,IdCamp) ON DELETE CASCADE
// gomacro:SQL ADD FOREIGN KEY (IdGroupe, IdCamp) REFERENCES Groupe (Id,IdCamp) ON DELETE CASCADE
type GroupeParticipant struct {
	IdParticipant IdParticipant `gomacro-sql-on-delete:"CASCADE"`
	IdGroupe      IdGroupe      `gomacro-sql-on-delete:"CASCADE"`
	// redondance pour assurer l'intégrité
	IdCamp IdCamp

	// Manuel indique si l'attribution a été faite
	// en modifiant directement la fiche du participant ou
	// en fonction de l'âge
	Manuel bool
}

// Sondage enregistre les retours sur un séjour
//
// gomacro:SQL ADD UNIQUE(IdCamp, IdDossier)
//
// gomacro:QUERY SwitchSondageDossier UPDATE Sondage SET IdDossier = $to$ WHERE IdDossier = $from$;
type Sondage struct {
	IdSondage IdSondage
	IdCamp    IdCamp             `gomacro-sql-on-delete:"CASCADE"`
	IdDossier dossiers.IdDossier `gomacro-sql-on-delete:"CASCADE"`
	Modified  time.Time

	ReponseSondage
}

type Structureaide struct {
	Id              IdStructureaide
	Nom             string
	Immatriculation string
	Adresse         string
	CodePostal      string
	Ville           string
	Telephone       pr.Tel
	Info            string
}

type Aide struct {
	Id              IdAide
	IdStructureaide IdStructureaide
	IdParticipant   IdParticipant `gomacro-sql-on-delete:"CASCADE"`

	Valide bool

	Valeur     Montant
	ParJour    bool
	NbJoursMax int
}

// ---------------------------- Equipiers ----------------------------

// Equipier représente un participant dans l'équipe d'un séjour
//
// gomacro:SQL ADD UNIQUE(IdCamp, IdPersonne)
// gomacro:SQL CREATE UNIQUE INDEX ON Equipier(IdCamp) WHERE #[Role.Direction] = ANY(Roles)
//
// gomacro:QUERY SwitchEquipierPersonne UPDATE Equipier SET IdPersonne = $target$ WHERE IdPersonne = $temporaire$;
type Equipier struct {
	Id         IdEquipier
	IdCamp     IdCamp        `gomacro-sql-on-delete:"CASCADE"`
	IdPersonne pr.IdPersonne `gomacro-sql-on-delete:"CASCADE"`

	Roles    Roles
	Presence PresenceOffsets

	FormStatus FormStatusEquipier

	// validation de la charte ACVE
	AccepteCharte sql.NullBool
}

package files

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro models.go go/sqlcrud:gen_scans.go sql:gen_create.sql go/randdata:gen_randdata_test.go

import (
	"fmt"
	"path/filepath"
	"time"

	cps "registro/sql/camps"
	"registro/sql/personnes"
)

type (
	IdFile    int64
	IdDemande int64
)

// File représente les méta données d'un document stocké sur le serveur
//
// Le contenu et la miniature sont stockés dans un dossier, pour ne pas alourdir la
// base de données.
type File struct {
	Id IdFile

	// En bytes
	Taille int
	// as provided by the client,
	// different from the file path
	NomClient      string
	DateHeureModif time.Time
}

func NewFile(fileContent []byte, filename string) File {
	return File{Taille: len(fileContent), NomClient: filename, DateHeureModif: time.Now()}
}

// Filepath returns the file location
func (f File) Filepath(dir string, forMiniature bool) string {
	id := fmt.Sprintf("file_%d", f.Id)
	if forMiniature {
		id += "_min"
	}
	return filepath.Join(dir, id)
}

// Demande encode la catégorie d'un fichier à fournir.
// On différencie deux types de catégories :
//   - les documents connus en avances (doc. équipiers, vaccins)
//   - des documents spécifiques à chaque camp et pouvant donc varier
//
// L'attribut 'Categorie' permet d'identifier des contraintes universelles
//
// gomacro:SQL ADD CONSTRAINT constraint_1 CHECK(Categorie = 0 OR IdDirecteur IS NULL)
// gomacro:SQL ADD CONSTRAINT constraint_2 CHECK(Categorie <> 0 OR IdDirecteur IS NOT NULL)
// gomacro:SQL ADD CONSTRAINT constraint_3 CHECK(MaxDocs >= 1)
// gomacro:SQL CREATE UNIQUE INDEX ON Demande(Categorie) WHERE Categorie <> 0
type Demande struct {
	Id IdDemande

	// Document à télécharger et remplir, optionnel
	IdFile OptIdFile `gomacro-sql-foreign:"File"`

	// Pour les demandes 'custom', le directeur proprietaire de la contrainte
	IdDirecteur OptIdPersonne `gomacro-sql-on-delete:"CASCADE" gomacro-sql-foreign:"Personne"`

	Categorie Categorie

	// Optionnelle, affichée sur l'espace perso
	Description string

	// Nombre max de documents qui peuvent satisfaire la contrainte
	// (1 par défaut)
	MaxDocs int

	// JoursValide, si > 0, indique un document temporaire :
	// une alerte est donnée pour les documents périmés
	JoursValide int
}

// DemandeEquipier représente un document demandé à un équpier
//
// gomacro:SQL ADD UNIQUE(IdEquipier, IdDemande)
type DemandeEquipier struct {
	IdEquipier cps.IdEquipier `gomacro-sql-on-delete:"CASCADE"`
	IdDemande  IdDemande
	Optionnel  bool
}

// DemandeCamp représente un document demandé
// à tous les participants
//
// gomacro:SQL ADD UNIQUE(IdCamp, IdDemande)
type DemandeCamp struct {
	IdCamp    cps.IdCamp `gomacro-sql-on-delete:"CASCADE"`
	IdDemande IdDemande
}

// FileCamp est une table de lien pour les lettres des séjours et les documents additionnels
//
// gomacro:SQL ADD UNIQUE(IdFile)
// gomacro:SQL CREATE UNIQUE INDEX ON FileCamp(IdCamp) WHERE IsLettre IS true
type FileCamp struct {
	IdFile   IdFile `gomacro-sql-on-delete:"CASCADE"`
	IdCamp   cps.IdCamp
	IsLettre bool // sinon, document additionnel
}

// FilePersonne est une table de lien pour les documents liés aux personnes.
//
// gomacro:SQL ADD UNIQUE(IdFile)
type FilePersonne struct {
	IdFile     IdFile `gomacro-sql-on-delete:"CASCADE"`
	IdPersonne personnes.IdPersonne
	IdDemande  IdDemande
}

// FileAide est une table de lien pour les justificatifs des aides
//
// gomacro:SQL ADD UNIQUE(IdFile)
// gomacro:SQL ADD UNIQUE(IdAide)
type FileAide struct {
	IdFile IdFile `gomacro-sql-on-delete:"CASCADE"`
	IdAide cps.IdAide
}

// Implémente des outils de recherche et de fusion de profil
package search

import (
	"strings"

	pr "registro/sql/personnes"
	"registro/utils"
)

// Ce fichier implémente un algorithme de comparaison
// champ par champ entre deux personnes (une entrante et une existante).
// Il est utilisé de façon automatique pour fusionner
// une inscription avec une personne déjà existante,
// et de façon semi-automatique sur le client pour rapprocher
// une personne temporaire d'une personne existante.

// TODO: cleanup ?

// IdentResult est une union décrivant le résultat
// d'une identification ("vers un nouveau profil" ou "rapprochement")
type IdentResult interface {
	isIdentResult()
}

type NouveauProfil struct{}

func (NouveauProfil) isIdentResult() {}

type Rattache struct {
	IdTarget      pr.IdPersonne // la personne à laquelle se rattacher
	Modifications pr.Etatcivil  // le résultat de la fusion, à appliquer à la cible
}

func (Rattache) isIdentResult() {}

// résultat d'une comparaison
type diff uint8

const (
	inZero    diff = iota // la valeur entrante est vide
	existZero             // la valeur existante est vide
	equal                 // les deux valeurs sont similaires
	conflict              // les deux valeurs sont vraiments différentes
)

func cmpString[T interface{ ~string }](s1, s2 T) diff {
	if s1 == "" {
		return inZero
	}
	if s2 == "" {
		return existZero
	}
	ss1 := strings.Replace(utils.Normalize(string(s1)), " ", "", -1)
	ss2 := strings.Replace(utils.Normalize(string(s2)), " ", "", -1)
	if ss1 == ss2 {
		return equal
	}
	return conflict
}

func cmpTels(t1, t2 pr.Tels) diff {
	if len(t1) == 0 {
		return inZero
	}
	if len(t2) == 0 {
		return existZero
	}
	uniques1, uniques2 := make(map[string]bool), make(map[string]bool)
	for _, t := range t1 {
		uniques1[pr.StripTel(t)] = true
	}

	for _, t := range t2 {
		t = pr.StripTel(t)
		if !uniques1[t] {
			return conflict
		}
		uniques2[t] = true
	}
	for t := range uniques1 {
		if !uniques2[t] {
			return conflict
		}
	}
	return equal
}

func cmpBool(in, out bool) diff {
	if in == out {
		return equal
	}
	return conflict
}

func cmpDate(in, out pr.Date) diff {
	in_, out_ := in.Time(), out.Time()
	if in_.IsZero() {
		return inZero
	}
	if out_.IsZero() {
		return existZero
	}
	if in_.Equal(out_) {
		return equal
	}
	return conflict
}

func cmpEnum[T interface {
	pr.Sexe | pr.Approfondissement | pr.Diplome
}](in, out T) diff {
	if in == 0 {
		return inZero
	}
	if out == 0 {
		return existZero
	}
	if in == out {
		return equal
	}
	return conflict
}

// Conflicts indique quels champs n'ont pu être automatiquement fusionnés
type Conflicts struct {
	Nom                  bool
	NomJeuneFille        bool
	Prenom               bool
	DateNaissance        bool
	VilleNaissance       bool
	DepartementNaissance bool
	Sexe                 bool
	Tels                 bool
	Mail                 bool
	Adresse              bool
	CodePostal           bool
	Ville                bool
	Pays                 bool
	SecuriteSociale      bool
	Profession           bool
	Etudiant             bool
	Fonctionnaire        bool
	Diplome              bool
	Approfondissement    bool
}

type fields interface {
	string | pr.Date | pr.Departement | pr.Sexe | pr.Tels | pr.Pays | bool | pr.Diplome | pr.Approfondissement
}

func cmpGeneric[T fields](entrant, existant T) diff {
	v1, v2 := any(entrant), any(existant)
	switch v1.(type) {
	case string:
		return cmpString(v1.(string), v2.(string))
	case pr.Date:
		return cmpDate(v1.(pr.Date), v2.(pr.Date))
	case pr.Departement:
		return cmpString(v1.(pr.Departement), v2.(pr.Departement))
	case pr.Sexe:
		return cmpEnum(v1.(pr.Sexe), v2.(pr.Sexe))
	case pr.Tels:
		return cmpTels(v1.(pr.Tels), v2.(pr.Tels))
	case pr.Pays:
		return cmpString(v1.(pr.Pays), v2.(pr.Pays))
	case bool:
		return cmpBool(v1.(bool), v2.(bool))
	case pr.Diplome:
		return cmpEnum(v1.(pr.Diplome), v2.(pr.Diplome))
	case pr.Approfondissement:
		return cmpEnum(v1.(pr.Approfondissement), v2.(pr.Approfondissement))
	default:
		panic("exhaustive type switch")
	}
}

func choose[T fields](entrant, existant T) (T, bool) {
	d := cmpGeneric(entrant, existant)
	switch d {
	case inZero: // on garde l'existant
		return existant, false
	case existZero: // on ecrase l'existant
		return entrant, false
	case equal: // on ecrase (c'est plus cohérent avec l'attente utilisateur)
		return entrant, false
	case conflict: // on écrase et on alerte
		return entrant, true
	default:
		panic("unknow diff type")
	}
}

// Merge compare champs par champs les deux personnes et renvoie
// le résultat de la fusion et un crible d'alerte
func Merge(entrant pr.Etatcivil, existant pr.Etatcivil) (merged pr.Etatcivil, conflicts Conflicts) {
	merged.Nom, conflicts.Nom = choose(entrant.Nom, existant.Nom)
	merged.NomJeuneFille, conflicts.NomJeuneFille = choose(entrant.NomJeuneFille, existant.NomJeuneFille)
	merged.Prenom, conflicts.Prenom = choose(entrant.Prenom, existant.Prenom)
	merged.DateNaissance, conflicts.DateNaissance = choose(entrant.DateNaissance, existant.DateNaissance)
	merged.VilleNaissance, conflicts.VilleNaissance = choose(entrant.VilleNaissance, existant.VilleNaissance)
	merged.DepartementNaissance, conflicts.DepartementNaissance = choose(entrant.DepartementNaissance, existant.DepartementNaissance)
	merged.Sexe, conflicts.Sexe = choose(entrant.Sexe, existant.Sexe)
	merged.Tels, conflicts.Tels = choose(entrant.Tels, existant.Tels)
	merged.Mail, conflicts.Mail = choose(entrant.Mail, existant.Mail)
	merged.Adresse, conflicts.Adresse = choose(entrant.Adresse, existant.Adresse)
	merged.CodePostal, conflicts.CodePostal = choose(entrant.CodePostal, existant.CodePostal)
	merged.Ville, conflicts.Ville = choose(entrant.Ville, existant.Ville)
	merged.Pays, conflicts.Pays = choose(entrant.Pays, existant.Pays)
	merged.SecuriteSociale, conflicts.SecuriteSociale = choose(entrant.SecuriteSociale, existant.SecuriteSociale)
	merged.Profession, conflicts.Profession = choose(entrant.Profession, existant.Profession)
	merged.Etudiant, conflicts.Etudiant = choose(entrant.Etudiant, existant.Etudiant)
	merged.Fonctionnaire, conflicts.Fonctionnaire = choose(entrant.Fonctionnaire, existant.Fonctionnaire)
	merged.Diplome, conflicts.Diplome = choose(entrant.Diplome, existant.Diplome)
	merged.Approfondissement, conflicts.Approfondissement = choose(entrant.Approfondissement, existant.Approfondissement)
	return merged, conflicts
}

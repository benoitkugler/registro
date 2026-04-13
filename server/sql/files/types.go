package files

import (
	"fmt"

	cp "registro/sql/camps"
	"registro/sql/shared"
	"registro/utils"
)

type OptIdFile = shared.OptID[IdFile]

type Categorie uint8

const (
	NoBuiltin Categorie = iota // -

	CarteId                 // Carte d'identité/Passeport
	Permis                  // Permis de conduire
	SB                      // Surveillant de baignade
	Secourisme              // Secourisme (PSC1 - AFPS)
	Bafa                    // BAFA
	Bafd                    // BAFD
	CarteVitale             // Carte Vitale
	Vaccins                 // Vaccins
	Haccp                   // Cuisine (HACCP)
	BafdEquiv               // Equivalent BAFD
	BafaEquiv               // Equivalent BAFA
	CertMedicalCuisine      // Certificat médical cuisine
	PhotoIdentite           // Photo d'identité
	ExtraitCasierJudiciaire // Extrait du casier judiciaire
	Autre                   // Autre

)

const nbCategorieEquipier = int(Autre) + 1

func (c Categorie) String() string {
	switch c {
	case CarteId:
		return "Carte d'identité/Passeport"
	case Permis:
		return "Permis de conduire"
	case SB:
		return "Surveillant de baignade"
	case Secourisme:
		return "Secourisme (PSC1 - AFPS)"
	case Bafa:
		return "BAFA"
	case Bafd:
		return "BAFD"
	case CarteVitale:
		return "Carte Vitale"
	case Vaccins:
		return "Vaccins"
	case Haccp:
		return "Cuisine (HACCP)"
	case BafdEquiv:
		return "Equivalent BAFD"
	case BafaEquiv:
		return "Equivalent BAFA"
	case CertMedicalCuisine:
		return "Certificat médical cuisine"
	case PhotoIdentite:
		return "Photo d'identité"
	case ExtraitCasierJudiciaire:
		return "Extrait du casier judiciaire"
	case Autre:
		return "Autre"
	default:
		return ""
	}
}

// Builtins indique les demandes connues à l'avance
type Builtins [nbCategorieEquipier]Demande

// LoadBuiltins charge les demandes 'builtin', qui doivent
// être prédéclarées dans la base de données.
func LoadBuiltins(db DB) (out Builtins, err error) {
	ds, err := SelectAllDemandes(db)
	if err != nil {
		return out, utils.SQLError(err)
	}
	return ds.builtins()
}

func (bs Builtins) List() []Demande {
	return bs[1:] // remove the NoBuiltin
}

func (ds Demandes) builtins() (out [nbCategorieEquipier]Demande, err error) {
	for _, demande := range ds {
		if demande.Categorie == NoBuiltin {
			continue
		}
		out[demande.Categorie] = demande
	}

	// check that all builtins are properly defined
	for cat, v := range out {
		if cat == 0 { //  ignore the empty categorie
			continue
		}
		if v.Id == 0 {
			return out, fmt.Errorf("table 'demandes': missing builtin Categorie %d (%s)", cat, Categorie(cat))
		}
	}
	return out, nil
}

// Defaut renvoie les demandes par défaut pour l'équipier donné.
func (builtinDemandes Builtins) Defaut(id cp.IdEquipier, roles cp.Roles, asso string) DemandeEquipiers {
	var demandes []optionalCategorie
	switch asso {
	case "acve":
		demandes = defaultACVE(roles)
	case "repere":
		demandes = defaultRepere(roles)
	default:
		panic("invalid asso ID: " + asso)
	}
	out := make(DemandeEquipiers, len(demandes))
	for i, d := range demandes {
		out[i] = DemandeEquipier{
			IdEquipier:  id,
			IdDemande:   builtinDemandes[d.Categorie].Id,
			Optionnelle: d.Optionnelle,
		}
	}
	return out
}

type optionalCategorie struct {
	Categorie   Categorie
	Optionnelle bool
}

var defaultDemandesACVE = [cp.NbRoles][]Categorie{
	cp.Direction:     {CarteId, Permis, SB, Bafa, Bafd, CarteVitale, Vaccins, BafdEquiv},
	cp.Adjoint:       {CarteId, Permis, SB, CarteVitale, Vaccins},
	cp.Animation:     {CarteId, Permis, SB, Bafa, BafaEquiv, CarteVitale, Vaccins},
	cp.Menage:        {CarteId, CarteVitale, Vaccins},
	cp.Cuisine:       {CarteId, CarteVitale, Vaccins, Haccp, CertMedicalCuisine},
	cp.Intendance:    {CarteId, CarteVitale, Vaccins, Haccp},
	cp.Infirmerie:    {CarteId, Secourisme, CarteVitale, Vaccins},
	cp.AideAnimation: {CarteId, CarteVitale, Vaccins},
	cp.Lingerie:      {CarteId, CarteVitale, Vaccins},
	cp.Chauffeur:     {CarteId, CarteVitale, Vaccins},
	cp.Factotum:      {CarteId, CarteVitale, Vaccins},
	cp.Babysiter:     {CarteId, CarteVitale, Vaccins},
	cp.AutreRole:     {}, // rien
}

func defaultACVE(roles cp.Roles) []optionalCategorie {
	// on aggrège les demandes de chaque rôle
	var categories [nbCategorieEquipier]bool
	for _, role := range roles {
		for _, cat := range defaultDemandesACVE[role] {
			categories[cat] = true
		}
	}
	// on prend en compte le caractère optionnelle
	var demandes []optionalCategorie
	for cat, has := range categories {
		if !has {
			continue
		}
		demandes = append(demandes, optionalCategorie{
			Categorie:   Categorie(cat),
			Optionnelle: isDemandeOptACVE(Categorie(cat), roles),
		})
	}
	return demandes
}

func isDemandeOptACVE(cat Categorie, roles cp.Roles) bool {
	// un certificat de secourisme est obligatoire pour l'infirmerie
	if roles.Is(cp.Infirmerie) && cat == Secourisme {
		return false
	}
	if roles.Is(cp.Cuisine) && cat == CertMedicalCuisine {
		return false
	}
	return true
}

func defaultRepere(_ cp.Roles) []optionalCategorie {
	return []optionalCategorie{
		{SB, true},
		{Secourisme, true},
		{Bafa, true},
		{Bafd, true},
		{BafaEquiv, true},
		{BafdEquiv, true},
		{ExtraitCasierJudiciaire, false},
	}
}

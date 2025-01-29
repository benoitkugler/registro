package files

import (
	"registro/sql/personnes"
	"registro/sql/shared"
)

type OptIdFile shared.OptID[IdFile]

type OptIdPersonne shared.OptID[personnes.IdPersonne]

type Categorie uint8

const (
	Unknown            Categorie = iota // -
	Bafa                                // BAFA
	Bafd                                // BAFD
	CarteId                             // Carte d'identité/Passeport
	Permis                              // Permis de conduire
	Sb                                  // Surveillant de baignade
	Secour                              // Secourisme (PSC1 - AFPS)
	BafdEquiv                           // Equivalent BAFD
	BafaEquiv                           // Equivalent BAFA
	CarteVitale                         // Carte Vitale
	Haccp                               // Cuisine (HACCP)
	CertMedicalCuisine                  // Certificat médical Cuisine
	Scolarite                           // Certificat de scolarité
	Autre                               // Autre
	Vaccin                              // Vaccin

	TestNautique // Test nautique
)

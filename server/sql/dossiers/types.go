package dossiers

import (
	"math"

	"registro/sql/shared"
)

type OptIdDossier = shared.OptID[IdDossier]

func (id IdDossier) Opt() OptIdDossier { return OptIdDossier{Id: id, Valid: true} }

// Mode de paiement
type ModePaiement uint8

const (
	Cheque   ModePaiement = iota // Chèque
	EnLigne                      // Carte bancaire (en ligne)
	Virement                     // Virement
	Especes                      // Espèces
	Ancv                         // ANCV
	// uniquement pour les dons
	Helloasso // Helloasso
)

func (mp ModePaiement) String() string {
	switch mp {
	case Cheque:
		return "Chèque"
	case EnLigne:
		return "Carte bancaire (en ligne)"
	case Virement:
		return "Virement"
	case Especes:
		return "Espèces"
	case Ancv:
		return "ANCV"
	case Helloasso:
		return "Helloasso"
	default:
		return "unknown ModePaiement"
	}
}

// Montant représente un prix (avec son unité).
type Montant struct {
	Cent     int
	Currency Currency
}

func NewEuros(f float64) Montant        { return Montant{int(math.Round(f * 100)), Euros} }
func NewFrancsuisses(f float64) Montant { return Montant{int(math.Round(f * 100)), FrancsSuisse} }

type Currency uint8

const (
	Euros        Currency = iota // €
	FrancsSuisse                 // CHF
)

const nbCurrencies = FrancsSuisse + 1 // gomacro:no-enum

func (c Currency) String() string {
	switch c {
	case Euros:
		return "€"
	case FrancsSuisse:
		return "CHF"
	default:
		return "<invalid currency>"
	}
}

const PayeurFondSoutien = "Fonds de soutien"

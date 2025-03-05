package dossiers

import (
	"math"
	"registro/sql/shared"
)

type OptIdTaux shared.OptID[IdTaux]

func (Id IdTaux) Opt() OptIdTaux { return OptIdTaux{Id: Id, Valid: true} }

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

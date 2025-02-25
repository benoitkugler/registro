package dossiers

import (
	"fmt"
	"strings"

	"registro/sql/shared"
)

type OptIdTaux shared.OptID[IdTaux]

func (Id IdTaux) Opt() OptIdTaux { return OptIdTaux{Id: Id, Valid: true} }

// Mode de paiement
type ModePaiement uint8

const (
	Cheque  ModePaiement = iota
	EnLigne              // (carte bancaire, en ligne)
	Virement
	Especes
	Ancv
	// uniquement pour les dons
	Helloasso
)

// Montant représente un prix (avec son unité).
type Montant struct {
	Cent     int
	Currency Currency
}

func NewEuros(f float32) Montant        { return Montant{int(f * 100), Euros} }
func NewFrancsuisses(f float32) Montant { return Montant{int(f * 100), FrancsSuisse} }

func (m Montant) String() string {
	val := strings.ReplaceAll(fmt.Sprintf("%g", float64(m.Cent)/100), ".", ",")
	switch m.Currency {
	case FrancsSuisse:
		return m.Currency.String() + " " + val
	default:
		return val + m.Currency.String()
	}
}

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

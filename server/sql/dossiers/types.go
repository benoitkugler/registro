package dossiers

import (
	"fmt"
	"strings"
)

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

func NewEuros(f float32) Montant { return Montant{int(f * 100), Euros} }

func (s Montant) String() string {
	return strings.ReplaceAll(fmt.Sprintf("%g %s", float64(s.Cent)/100, s.Currency), ".", ",")
}

// Add assume the currency is the same.
func (s *Montant) Add(other Montant) { s.Cent += other.Cent }

type Currency uint8

const (
	Empty Currency = iota
	Euros
	FrancsSuisse
)

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

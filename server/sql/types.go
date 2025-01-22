package sql

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type currency uint8

const (
	_ currency = iota
	euros
	francsSuisse
)

func (c currency) String() string {
	switch c {
	case euros:
		return "€"
	case francsSuisse:
		return "CHF"
	default:
		return "<invalid currency>"
	}
}

// Montant représente un prix (avec son unité).
//
// SQL: CREATE TYPE montant AS (cent int, currency smallint);
type Montant struct {
	cent     int
	currency currency
}

func Euros(f float32) Montant { return Montant{int(f * 100), euros} }

func (s Montant) String() string {
	return strings.ReplaceAll(fmt.Sprintf("%g %s", float64(s.cent)/100, s.currency), ".", ",")
}

func (s *Montant) Scan(src interface{}) error {
	bs, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type %T", src)
	}
	if len(bs) < 3 {
		return fmt.Errorf("invalid Montant %s", bs)
	}
	centV, currencyV, ok := strings.Cut(string(bs[1:len(bs)-1]), ",")
	if !ok {
		return fmt.Errorf("invalid Montant %s", bs)
	}
	var err error
	s.cent, err = strconv.Atoi(centV)
	if err != nil {
		return err
	}
	c, err := strconv.Atoi(currencyV)
	if err != nil {
		return err
	}
	if c > int(francsSuisse) {
		return fmt.Errorf("invalid Montant: invalid currency %d", c)
	}
	s.currency = currency(c)

	return nil
}

func (s Montant) Value() (driver.Value, error) {
	bs := fmt.Appendf(nil, "(%d,%d)", s.cent, s.currency)
	return driver.Value(bs), nil
}

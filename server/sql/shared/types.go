package shared

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Currency uint8

const (
	_ Currency = iota
	euros
	francsSuisse
)

func (c Currency) String() string {
	switch c {
	case euros:
		return "€"
	case francsSuisse:
		return "CHF"
	default:
		return "<invalid currency>"
	}
}

// Date is a date (without notion of time)
type Date time.Time

func NewDate(year int, month time.Month, day int) Date {
	return Date(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

func NewDateFrom(ti time.Time) Date {
	return NewDate(ti.Year(), ti.Month(), ti.Day())
}

func (d Date) Time() time.Time {
	ti := time.Time(d)
	return time.Date(ti.Year(), ti.Month(), ti.Day(), 0, 0, 0, 0, time.UTC)
}

func (d Date) String() string {
	da := d.Time()
	if da.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d", da.Day(), da.Month(), da.Year())
}

var (
	_ encoding.TextMarshaler   = (*Date)(nil)
	_ encoding.TextUnmarshaler = (*Date)(nil)
)

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.Time().Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	ti, err := time.Parse(time.DateOnly, string(text))
	*d = NewDateFrom(ti)
	return err
}

func (s *Date) Scan(src interface{}) error {
	var tmp pq.NullTime
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = NewDateFrom(tmp.Time)
	return nil
}

func (s Date) Value() (driver.Value, error) {
	return pq.NullTime{Time: s.Time(), Valid: true}.Value()
}

type Plage struct {
	From Date
	// Duree est le nombre de jour, début et fin inclus
	Duree int
}

func (pl Plage) To() Date {
	out := pl.From.Time()
	return NewDateFrom(out.Add(time.Hour * 24 * time.Duration(pl.Duree-1)))
}

type OptID[T ~int64] struct {
	Id    T
	Valid bool
}

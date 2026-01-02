package shared

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
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

// String returns the date with format DD/MM/AAAA
func (d Date) String() string {
	da := d.Time()
	if da.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d", da.Day(), da.Month(), da.Year())
}

// avantNoYear renvoie true si le jour et le mois de 'd1' sont avantNoYear (au sens large) le jour et le mois de 'd2',
// même si les années sont différentes
func avantNoYear(d1, d2 time.Time) bool {
	if d1.Month() < d2.Month() {
		return true
	}
	if d1.Month() > d2.Month() {
		return false
	}
	return d1.Day() <= d2.Day()
}

// before returns 'true' if d1 <= d2,
// only comparing Year, Month and Day
func before(d1, d2 time.Time) bool {
	y1, y2 := d1.Year(), d2.Year()
	if y1 < y2 {
		return true
	} else if y1 > y2 {
		return false
	}
	// years are the same
	return d1.YearDay() <= d2.YearDay()
}

// CalculeAge renvoie l'âge qu'aura une personne au jour `now`.
// Si la date de naissance est invalide, l'âge renvoyé est 0.
func (d Date) Age(now time.Time) int {
	dt := d.Time()
	if dt.Year() < 1000 { // vrai en particulier si la date est vide
		return 0
	}

	years := now.Year() - dt.Year()
	// check if the birth is before the present
	if avantNoYear(dt, now) {
		return years
	}
	return years - 1
}

// HasBirthday returns 'true' if someone born at [d]
// has his birthday during [pl]
func (pl Plage) HasBirthday(d Date) bool {
	ti := d.Time()
	if ti.IsZero() {
		return false
	}
	from, to := pl.From.Time(), pl.toT()
	if to.Year() > from.Year() { // 2000 -> 2001 : adjust the comparison
		return avantNoYear(from, ti) || avantNoYear(ti, to)
	}
	return avantNoYear(from, ti) && avantNoYear(ti, to)
}

var weekDays = [...]string{
	"Dim",
	"Lun",
	"Mar",
	"Mer",
	"Jeu",
	"Ven",
	"Sam",
}

// ShortString renvoie la date au format Mer 2
func (d Date) ShortString() string {
	t := d.Time()
	return fmt.Sprintf("%s %d", weekDays[t.Weekday()], t.Day())
}

func (d Date) AddDays(jours int) Date {
	const day = 24 * time.Hour
	return NewDateFrom(d.Time().Add(day * time.Duration(jours)))
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

func (pl Plage) toT() time.Time {
	from := pl.From.Time()
	return from.Add(time.Hour * 24 * time.Duration(pl.Duree-1))
}

func (pl Plage) To() Date { return NewDateFrom(pl.toT()) }

func (pl Plage) Contains(d Date) bool {
	dt := d.Time()
	return before(pl.From.Time(), dt) && before(dt, pl.toT())
}

func (s *Plage) Scan(src interface{}) error {
	if src == nil {
		return nil // zero value out
	}
	bs, ok := src.([]byte)
	if !ok {
		return errors.New("not a []byte")
	}
	return json.Unmarshal(bs, s)
}

func (s Plage) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return driver.Value(string(b)), nil
}

type OptID[T ~int64] struct {
	Id    T
	Valid bool
}

// Is returns [true] if the ID is valid and equal to v.
func (s OptID[T]) Is(v T) bool { return s.Valid && s.Id == v }

func (s *OptID[T]) Scan(src any) error {
	var tmp sql.NullInt64
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = OptID[T]{
		Valid: tmp.Valid,
		Id:    T(tmp.Int64),
	}
	return nil
}

func (s OptID[T]) Value() (driver.Value, error) {
	return sql.NullInt64{Int64: int64(s.Id), Valid: s.Valid}.Value()
}

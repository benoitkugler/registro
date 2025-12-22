package rawdata

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // SQL driver registration
)

// Handling of NULL values
func (s *Bool) Scan(src interface{}) error {
	var tmp sql.NullBool
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = Bool(tmp.Bool)
	return nil
}

func (s *Int) Scan(src interface{}) error {
	var tmp sql.NullInt64
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = Int(tmp.Int64)
	return nil
}

func (s *Float) Scan(src interface{}) error {
	var tmp sql.NullFloat64
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = Float(tmp.Float64)
	return nil
}

func (s *String) Scan(src interface{}) error {
	var tmp sql.NullString
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = String(tmp.String)
	return nil
}

func (s *Time) Scan(src interface{}) error {
	var tmp pq.NullTime
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = Time(tmp.Time)
	return nil
}

func (s Time) Value() (driver.Value, error) {
	pqTime := pq.NullTime{Time: time.Time(s), Valid: true}
	if s.Time().IsZero() {
		pqTime = pq.NullTime{}
	}
	return pqTime.Value()
}

func (s *Date) Scan(src interface{}) error {
	var tmp pq.NullTime
	err := tmp.Scan(src)
	if err != nil {
		return err
	}
	*s = Date(tmp.Time)
	return nil
}

func (s Date) Value() (driver.Value, error) {
	return pq.NullTime{Time: time.Time(s), Valid: true}.Value()
}

// custom types
func (s *OptionnalId) Scan(src interface{}) error {
	return (*sql.NullInt64)(s).Scan(src)
}

func (s OptionnalId) Value() (driver.Value, error) {
	return (sql.NullInt64)(s).Value()
}

func (s *Tels) Scan(src interface{}) error {
	return (*pq.StringArray)(s).Scan(src)
}

func (s Tels) Value() (driver.Value, error) {
	return (pq.StringArray)(s).Value()
}

func (s *Cotisation) Scan(src interface{}) error {
	return (*pq.Int64Array)(s).Scan(src)
}

func (s Cotisation) Value() (driver.Value, error) {
	return (pq.Int64Array)(s).Value()
}

func (s *OptionQuotientFamilial) Scan(src interface{}) error {
	var tmp pq.Float64Array
	if err := tmp.Scan(src); err != nil {
		return err
	}
	if len(tmp) != 4 {
		return fmt.Errorf("invalid length for type OptionQuotientFamilial: %d", len(tmp))
	}
	copy(s[:], tmp)
	return nil
}

func (s OptionQuotientFamilial) Value() (driver.Value, error) {
	return (pq.Float64Array)(s[:]).Value()
}

func (rs *Roles) Scan(src interface{}) error {
	var tmp pq.StringArray
	err := tmp.Scan(src)
	// on convertit les strings en Role
	b := make(Roles, len(tmp))
	for i, v := range tmp {
		b[i] = Role(v)
	}
	// on met à jour la cible du  pointeur
	*rs = b
	return err
}

func (rs Roles) Value() (driver.Value, error) {
	tmp := make(pq.StringArray, len(rs))
	for i, v := range rs {
		tmp[i] = string(v)
	}
	return tmp.Value()
}

func ScanIds(rs *sql.Rows) (Ids, error) {
	defer rs.Close()
	ints := make(Ids, 0, 16)
	var err error
	for rs.Next() {
		var s int64
		if err = rs.Scan(&s); err != nil {
			return nil, err
		}
		ints = append(ints, s)
	}
	if err = rs.Err(); err != nil {
		return nil, err
	}
	return ints, nil
}

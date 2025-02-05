package shared

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tu "registro/utils/testutils"

	"github.com/lib/pq"
)

func TestDateAndPlage(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TABLE t1 (id serial, date date, plage jsonb);
	`)
	tu.AssertNoErr(t, err)

	ti, pi := NewDate(2025, 5, 23), Plage{NewDateFrom(time.Now()), 5}
	_, err = db.Exec("INSERT INTO t1 (date, plage) VALUES ($1, $2)", ti, pi)
	tu.AssertNoErr(t, err)

	var (
		v Date
		p Plage
	)
	err = db.QueryRow("SELECT date, plage FROM t1;").Scan(&v, &p)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == ti)
	tu.Assert(t, p == pi)

	// JSON
	b, err := json.Marshal(v)
	tu.AssertNoErr(t, err)
	tu.Assert(t, string(b) == `"2025-05-23"`)
	var v2 Date
	err = json.Unmarshal(b, &v2)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == v2)
}

func TestPlage_To(t *testing.T) {
	for _, test := range []struct {
		debut    Date
		duree    int
		expected Date
	}{
		{NewDate(2000, time.January, 15), 1, NewDate(2000, time.January, 15)},
		{NewDate(2000, time.January, 15), 10, NewDate(2000, time.January, 24)},
		{NewDate(2000, time.January, 30), 3, NewDate(2000, time.February, 1)},
		{NewDate(2000, time.December, 30), 3, NewDate(2001, time.January, 1)},
	} {
		got := Plage{From: test.debut, Duree: test.duree}.To()
		tu.Assert(t, got == test.expected)
	}
}

func TestUint8Array(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TABLE t1 (id serial, roles smallint[]);
	`)
	tu.AssertNoErr(t, err)

	ti := pq.Int32Array{-1, 2, 8, 1000}
	_, err = db.Exec("INSERT INTO t1 (roles) VALUES ($1)", ti)
	tu.AssertNoErr(t, err)

	var v pq.Int32Array
	err = db.QueryRow("SELECT roles FROM t1;").Scan(&v)
	tu.AssertNoErr(t, err)
	fmt.Println(v)
}

func TestDate_Age(t *testing.T) {
	tests := []struct {
		d    Date
		now  Date
		want int
	}{
		{Date{}, Date{}, 0},
		{NewDate(2000, time.February, 5), NewDate(2000, time.February, 8), 0},
		{NewDate(2000, time.February, 5), NewDate(2001, time.February, 4), 0},
		{NewDate(2000, time.February, 5), NewDate(2001, time.February, 5), 1},
		{NewDate(2000, time.February, 5), NewDate(2001, time.September, 5), 1},
		{NewDate(2000, time.February, 5), NewDate(2001, time.January, 5), 0},
	}
	for _, tt := range tests {
		tu.Assert(t, tt.d.Age(tt.now) == tt.want)
	}
}

func TestDate_ShortString(t *testing.T) {
	tests := []struct {
		d    Date
		want string
	}{
		{NewDate(2025, time.February, 2), "Dim 2"},
		{NewDate(2025, time.February, 5), "Mer 5"},
	}
	for _, tt := range tests {
		tu.Assert(t, tt.d.ShortString() == tt.want)
	}
}

func TestDate_AddDays(t *testing.T) {
	tests := []struct {
		d     Date
		jours int
		want  Date
	}{
		{NewDate(2000, time.February, 2), 10, NewDate(2000, time.February, 12)},
		{NewDate(2000, time.February, 20), 10, NewDate(2000, time.March, 1)},
		{NewDate(2001, time.February, 20), 10, NewDate(2001, time.March, 2)},
	}
	for _, tt := range tests {
		tu.Assert(t, tt.d.AddDays(tt.jours) == tt.want)
	}
}

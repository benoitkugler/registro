package shared

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tu "registro/utils/testutils"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func TestDate(t *testing.T) {
	db := tu.NewTestDB(t)
	defer db.Remove()

	_, err := db.Exec(`
	CREATE TABLE t1 (id serial, date date);
	`)
	tu.AssertNoErr(t, err)

	ti := NewDate(2025, 5, 23)
	_, err = db.Exec("INSERT INTO t1 (date) VALUES ($1)", ti)
	tu.AssertNoErr(t, err)

	var v Date
	err = db.QueryRow("SELECT date FROM t1;").Scan(&v)
	tu.AssertNoErr(t, err)
	tu.Assert(t, v == ti)

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

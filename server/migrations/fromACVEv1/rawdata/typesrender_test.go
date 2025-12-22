package rawdata

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTel(t *testing.T) {
	tel := "05 48-5 956 "
	if f := FormatTel(tel); f != " 05 48 59 56" {
		t.Errorf("expected  05 48 59 56, got %s", f)
	}

	if f := (Tels{tel}); f.String() != "-05-48-59-56" {
		t.Errorf("expected -05-48-59-56, got %s", f)
	}
}

func TestDateSort(t *testing.T) {
	d1 := time.Date(2000, 10, 2, 12, 5, 4, 0, time.UTC)
	d2 := time.Date(2001, 10, 2, 12, 5, 4, 0, time.UTC)
	d3 := time.Date(2001, 9, 5, 12, 5, 4, 0, time.UTC)
	s1 := Date(d1)
	s2 := Date(d2)
	s3 := Date(d3)

	if s1.Sortable() < s2.Sortable() != d1.Before(d2) {
		t.Fail()
	}
	if s2.Sortable() < s3.Sortable() != d2.Before(d3) {
		t.Fail()
	}
	if s1.Sortable() < s3.Sortable() != d1.Before(d3) {
		t.Fail()
	}
}

func TestPlage(t *testing.T) {
	p := Plage{From: Date(time.Now().AddDate(-8, 0, 0)), To: Date(time.Now())}
	fmt.Println(p.ExpandMonths())
}

func TestSS(t *testing.T) {
	fmt.Println(FormatSecuriteSocial("19404  7855126776"))
	fmt.Println(FormatSecuriteSocial("  194  5516776"))
}

func TestPrenom(t *testing.T) {
	s := String("éean-pieàérre")
	fmt.Println(FormatPrenom(s))
}

func TestFloat(t *testing.T) {
	if s := Euros(1.325).String(); s != "1,33 €" {
		t.Error(s)
	}
	if s := Euros(1.324).String(); s != "1,32 €" {
		t.Error(s)
	}
}

func TestFormat(t *testing.T) {
	fmt.Println(strings.Replace(Euros(5.56).String(), ".", ",", -1))
}

func TestTimeZone(t *testing.T) {
	ts := "2020-09-01T08:43:58Z"
	ti, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(Time(ti))
}

func TestCotisation(t *testing.T) {
	c := Cotisation{2000, 2002, 2015, 2021, 2018}
	fmt.Println(c.Sortable())
	fmt.Println(c)
}

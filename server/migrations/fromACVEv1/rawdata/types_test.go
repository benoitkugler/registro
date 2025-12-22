package rawdata

import (
	"fmt"
	"testing"
	"time"
)

func TestPLage(t *testing.T) {
	p := Plage{From: Date(time.Now()), To: Date(time.Now().Add(48 * time.Hour))}
	if d := p.NbJours(); d != 3 {
		t.Errorf("expected 2 days, got %d", d)
	}

	p = Plage{From: Date(time.Date(2002, 1, 1, 12, 0, 0, 0, time.UTC)), To: Date(time.Date(2002, 1, 12, 0, 0, 0, 0, time.UTC))}
	if d := p.NbJours(); d != 12 {
		t.Errorf("expected 2 days, got %d", d)
	}
}

func TestContains(t *testing.T) {
	n := time.Now()
	p := Plage{From: Date(n), To: Date(n)}
	if !p.Contains(n.Add(2 * time.Second)) {
		t.Error("should contain itself")
	}
	p = Plage{}
	if !p.Contains(time.Now()) {
		t.Error("empty plage should contain everything")
	}
	p = Plage{From: Date(n), To: Date(n.Add(10 * 24 * time.Hour))}
	if !p.Contains(n.Add(5 * 24 * time.Hour)) {
		t.Error("should contain")
	}
}

func TestRound(t *testing.T) {
	if Euros(1.31349413).Round() != 1.31 {
		t.Fail()
	}
	if Euros(1.355).Round() != 1.36 {
		t.Fail()
	}
	// fmt.Println(math.Round(544.425*100) / 100)
	// fmt.Println(Euros(544.425).Round())
	// fmt.Println(float64(Euros(544.425)) * 100)
	// fmt.Println(float64(544.425) * 100)
	// fmt.Println(float64(Euros(544.425) - Euros(544.42)))

	e := Euros(0.3999999999999)

	if v := e.Centimes(); v != 40 {
		t.Fatalf("expected 40, got %d", v)
	}

	e = Euros(0.07999999999992724)
	if v := e.Centimes(); v != 8 {
		t.Fatalf("expected 40, got %d", v)
	}
}

func TestCompare(t *testing.T) {
	if !Euros(1.321).IsEqual(Euros(1.324)) {
		t.Fail()
	}
	if !Euros(1.321).IsLess(Euros(1.322)) {
		t.Fail()
	}
}

func TestHintsAttente(t *testing.T) {
	type testHA struct {
		ha       HintsAttente
		expected StatutAttente
	}
	values := []testHA{
		{ha: HintsAttente{AgeMax: Attente}, expected: Attente},
		{ha: HintsAttente{AgeMax: Refuse, AgeMin: Attente}, expected: Refuse},
		{ha: HintsAttente{}, expected: Inscrit},
		{ha: HintsAttente{EquilibreGF: Refuse}, expected: Refuse},
		{ha: HintsAttente{AgeMin: Attente, Place: Attente}, expected: Attente},
	}
	for _, v := range values {
		got := v.ha.Hint()
		if got != v.expected {
			t.Errorf("for hints %v, expected %s got %s", v.ha, v.expected, got)
		}
	}
}

func TestDescriptionJours(t *testing.T) {
	js := Jours{0, 1, 1, 4, 8}
	s := js.Description(Camp{DateDebut: Date(time.Now()), DateFin: Date(time.Now().Add(456 * time.Hour))})
	fmt.Println(s)
}

package camps

import (
	"reflect"
	"testing"
)

func TestPrixQuotientFamilial_Percentage(t *testing.T) {
	grid := DefaultPrixQuotientFamilial()
	tests := []struct {
		qf   int
		want int
	}{
		{0, 100},
		{1, 20},
		{359, 20},
		{360, 40},
		{564, 40},
		{565, 60},
		{714, 60},
		{715, 100},
		{2000, 100},
	}
	for _, tt := range tests {
		if got := grid.Percentage(tt.qf); got != tt.want {
			t.Errorf("Percentage(%d) = %d, want %d", tt.qf, got, tt.want)
		}
	}

	custom := PrixQuotientFamilial{
		{QuotientMin: 100, Pourcentage: 10},
		{QuotientMin: 900, Pourcentage: 100},
	}
	if got := custom.Percentage(50); got != 10 {
		t.Errorf("Percentage below first tranche = %d, want 10", got)
	}
	if got := custom.Percentage(900); got != 100 {
		t.Errorf("Percentage above last tranche = %d, want 100", got)
	}
	if got := (PrixQuotientFamilial{}).Percentage(100); got != 100 {
		t.Errorf("Percentage with empty grid = %d, want 100", got)
	}
}

func TestPrixQuotientFamilial_IsActive(t *testing.T) {
	if (PrixQuotientFamilial{}).IsActive() {
		t.Error("an empty grid should be inactive")
	}
	if !DefaultPrixQuotientFamilial().IsActive() {
		t.Error("a non-empty grid should be active")
	}
}

func TestPrixQuotientFamilial_Check(t *testing.T) {
	tests := []struct {
		name    string
		grid    PrixQuotientFamilial
		wantErr bool
	}{
		{"empty", PrixQuotientFamilial{}, false},
		{"default", DefaultPrixQuotientFamilial(), false},
		{"single", PrixQuotientFamilial{{QuotientMin: 0, Pourcentage: 100}}, false},
		{"unsorted", PrixQuotientFamilial{{QuotientMin: 10, Pourcentage: 50}, {QuotientMin: 5, Pourcentage: 100}}, true},
		{"bad percentage", PrixQuotientFamilial{{QuotientMin: 0, Pourcentage: 101}, {QuotientMin: 10, Pourcentage: 100}}, true},
		{"last not 100", PrixQuotientFamilial{{QuotientMin: 0, Pourcentage: 20}, {QuotientMin: 10, Pourcentage: 90}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.grid.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrixQuotientFamilial_SQLValue(t *testing.T) {
	grid := DefaultPrixQuotientFamilial()
	value, err := grid.Value()
	if err != nil {
		t.Fatal(err)
	}
	var got PrixQuotientFamilial
	if err := got.Scan([]byte(value.(string))); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, grid) {
		t.Errorf("round trip = %v, want %v", got, grid)
	}

	// La valeur zéro est stockée comme un tableau vide, jamais comme NULL.
	value, err = (PrixQuotientFamilial)(nil).Value()
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "[]" {
		t.Errorf("zero value = %q, want []", value)
	}
}

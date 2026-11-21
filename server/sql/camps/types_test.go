package camps

import (
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

func TestPrixParAge_Check(t *testing.T) {
	tests := []struct {
		name    string
		grid    PrixParAge
		wantErr bool
	}{
		{"empty", PrixParAge{}, true},
		{"single", PrixParAge{{AgeMax: 18, Prix: 5000}}, false},
		{"sorted", PrixParAge{{AgeMax: 10, Prix: 5000}, {AgeMax: 18, Prix: 8000}}, false},
		{"unsorted", PrixParAge{{AgeMax: 18, Prix: 5000}, {AgeMax: 10, Prix: 8000}}, true},
		{"zero AgeMax", PrixParAge{{AgeMax: 0, Prix: 5000}}, true},
		{"negative price", PrixParAge{{AgeMax: 10, Prix: -1}}, true},
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

func TestPrixParAge_Prix(t *testing.T) {
	grid := PrixParAge{{AgeMax: 10, Prix: 5000}, {AgeMax: 18, Prix: 8000}}
	tests := []struct {
		age      int
		want     int
		wantFind bool
	}{
		{6, 5000, true},
		{9, 5000, true},
		{10, 8000, true},
		{17, 8000, true},
		{18, 0, false},
		{30, 0, false},
	}
	for _, tt := range tests {
		got, ok := grid.Prix(tt.age)
		if got != tt.want || ok != tt.wantFind {
			t.Errorf("Prix(%d) = (%d, %v), want (%d, %v)", tt.age, got, ok, tt.want, tt.wantFind)
		}
	}
	if (PrixParAge{}).IsActive() {
		t.Error("an empty grid should be inactive")
	}
	if !grid.IsActive() {
		t.Error("a non-empty grid should be active")
	}
}

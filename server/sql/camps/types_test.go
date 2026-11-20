package camps

import "testing"

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

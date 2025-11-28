package communes_test

import (
	"registro/controllers/equipier/communes"
	"testing"
)

func TestCommuneByCode(t *testing.T) {
	tests := []struct {
		insee   string
		want    string
		want2   string
		wantErr bool
	}{
		{"28386", "Eure-et-Loir", "Thimert-Gâtelles", false},
		{"2B355", "Haute-Corse", "Volpajola", false},
		{"2A004", "Corse-du-Sud", "Ajaccio", false},
		{"200001", "", "", true},
		{"78000", "", "", true},
		{"780R0", "", "", true},
		{"99001", "Hors France", "", false},
		{"97112", "Guadeloupe", "Grand-Bourg", false},
		{"97901", "", "", false},
	}
	for _, tt := range tests {
		got, got2, gotErr := communes.CommuneByCode(tt.insee)
		if gotErr != nil && !tt.wantErr {
			t.Fatalf("CommuneByCode() failed: %v", gotErr)
		}
		if tt.wantErr && gotErr == nil {
			t.Fatal("CommuneByCode() succeeded unexpectedly", tt.insee)
		}
		if got != tt.want {
			t.Fatalf("CommuneByCode() = %v, want %v", got, tt.want)
		}
		if got2 != tt.want2 {
			t.Fatalf("CommuneByCode() = %v, want %v", got2, tt.want2)
		}
	}
}

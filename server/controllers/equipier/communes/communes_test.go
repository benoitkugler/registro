package communes_test

import (
	"testing"

	"registro/controllers/equipier/communes"
)

func TestCommuneByCode(t *testing.T) {
	tests := []struct {
		insee   string
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{"28386", "Eure-et-Loir", "28", "Thimert-Gâtelles", false},
		{"2B355", "Haute-Corse", "2B", "Volpajola", false},
		{"2A004", "Corse-du-Sud", "2A", "Ajaccio", false},
		{"200001", "", "", "", true},
		{"78000", "", "", "", true},
		{"780R0", "", "", "", true},
		{"99001", "Hors France", "99", "", false},
		{"97112", "Guadeloupe", "971", "Grand-Bourg", false},
		{"97901", "", "979", "", false},
	}
	for _, tt := range tests {
		got, got1, got2, gotErr := communes.CommuneByCode(tt.insee)
		if gotErr != nil && !tt.wantErr {
			t.Fatalf("CommuneByCode() failed: %v", gotErr)
		}
		if tt.wantErr && gotErr == nil {
			t.Fatal("CommuneByCode() succeeded unexpectedly", tt.insee)
		}
		if got != tt.want {
			t.Fatalf("CommuneByCode() = %v, want %v", got, tt.want)
		}
		if got1 != tt.want1 {
			t.Fatalf("CommuneByCode() = %v, want %v", got1, tt.want1)
		}
		if got2 != tt.want2 {
			t.Fatalf("CommuneByCode() = %v, want %v", got2, tt.want2)
		}
	}
}

package equipier

import (
	"testing"
	"time"

	pr "registro/sql/personnes"
)

func d(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}

func Test_checkSecuriteSociale(t *testing.T) {
	tests := []struct {
		sexe            pr.Sexe
		dateNaissance   time.Time
		securiteSociale string
		wantErr         bool
	}{
		{pr.Man, d(1994, 5), "1 94 05 78 551 268 91", false},
		{pr.Man, d(1994, 5), "194057855126891", false},
		{pr.Man, d(1994, 5), "1 94 05 78 551 AER 91", true},
		{pr.Man, d(1994, 5), "1 94 05 78 551 268 AF", true},
		{pr.Man, d(1994, 5), "1 94 05 AF 000 268 91", true},
		{pr.Man, d(1994, 5), "1 94 05 78", true},
		{pr.Man, d(1994, 5), "1 ER 05 78 551 268 91", true},
		{pr.Man, d(1994, 5), "1 94 AB 78 551 268 91", true},
		{pr.Man, d(1994, 5), "1 94 12 78 551 268 91", true},
		{pr.Man, d(1994, 5), "1 94 05 78 551 290 91", true},
		{pr.Man, d(1995, 5), "1 94 05 78 551 268 91", true},
		{pr.Man, d(1994, 5), "2 94 05 78 551 268 91", true},
	}
	for _, tt := range tests {
		got := checkSecuriteSociale(tt.sexe, tt.dateNaissance, tt.securiteSociale)
		if (got.Err != "") != tt.wantErr {
			t.Errorf("checkSecuriteSociale(%s) = %v", tt.securiteSociale, got)
		}
	}
}

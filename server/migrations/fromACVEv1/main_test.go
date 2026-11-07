package main

import (
	"testing"
	"time"

	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/shared"
)

func TestPresenceOffsets(t *testing.T) {
	camp := shared.Plage{From: shared.NewDate(2025, time.July, 10), Duree: 5}

	t.Run("inactive presence", func(t *testing.T) {
		got := presenceOffsets(rd.OptionnalPlage{
			Plage:  rd.Plage{From: rd.Date(time.Date(2025, time.July, 11, 0, 0, 0, 0, time.UTC))},
			Active: false,
		}, camp)
		if got != (cps.PresenceOffsets{}) {
			t.Errorf("presenceOffsets(inactive) = %+v, want zero offsets", got)
		}
	})

	t.Run("active presence dates are relative to camp", func(t *testing.T) {
		got := presenceOffsets(rd.OptionnalPlage{
			Plage: rd.Plage{
				From: rd.Date(time.Date(2025, time.July, 12, 0, 0, 0, 0, time.UTC)),
				To:   rd.Date(time.Date(2025, time.July, 14, 0, 0, 0, 0, time.UTC)),
			},
			Active: true,
		}, camp)
		want := cps.PresenceOffsets{Debut: 2, Fin: 0}
		if got != want {
			t.Errorf("presenceOffsets(active) = %+v, want %+v", got, want)
		}

		got = presenceOffsets(rd.OptionnalPlage{
			Plage: rd.Plage{
				From: rd.Date(time.Date(2025, time.July, 10, 0, 0, 0, 0, time.UTC)),
				To:   rd.Date(time.Date(2025, time.July, 13, 0, 0, 0, 0, time.UTC)),
			},
			Active: true,
		}, camp)
		want = cps.PresenceOffsets{Debut: 0, Fin: -1}
		if got != want {
			t.Errorf("presenceOffsets(active) = %+v, want %+v", got, want)
		}
	})
}

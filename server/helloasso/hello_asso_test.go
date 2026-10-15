package helloasso

import (
	"testing"
	"time"

	"registro/config"
	tu "registro/utils/testutils"
)

func devCreds(t *testing.T) config.Helloasso {
	tu.LoadEnv(t, "../env.sh")
	creds, err := config.NewHelloasso()
	tu.AssertNoErr(t, err)
	return creds
}

func TestPing(t *testing.T) {
	_, err := NewApi(devCreds(t))
	tu.AssertNoErr(t, err)
}

func TestDate(t *testing.T) {
	var d time.Time
	err := d.UnmarshalJSON([]byte(`"2025-12-23T13:25:40.276Z"`))
	tu.AssertNoErr(t, err)
	tu.Assert(t, d.Year() == 2025 && d.Month() == time.December && d.Day() == 23)
}

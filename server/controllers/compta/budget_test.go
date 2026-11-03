package compta

import (
	"testing"

	ca "registro/sql/compta"
	tu "registro/utils/testutils"
)

func TestSelectConfig(t *testing.T) {
	for _, test := range []struct {
		configYears []int
		wantYear    int
		wantError   bool
	}{
		{[]int{2000, 2001, 2023, 2024}, 2023, false},
		{[]int{2000, 2001, 2024}, 2001, false},
		{[]int{2024}, 0, true},
		{[]int{}, 0, true},
		{[]int{2017}, 2017, false},
	} {
		configs := ca.ComptaConfigs{}
		for i, y := range test.configYears {
			configs[ca.IdComptaConfig(i)] = ca.ComptaConfig{Year: y}
		}
		got, err := selectBudgetConfig(2023, configs)
		tu.Assert(t, err != nil == test.wantError)
		tu.Assert(t, got.Year == test.wantYear)
	}
}

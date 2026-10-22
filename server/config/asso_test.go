package config

import "testing"

func TestAsso(t *testing.T) {
	if repere.ID != "repere" || acve.ID != "acve" {
		t.Fatal()
	}
}

package config

import "testing"

func TestAsso(t *testing.T) {
	if repere.ID != "repere" || acve.ID != "acve" {
		t.Fatal()
	}
	if !repere.AskCarteIdInscrits || acve.AskCarteIdInscrits {
		t.Fatal("CarteId request setting should be enabled for Repère and disabled for ACVE")
	}
}

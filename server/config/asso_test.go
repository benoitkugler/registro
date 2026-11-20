package config

import "testing"

func TestAsso(t *testing.T) {
	if repere.ID != "repere" || acve.ID != "acve" {
		t.Fatal()
	}
	if !repere.AskedDocs.CarteId || acve.AskedDocs.CarteId {
		t.Fatal("CarteId request setting should be enabled for Repère and disabled for ACVE")
	}
	if acve.ComptaContactLabel != "le centre" || repere.ComptaContactLabel != "le comptable" {
		t.Fatal("unexpected ComptaContactLabel")
	}
}

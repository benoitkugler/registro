package rawdata

import (
	"testing"
)

func TestUrlEspacePerso(t *testing.T) {
	f := Facture{Key: "lmdsmkdlmd65687d32sd"}
	if s := f.UrlEspacePerso("https://acve.fr/espace_perso/"); s != "https:/acve.fr/espace_perso/lmdsmkdlmd65687d32sd" {
		t.Errorf("got %s", s)
	}
}

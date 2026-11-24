package caf

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	tu "registro/utils/testutils"

	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/fonts/builder"
	"github.com/benoitkugler/pdf/fonts/standardfonts"
	"github.com/benoitkugler/pdf/model"
)

func TestParseAttestationText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    Attestation
		wantErr error
	}{
		{
			name: "native CAF wording - quotient",
			text: "Le directeur de la Caf de Vaucluse certifie que votre quotient familial (QF) s’élève pour le mois de octobre 2026 à : 1594 €.",
			want: Attestation{QuotientFamilial: 1594, Mois: time.October, Annee: 2026},
		},
		{
			name: "CAF currency sign",
			text: "Votre quotient familial (QF) s’élève pour le mois de octobre 2025 à : 1654 ¤.",
			want: Attestation{QuotientFamilial: 1654, Mois: time.October, Annee: 2025},
		},
		{
			name: "glued words from native CAF PDF",
			text: "Attestationdequotientfamilial Votrequotientfamilial(QF)s’élèvepourlemoisdeoctobre2025à: 1654¤.",
			want: Attestation{QuotientFamilial: 1654, Mois: time.October, Annee: 2025},
		},
		{
			name: "payment attestation with prestations",
			text: "Attestation de paiement. Prestations pour le mois d'octobre 2025 : Allocation de logement 429,00 ¤ Soit au total 1615,97 ¤ QUOTIENT FAMILIAL octobre 2025 : 645 ¤.",
			want: Attestation{QuotientFamilial: 645, Mois: time.October, Annee: 2025},
		},
		{
			name: "glued payment attestation",
			text: "Attestationdepaiement. Prestationspourlemoisd'octobre2025: Allocationdelogement429,00¤ Soitautotal1615,97¤ QUOTIENTFAMILIAL octobre2025:689¤.",
			want: Attestation{QuotientFamilial: 689, Mois: time.October, Annee: 2025},
		},
		{
			name: "native CAF wording",
			text: "Caf de Seine-et-Marne. Votre quotient familial est de 1 234 € pour le mois de janvier 2026.",
			want: Attestation{QuotientFamilial: 1234, Mois: time.January, Annee: 2026},
		},
		{
			name: "narrow spaces and accent",
			text: "Votre quotient familial s’élève à 2\u00a0345\u202f€ pour le mois d’août 2025.",
			want: Attestation{QuotientFamilial: 2345, Mois: time.August, Annee: 2025},
		},
		{
			name: "uppercase and euros word",
			text: "QUOTIENT FAMILIAL : 3 000 EUROS POUR LE MOIS DE FÉVRIER 2024",
			want: Attestation{QuotientFamilial: 3000, Mois: time.February, Annee: 2024},
		},
		{
			name: "QF abbreviation",
			text: "QF : 987 € - mois de mars 2026",
			want: Attestation{QuotientFamilial: 987, Mois: time.March, Annee: 2026},
		},
		{
			name: "amount before anchor",
			text: "1 456 € au titre du quotient familial pour le mois de mai 2026",
			want: Attestation{QuotientFamilial: 1456, Mois: time.May, Annee: 2026},
		},
		{
			name: "zero cents are accepted",
			text: "quotient familial 1234,00 € pour le mois de juin 2026",
			want: Attestation{QuotientFamilial: 1234, Mois: time.June, Annee: 2026},
		},
		{
			name:    "non-zero cents are rejected",
			text:    "quotient familial 1234,56 € pour le mois de janvier 2026",
			wantErr: ErrUnsupportedDocument,
		},
		{
			name:    "several amounts",
			text:    "quotient familial 500 € puis quotient familial 600 € pour le mois de janvier 2026",
			wantErr: ErrAmbiguous,
		},
		{
			name:    "MSA attestation",
			text:    "Attestation MSA. quotient familial 500 € pour le mois de janvier 2026",
			wantErr: ErrUnsupportedDocument,
		},
		{
			name:    "no amount",
			text:    "Attestation de présence pour le mois de janvier 2026",
			wantErr: ErrUnsupportedDocument,
		},
		{
			name:    "unknown month",
			text:    "quotient familial 500 € pour le mois de foobar 2026",
			wantErr: ErrUnsupportedDocument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAttestationText(tc.text)
			if tc.wantErr != nil {
				tu.Assert(t, errors.Is(err, tc.wantErr))
				return
			}
			tu.AssertNoErr(t, err)
			tu.Assert(t, got == tc.want)
		})
	}
}

func TestParseAttestationSyntheticPDF(t *testing.T) {
	pdf := makeTextPDF(t,
		"REPUBLIQUE FRANCAISE",
		"Caf de Seine-et-Marne",
		"Attestation de quotient familial",
		"Votre quotient familial est de 1 234 €",
		"pour le mois de janvier 2026.",
	)

	got, err := ParseAttestation(pdf)
	tu.AssertNoErr(t, err)
	want := Attestation{QuotientFamilial: 1234, Mois: time.January, Annee: 2026}
	tu.Assert(t, got == want)
}

func TestParseAttestationInvalidPDF(t *testing.T) {
	_, err := ParseAttestation([]byte("not a PDF"))
	tu.Assert(t, errors.Is(err, ErrInvalidPDF))
}

func TestParseAttestationWithoutTextLayer(t *testing.T) {
	var doc model.Document
	var page model.PageObject
	page.MediaBox = &model.Rectangle{Urx: 595.28, Ury: 841.89}
	doc.Catalog.Pages.Kids = []model.PageNode{&page}

	var buf bytes.Buffer
	err := doc.Write(&buf, nil)
	tu.AssertNoErr(t, err)

	_, err = ParseAttestation(buf.Bytes())
	tu.Assert(t, errors.Is(err, ErrNoExtractableText))
}

// TestParseAttestationRealFixture runs once the anonymized PDF and its
// expected values are added to the test/ directory.
func TestParseAttestationRealFixture(t *testing.T) {
	t.Run("attestation de quotient", func(t *testing.T) {
		pdfPath := "test/attestation_quotient.pdf"
		want := Attestation{1894, time.October, 2026}

		pdf, err := os.ReadFile(pdfPath)
		tu.AssertNoErr(t, err)

		got, err := ParseAttestation(pdf)
		tu.AssertNoErr(t, err)
		tu.Assert(t, got == want)
	})

	t.Run("attestation de paiement", func(t *testing.T) {
		pdfPath := "test/attestation_paiement.pdf"
		want := Attestation{657, time.October, 2026}

		pdf, err := os.ReadFile(pdfPath)
		tu.AssertNoErr(t, err)

		got, err := ParseAttestation(pdf)
		tu.AssertNoErr(t, err)
		tu.Assert(t, got == want)
	})
}

func makeTextPDF(t *testing.T, lines ...string) []byte {
	t.Helper()

	fontDict := model.FontDict{Subtype: standardfonts.Helvetica.WesternType1Font()}
	font, err := builder.BuildFont(&fontDict)
	tu.AssertNoErr(t, err)

	stream := contentstream.NewGraphicStream(model.Rectangle{Urx: 595.28, Ury: 841.89})
	stream.BeginText()
	stream.SetFontAndSize(font, 12)
	stream.SetLeading(16)
	stream.MoveText(48, 760)
	for i, line := range lines {
		if i == 0 {
			err = stream.ShowText(line)
		} else {
			err = stream.NewlineShowText(line)
		}
		tu.AssertNoErr(t, err)
	}
	stream.EndText()

	var page model.PageObject
	stream.ApplyToPageObject(&page, false)
	var doc model.Document
	doc.Catalog.Pages.Kids = []model.PageNode{&page}

	var buf bytes.Buffer
	err = doc.Write(&buf, nil)
	tu.AssertNoErr(t, err)
	return buf.Bytes()
}

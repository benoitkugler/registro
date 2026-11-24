// Package caf extracts the quotient familial and its period from a native
// attestation issued by a French CAF (CNAF) from its "Mon Compte" service.
//
// Only native PDFs with a text layer are supported. Scans, MSA attestations
// and Service-Public attestations are deliberately rejected.
package caf

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidPDF indicates that the input could not be read as a PDF.
	ErrInvalidPDF = errors.New("invalid PDF")
	// ErrNoExtractableText indicates a scanned or image-only PDF.
	ErrNoExtractableText = errors.New("no extractable text")
	// ErrUnsupportedDocument indicates a PDF that is not a native CAF attestation.
	ErrUnsupportedDocument = errors.New("unsupported document")
	// ErrAmbiguous indicates conflicting values in the document.
	ErrAmbiguous = errors.New("ambiguous attestation")
)

// Attestation holds the values extracted from the attestation.
type Attestation struct {
	// QuotientFamilial is expressed in euros. The CAF publishes it as an integer.
	QuotientFamilial int
	Mois             time.Month
	Annee            int
}

// ParseAttestation extracts the quotient familial and its month from pdf.
func ParseAttestation(pdf []byte) (Attestation, error) {
	text, err := extractText(pdf)
	if err != nil {
		return Attestation{}, err
	}
	return parseAttestationText(text)
}

const amountPattern = `\d+(?:[\s’']\d+)*(?:[.,]\d{1,2})?`

var (
	qfAnchorRE = regexp.MustCompile(`(?i)(?:quotient\s*familial|\bQF\b)`)
	amountRE   = regexp.MustCompile(`(?i)(` + amountPattern + `)\s*(?:€|¤|euros?\b)`)
	// datedQFRE matches the explicit "month year [à] : amount" wording used by
	// native CAF attestations, for example "octobre 2026 à : 1894 ¤" or
	// "octobre2026: 657¤". It identifies the quotient familial even when the
	// document also lists other euro amounts (prestations).
	datedQFRE = regexp.MustCompile(`(?i)20\d{2}\s*(?:à\s*)?:\s*(` + amountPattern + `)\s*(?:€|¤|euros?\b)`)
	monthRE   = regexp.MustCompile(`(?i)mois\s*(?:de\s*|d['’]\s*)([a-zà-ÿ]+)\s*(20\d{2})`)
	otherRE   = regexp.MustCompile(`(?i)\b(?:MSA|mutualité\s+sociale\s+agricole|service-public\.fr|non-allocataire)\b`)
)

const amountWindow = 200

type period struct {
	mois  string
	annee int
}

func parseAttestationText(text string) (Attestation, error) {
	text = normalizeText(text)
	if otherRE.MatchString(text) {
		return Attestation{}, fmt.Errorf("%w: MSA or Service-Public attestation", ErrUnsupportedDocument)
	}

	amounts := findAmounts(text)
	if len(amounts) == 0 {
		return Attestation{}, fmt.Errorf("%w: no quotient familial amount", ErrUnsupportedDocument)
	}
	if len(amounts) > 1 {
		return Attestation{}, fmt.Errorf("%w: several quotient familial amounts (%s)", ErrAmbiguous, strings.Join(amounts, ", "))
	}

	periods := findPeriods(text)
	if len(periods) == 0 {
		return Attestation{}, fmt.Errorf("%w: no month of validity", ErrUnsupportedDocument)
	}
	if len(periods) > 1 {
		return Attestation{}, fmt.Errorf("%w: several validity periods", ErrAmbiguous)
	}

	qf, err := parseMontant(amounts[0])
	if err != nil {
		return Attestation{}, err
	}
	p := periods[0]
	mois, ok := moisFR[p.mois]
	if !ok {
		return Attestation{}, fmt.Errorf("%w: unknown month %q", ErrUnsupportedDocument, p.mois)
	}
	if p.annee < 2000 || p.annee > 2100 {
		return Attestation{}, fmt.Errorf("%w: invalid year %d", ErrUnsupportedDocument, p.annee)
	}

	return Attestation{QuotientFamilial: qf, Mois: mois, Annee: p.annee}, nil
}

// findAmounts looks for a euro amount close to a "quotient familial" anchor,
// either before or after it. Duplicate amounts are collapsed.
func findAmounts(text string) []string {
	// Native CAF attestations state the quotient familial explicitly after a
	// month and year. Prefer it: payment attestations also list prestations
	// with euro amounts, which would otherwise make the result ambiguous.
	if qfAnchorRE.MatchString(text) {
		if matches := datedQFRE.FindAllStringSubmatch(text, -1); len(matches) == 1 {
			return []string{matches[0][1]}
		}
	}

	seen := make(map[string]bool)
	anchors := qfAnchorRE.FindAllStringIndex(text, -1)

	for i, loc := range anchors {
		start := loc[1]
		end := len(text)
		if i+1 < len(anchors) && anchors[i+1][0] > start {
			end = anchors[i+1][0]
		}
		if end-start > amountWindow {
			end = start + amountWindow
		}
		if m := amountRE.FindStringSubmatch(text[start:end]); m != nil {
			seen[m[1]] = true
		}

		before := loc[0] - amountWindow
		if before < 0 {
			before = 0
		}
		if m := amountRE.FindStringSubmatch(text[before:loc[0]]); m != nil {
			seen[m[1]] = true
		}
	}

	out := make([]string, 0, len(seen))
	for amount := range seen {
		out = append(out, amount)
	}
	sort.Strings(out)
	return out
}

func findPeriods(text string) []period {
	seen := make(map[period]bool)
	for _, match := range monthRE.FindAllStringSubmatch(text, -1) {
		annee, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		seen[period{mois: strings.ToLower(match[1]), annee: annee}] = true
	}

	out := make([]period, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].annee != out[j].annee {
			return out[i].annee < out[j].annee
		}
		return out[i].mois < out[j].mois
	})
	return out
}

var moisFR = map[string]time.Month{
	"janvier":   time.January,
	"février":   time.February,
	"mars":      time.March,
	"avril":     time.April,
	"mai":       time.May,
	"juin":      time.June,
	"juillet":   time.July,
	"août":      time.August,
	"septembre": time.September,
	"octobre":   time.October,
	"novembre":  time.November,
	"décembre":  time.December,
}

// parseMontant converts a french-formatted euro amount to an integer number
// of euros. Non-zero cents are rejected since the CAF never publishes them.
func parseMontant(raw string) (int, error) {
	s := unicodeSpaces.Replace(raw)
	s = strings.Map(func(r rune) rune {
		if r == '\'' || r == '\u2019' {
			return -1
		}
		return r
	}, s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")

	if before, after, ok := strings.Cut(s, "."); ok {
		if strings.Trim(after, "0") != "" {
			return 0, fmt.Errorf("%w: non-integer quotient familial %q", ErrUnsupportedDocument, raw)
		}
		s = before
	}

	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%w: invalid quotient familial %q", ErrUnsupportedDocument, raw)
	}
	return n, nil
}

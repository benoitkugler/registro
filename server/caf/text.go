package caf

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/fonts/builder"
	"github.com/benoitkugler/pdf/fonts/cmaps"
	"github.com/benoitkugler/pdf/model"
	"github.com/benoitkugler/pdf/reader"
	"github.com/benoitkugler/pdf/reader/parser"
	"github.com/benoitkugler/pdf/reader/streams"
)

// fontDecoder converts the char codes found in a content stream into
// unicode. It is built once per font dictionary and reused for each page.
type fontDecoder struct {
	simple    map[byte]rune
	toUnicode map[model.CID][]rune
	cmap      *cmaps.CMap
	identity  bool
}

func newFontDecoder(font *model.FontDict) fontDecoder {
	var out fontDecoder
	if font == nil {
		return out
	}

	if font.ToUnicode != nil {
		content, err := streams.DecodeStream(font.ToUnicode.Stream)
		if err == nil {
			cmap, err := cmaps.ParseUnicodeCMap(content)
			if err == nil {
				out.toUnicode = cmap.ProperLookupTable()
			}
		}
	}

	switch ft := font.Subtype.(type) {
	case model.FontType0:
		switch enc := ft.Encoding.(type) {
		case model.CMapEncodingEmbedded:
			content, err := streams.DecodeStream(enc.Stream)
			if err == nil {
				cmap, err := cmaps.ParseCIDCMap(content)
				if err == nil {
					out.cmap = &cmap
				}
			}
		case model.CMapEncodingPredefined:
			name := strings.ToLower(string(enc))
			out.identity = name == "identity-h" || name == "identity-v"
		}
	default:
		if simple, ok := font.Subtype.(model.FontSimple); ok {
			out.simple = builder.ResolveSimpleEncoding(simple).ByteToRune()
		}
	}

	return out
}

func (d fontDecoder) decode(codes []byte) string {
	if len(codes) == 0 {
		return ""
	}

	var out strings.Builder
	if d.identity || d.cmap != nil || len(d.toUnicode) != 0 {
		for _, cid := range d.charCodesToCIDs(codes) {
			if runes, ok := d.toUnicode[cid]; ok {
				out.WriteString(string(runes))
				continue
			}
			if cid <= 255 {
				if r, ok := d.simple[byte(cid)]; ok {
					out.WriteRune(r)
				}
			}
		}
		if out.Len() != 0 {
			return out.String()
		}
	}

	for _, code := range codes {
		if r, ok := d.simple[code]; ok {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func (d fontDecoder) charCodesToCIDs(codes []byte) []model.CID {
	if d.cmap != nil {
		charCodes, ok := d.cmap.BytesToCharcodes(codes)
		if !ok {
			return nil
		}
		toCID := d.cmap.CharCodeToCID()
		out := make([]model.CID, 0, len(charCodes))
		for _, code := range charCodes {
			cid, ok := toCID[code]
			if !ok {
				continue
			}
			out = append(out, cid)
		}
		return out
	}

	if d.identity {
		out := make([]model.CID, 0, len(codes)/2)
		for i := 0; i+1 < len(codes); i += 2 {
			out = append(out, model.CID(uint16(codes[i])<<8|uint16(codes[i+1])))
		}
		return out
	}

	// For simple fonts the char code is the CID itself.
	out := make([]model.CID, 0, len(codes))
	for _, code := range codes {
		out = append(out, model.CID(code))
	}
	return out
}

// extractText reads every page of pdf and returns a normalized, single-line
// representation of the text it contains.
func extractText(pdf []byte) (string, error) {
	doc, _, err := reader.ParsePDFReader(bytes.NewReader(pdf), reader.Options{})
	if err != nil {
		return "", fmt.Errorf("%w : le fichier n'est pas un PDF lisible", ErrInvalidPDF)
	}

	decoders := make(map[*model.FontDict]fontDecoder)
	var all strings.Builder
	for _, page := range doc.Catalog.Pages.Flatten() {
		content, err := streams.DecodePageContents(page)
		if err != nil {
			return "", fmt.Errorf("%w : impossible de décoder le contenu de la page", ErrInvalidPDF)
		}
		var colorSpaces model.ResourcesColorSpace
		if page.Resources != nil {
			colorSpaces = page.Resources.ColorSpace
		}
		ops, err := parser.ParseContent(content, colorSpaces)
		if err != nil {
			return "", fmt.Errorf("%w : impossible d'analyser le contenu de la page", ErrInvalidPDF)
		}

		pageText, err := extractPageText(ops, page, decoders)
		if err != nil {
			return "", err
		}
		if pageText != "" {
			all.WriteString(pageText)
			all.WriteByte('\n')
		}
	}

	text := normalizeText(all.String())
	if text == "" {
		return "", fmt.Errorf("%w : le PDF ne contient pas de texte (document scanné ?)", ErrNoExtractableText)
	}
	return text, nil
}

func extractPageText(ops []contentstream.Operation, page *model.PageObject, decoders map[*model.FontDict]fontDecoder) (string, error) {
	var (
		line     strings.Builder
		pageText strings.Builder
		decoder  fontDecoder
		lastX    model.Fl
		lastY    model.Fl
		fontSize model.Fl
		havePos  bool
	)

	flushLine := func() {
		if line.Len() != 0 {
			pageText.WriteString(line.String())
			pageText.WriteByte('\n')
			line.Reset()
		}
	}

	// setPosition records a new text position. Text sharing a Y coordinate but
	// placed far apart in X (a right-aligned footer, a table header cell, ...)
	// belongs to different visual blocks: start a new line so normalizeText
	// separates them with a space instead of gluing the words together.
	setPosition := func(x, y model.Fl) {
		gap := fontSize
		if gap < 1 {
			gap = 1
		}
		if havePos && (math.Abs(float64(y-lastY)) > 0.5 || math.Abs(float64(x-lastX)) > float64(gap)) {
			flushLine()
		}
		lastX, lastY, havePos = x, y, true
	}
	show := func(codes []byte) {
		line.WriteString(decoder.decode(codes))
	}

	for _, op := range ops {
		switch op := op.(type) {
		case contentstream.OpSetFont:
			fontSize = op.Size
			if page.Resources == nil {
				decoder = fontDecoder{}
				continue
			}
			font := page.Resources.Font[op.Font]
			decoded, ok := decoders[font]
			if !ok {
				decoded = newFontDecoder(font)
				decoders[font] = decoded
			}
			decoder = decoded

		case contentstream.OpShowText:
			show([]byte(op.Text))

		case contentstream.OpShowSpaceText:
			for _, text := range op.Texts {
				show(text.CharCodes)
				if text.SpaceSubtractedAfter < -100 {
					line.WriteByte(' ')
				}
			}

		case contentstream.OpMoveShowText:
			flushLine()
			show([]byte(op.Text))

		case contentstream.OpMoveSetShowText:
			flushLine()
			show([]byte(op.Text))

		case contentstream.OpTextNextLine:
			flushLine()

		case contentstream.OpTextMove:
			if op.Y != 0 {
				flushLine()
			}

		case contentstream.OpTextMoveSet:
			if op.Y != 0 {
				flushLine()
			}

		case contentstream.OpSetTextMatrix:
			setPosition(op.Matrix[4], op.Matrix[5])
		}
	}
	flushLine()
	return pageText.String(), nil
}

var unicodeSpaces = strings.NewReplacer(
	"\u00a0", " ", // non-breaking space
	"\u202f", " ", // narrow non-breaking space
	"\u2009", " ", // thin space
	"\u2002", " ", // en space
	"\u00ad", "", // soft hyphen
)

func normalizeText(text string) string {
	return strings.Join(strings.Fields(unicodeSpaces.Replace(text)), " ")
}

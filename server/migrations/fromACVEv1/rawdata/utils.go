package rawdata

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/lib/pq"
)

const has = true

type Set map[int64]bool // on choisit bool pour l'interaction avec .js

func NewSet() Set {
	return map[int64]bool{}
}

func NewSetFromSlice(keys []int64) Set {
	out := make(Set, len(keys))
	for _, key := range keys {
		out[key] = has
	}
	return out
}

func (s Set) Keys() []int64 {
	out := make([]int64, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func (s Set) Has(key int64) bool {
	_, has := s[key]
	return has
}

func (s Set) Add(key int64) {
	s[key] = has
}

type Ids []int64

func (ids Ids) AsSQL() pq.Int64Array {
	return pq.Int64Array(ids)
}

func (ids Ids) AsSet() Set {
	return NewSetFromSlice(ids)
}

// Html renvoie le html d'une ligne de tableau (incluant <tr>)
func Html(r Item, headers []Header) string {
	templateRow := `<td bgcolor="%s"> %s %s %s </td>`
	fields := make([]string, len(headers))
	for index, field := range headers {
		color := ""
		if colorV := r.TextColor(field.Field); colorV != nil {
			color = colorV.Hex()
		}
		boldIn, boldOut := "", ""
		if r.Bolds[field.Field] {
			boldIn, boldOut = "<b>", "</b>"
		}
		fields[index] = fmt.Sprintf(templateRow, color, boldIn, r.Fields.Data(field.Field).String(), boldOut)
	}
	return "<tr>" + strings.Join(fields, "") + "</tr>"
}

var (
	letterRunes  = []rune("azertyuiopqsdfghjklmwxcvbn123456789")
	specialRunes = []rune("é@!?&èïab ")
)

func RandString(n int, specialChars bool) string {
	b := make([]rune, n)
	props, maxLength := letterRunes, len(letterRunes)
	if specialChars {
		props = append(props, specialRunes...)
		maxLength += len(specialRunes)
	}
	for i := range b {
		b[i] = props[rand.Intn(maxLength)]
	}
	return string(b)
}

type StringSet map[string]bool

func (ss StringSet) ToList() []string {
	out := make([]string, 0, len(ss))
	for s := range ss {
		out = append(out, s)
	}
	return out
}

// --------------------------------------------
// ---------------- Colors --------------------
// --------------------------------------------

type Color interface {
	Hex() string
	AHex() string
}

type RGBA struct {
	R, G, B, A uint8
}

func (c RGBA) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func (c RGBA) AHex() string {
	if c.A == 0 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.A, c.R, c.G, c.B)
}

type HexColor string

func (c HexColor) Hex() string {
	if len(c) == 9 {
		return "#" + string(c[3:])
	}
	return string(c)
}

func (c HexColor) AHex() string {
	if len(c) == 7 {
		return "#ff" + string(c[1:])
	}
	return string(c)
}

// remplace une couleur vide (`nil`) par du noir
func defaultC(c Color) Color {
	if c == nil {
		return HexColor("")
	}
	return c
}

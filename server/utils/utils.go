package utils

import (
	"bytes"
	"cmp"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	letterRunes  = []rune("azertyuiopqsdfghjklmwxcvbn123456789")
	specialRunes = []rune(" é @ ! ?&èïab ")
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

var noAccent = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func removeAccents(s []byte) []byte {
	output, _, err := transform.Bytes(noAccent, s)
	if err != nil {
		return s
	}
	return output
}

func Normalize(s string) string {
	return string(removeAccents(bytes.ToLower(bytes.TrimSpace([]byte(s)))))
}

// QueryParamInt parse the query param `name` to an int
func QueryParamInt[T interface{ ~int64 | int }](c echo.Context, name string) (T, error) {
	idS := c.QueryParam(name)
	id, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID parameter %s : %s", idS, err)
	}
	return T(id), nil
}

// SQLError wraps [*pq.Error] errors only
func SQLError(err error) error {
	if errPq, ok := err.(*pq.Error); ok {
		//lint:ignore ST1005 Erreur utilisateur
		return fmt.Errorf("La requête SQL (table %s) a échoué : %s", errPq.Table, errPq)
	}
	return err
}

// InTx démarre une transaction, execute [fn] et
// COMMIT. En cas d'erreur, la transaction est ROLLBACK,
// et l'erreur renvoyé est passée à [SQLError]
func InTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return SQLError(err)
	}
	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return SQLError(err)
	}
	err = tx.Commit()
	if err != nil {
		return SQLError(err)
	}
	return nil
}

type QParam struct{ k, v string }

func QP(key, value string) QParam { return QParam{key, value} }

// BuildUrl returns http(s)://<host>/<path>?<params>
func BuildUrl(host, path string, params ...QParam) string {
	pm := url.Values{}
	for _, v := range params {
		pm.Add(v.k, v.v)
	}
	u := url.URL{
		Host:     host,
		Scheme:   "https",
		Path:     path,
		RawQuery: pm.Encode(),
	}
	if strings.HasPrefix(host, "localhost") {
		u.Scheme = "http"
	}
	return u.String()
}

func MapValues[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func MapKeysSorted[K cmp.Ordered, V any](m map[K]V) []K {
	out := MapKeys(m)
	slices.Sort(out)
	return out
}

package utils

import (
	"bytes"
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
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
	letterRunes  = []rune("azertyuiopqsdfghjklmwxcvbn0123456789AZERTYUIOPQSDFGHJKLMWXCVBN")
	specialRunes = []rune(" é @ ! ?&èïab ")
)

func RandString(n int, specialChars bool) string {
	b := make([]rune, n)
	props := letterRunes
	if specialChars {
		props = append(props, specialRunes...)
	}
	maxLength := len(props)
	for i := range b {
		b[i] = props[rand.Intn(maxLength)]
	}
	return string(b)
}

// RandPassword returns a [a-z0-9] password
// with n chars.
func RandPassword(n int) string { return RandString(n, false) }

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

// parseInt parse [v] to an int
func parseInt[T interface{ ~int64 | int }](v string) (T, error) {
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID parameter %s : %s", v, err)
	}
	return T(id), nil
}

// QueryParamInt parse the query param `name` to an int
func QueryParamInt[T interface{ ~int64 | int }](c echo.Context, name string) (T, error) {
	return parseInt[T](c.QueryParam(name))
}

func QueryParamBool(c echo.Context, name string) bool {
	value := c.QueryParam(name)
	return value != ""
}

// ReadUpload checks the file size and reads its content
func ReadUpload(fileHeader *multipart.FileHeader) (content []byte, filename string, err error) {
	const MB = 1000000
	const maxSize = 5 * MB
	if fileHeader.Size > maxSize {
		return nil, "", fmt.Errorf("file too large (%d MB)", fileHeader.Size/MB)
	}

	f, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	content, err = io.ReadAll(f)
	if err != nil {
		return nil, "", err
	}
	if int64(len(content)) != fileHeader.Size {
		return nil, "", errors.New("invalid file size")
	}
	return content, fileHeader.Filename, nil
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

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	out := make(Set[T], len(values))
	for _, v := range values {
		out.Add(v)
	}
	return out
}

func (s Set[T]) Has(key T) bool { _, ok := s[key]; return ok }

func (s Set[T]) Add(key T) { s[key] = struct{}{} }

func (s Set[T]) Delete(key T) { delete(s, key) }

func (s Set[T]) Keys() []T { return MapKeys(s) }

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

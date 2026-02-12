package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"

	"github.com/golang-jwt/jwt/v5"
)

// Encrypter is used to encrypt exposed data, such as document or
// account IDs
type Encrypter [32]byte

func NewEncrypter(key string) Encrypter { return sha256.Sum256([]byte(key)) }

type IDs interface {
	cps.IdLettreImage | fs.IdFile | cps.IdEquipier | pr.IdPersonne | in.IdInscription | ds.IdDossier | int64
}

const (
	tOther = iota
	tLettreImage
	tFile
	tEquipier
	tPersonne
	tInscription
	tDossier
)

type wrappedID[T IDs] struct {
	ID   T
	Type uint8
}

// EncryptID return a public version of the given ID.
// In particular, it is suitable to be included in URLs
func EncryptID[T IDs](key Encrypter, ID T) string {
	out, _ := encryptID(key, ID) // errors should never happen on safe data
	return out
}

func encryptID[T IDs](key Encrypter, ID T) (string, error) {
	data := wrappedID[T]{ID: ID}

	switch any(ID).(type) {
	case pr.IdPersonne:
		data.Type = tPersonne
	case in.IdInscription:
		data.Type = tInscription
	case ds.IdDossier:
		data.Type = tDossier
	case cps.IdEquipier:
		data.Type = tEquipier
	case fs.IdFile:
		data.Type = tFile
	case cps.IdLettreImage:
		data.Type = tLettreImage
	default:
		data.Type = tOther
	}

	return key.EncryptJSON(data)
}

// DecryptID returns the ID encrypted by [EncryptID]
func DecryptID[T IDs](key Encrypter, enc string) (T, error) {
	var wr wrappedID[T]
	err := key.DecryptJSON(enc, &wr)
	if err != nil {
		return 0, fmt.Errorf("invalid crypted ID: %s", err)
	}
	var typeMatch bool
	switch any(wr.ID).(type) {
	case pr.IdPersonne:
		typeMatch = wr.Type == tPersonne
	case in.IdInscription:
		typeMatch = wr.Type == tInscription
	case ds.IdDossier:
		typeMatch = wr.Type == tDossier
	case cps.IdEquipier:
		typeMatch = wr.Type == tEquipier
	case fs.IdFile:
		typeMatch = wr.Type == tFile
	case cps.IdLettreImage:
		typeMatch = wr.Type == tLettreImage
	default:
		typeMatch = wr.Type == tOther
	}
	if !typeMatch {
		return 0, fmt.Errorf("invalid ID type in %s", enc)
	}
	return wr.ID, nil
}

// EncryptJSON marshals `data`, encrypts and escape
// using [base64.RawURLEncoding]
func (key Encrypter) EncryptJSON(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	b, err = encryptAES(key[:], nil, b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecryptJSON performs the reverse operation of [EncryptJSON],
// storing the data into `dst`
func (key Encrypter) DecryptJSON(data string, dst interface{}) error {
	b, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	b, err = decryptAES(key[:], b)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, dst)
	return err
}

// ShortEncrypter provides a semi-secure ID to short password
// encrypter
type ShortEncrypter struct {
	key   [32]byte
	nonce [12]byte
}

func NewShortEncrypter(seed string) ShortEncrypter {
	nonce := md5.Sum([]byte(seed))
	key := sha256.Sum256([]byte(seed))
	return ShortEncrypter{key, [12]byte(nonce[:12])}
}

var enc32 = base32.StdEncoding.WithPadding(base32.NoPadding)

func (se ShortEncrypter) shortKey(id pr.IdPersonne) (string, error) {
	input := binary.BigEndian.AppendUint32(nil, uint32(id))
	if id <= 0xFFFF {
		input = input[2:]
	}
	bytes, err := encryptAES(se.key[:], se.nonce[:], input)
	if err != nil {
		return "", err
	}
	return enc32.EncodeToString(bytes[len(se.nonce):])[:8], nil
}

// ShortKey returns a 8 digit password with only 0-9A-Z characters
func (key ShortEncrypter) ShortKey(id pr.IdPersonne) string {
	s, _ := key.shortKey(id) // errors should never happen on safe data
	return s
}

// -------------------- shared routines --------------------

// The key argument should be the AES key,
// either 16, 24, or 32 bytes.
func encryptAES(key, nonce, data []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if nonce == nil {
		nonce = make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) <= nonceSize {
		return nil, errors.New("data too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	return plaintext, err
}

type claims[T any] interface {
	*T
	jwt.Claims
}

func VerifyJWT[T any, C claims[T]](key Encrypter, token string) (T, bool) {
	parsed, err := jwt.ParseWithClaims(token, C(new(T)), func(t *jwt.Token) (any, error) {
		return key[:], nil
	})
	if err != nil {
		return *new(T), false // not a valid token
	}
	meta := parsed.Claims.(C) // the token is valid here
	return *meta, true
}

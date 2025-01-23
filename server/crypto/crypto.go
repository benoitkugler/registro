package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	pr "registro/sql/personnes"
)

// Encrypter is used to encrypt exposed data, such as document or
// account IDs
type Encrypter [32]byte

func NewEncrypter(key string) Encrypter { return sha256.Sum256([]byte(key)) }

func (enc Encrypter) encrypt(data []byte) ([]byte, error) {
	block, _ := aes.NewCipher(enc[:])
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (enc Encrypter) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(enc[:])
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

type IDs interface {
	~int64
	pr.IdPersonne | int64
}

const (
	tOther = iota
	tPersonne
)

type wrappedID[T IDs] struct {
	ID   T
	Type uint8
}

// EncryptID return a public version of the given ID.
// In particular, it is suitable to be included in URLs
func EncryptID[T IDs](key Encrypter, ID T) string {
	out, _ := newEncryptedID(key, ID) // errors should never happen on safe data
	return out
}

func newEncryptedID[T IDs](key Encrypter, ID T) (string, error) {
	data := wrappedID[T]{ID: ID}

	switch any(ID).(type) {
	case pr.IdPersonne:
		data.Type = tPersonne
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
	b, err = key.encrypt(b)
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
	b, err = key.decrypt(b)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, dst)
	return err
}

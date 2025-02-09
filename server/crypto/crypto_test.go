package crypto

import (
	"fmt"
	"testing"

	ds "registro/sql/dossiers"
	in "registro/sql/inscriptions"
	pr "registro/sql/personnes"
	tu "registro/utils/testutils"
)

func TestEncryptionID(t *testing.T) {
	key := NewEncrypter("5s64qsd897e4q87mùlds54")
	otherKey := NewEncrypter("4")
	for i := range [200]int{} {
		v1 := 456 + 100*int64(i)
		s, err := newEncryptedID(key, v1)
		tu.AssertNoErr(t, err)

		v2, err := DecryptID[int64](key, s)
		tu.AssertNoErr(t, err)
		tu.Assert(t, v1 == v2)
		_, err = DecryptID[pr.IdPersonne](key, s)
		tu.AssertErr(t, err)

		r1 := pr.IdPersonne(v1 - 5)
		s, err = newEncryptedID(key, r1)
		tu.AssertNoErr(t, err)
		r2, err := DecryptID[pr.IdPersonne](key, s)
		tu.AssertNoErr(t, err)
		tu.Assert(t, r1 == r2)

		i1 := in.IdInscription(v1 - 5)
		s, err = newEncryptedID(key, i1)
		tu.AssertNoErr(t, err)
		i2, err := DecryptID[in.IdInscription](key, s)
		tu.AssertNoErr(t, err)
		tu.Assert(t, i1 == i2)

		d1 := ds.IdDossier(v1 - 5)
		s, err = newEncryptedID(key, d1)
		tu.AssertNoErr(t, err)
		d2, err := DecryptID[ds.IdDossier](key, s)
		tu.AssertNoErr(t, err)
		tu.Assert(t, d1 == d2)

		// expected errors
		_, err = DecryptID[int64](key, s)
		tu.AssertErr(t, err)

		_, err = DecryptID[pr.IdPersonne](otherKey, s)
		tu.AssertErr(t, err)
		_, err = DecryptID[in.IdInscription](otherKey, s)
		tu.AssertErr(t, err)
		_, err = DecryptID[ds.IdDossier](otherKey, s)
		tu.AssertErr(t, err)
	}

	fmt.Println(EncryptID(key, pr.IdPersonne(456)))
}

func TestJSON(t *testing.T) {
	type T struct {
		A int
		B string
	}
	v := T{A: 456, B: "sld"}
	var k Encrypter
	s, err := k.EncryptJSON(v)
	tu.AssertNoErr(t, err)

	var v2 T
	err = k.DecryptJSON(s, &v2)
	tu.AssertNoErr(t, err)

	tu.Assert(t, v == v2)

	err = k.DecryptJSON("", &v2)
	tu.AssertErr(t, err)

	otherKey := NewEncrypter("44")
	err = otherKey.DecryptJSON(s, &v2)
	tu.AssertErr(t, err)
}

package backoffice

import (
	"testing"

	"registro/crypto"
	tu "registro/utils/testutils"
)

func TestToken(t *testing.T) {
	ct := Controller{key: crypto.Encrypter{}}
	token, err := ct.NewToken(true, false)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(token) > 10)
}

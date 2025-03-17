package directeurs

import (
	"testing"

	"registro/crypto"
	tu "registro/utils/testutils"
)

func TestToken(t *testing.T) {
	ct := Controller{key: crypto.Encrypter{}}
	token, err := ct.NewToken(25)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(token) > 10)
}

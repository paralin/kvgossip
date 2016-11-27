package grant

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

var GranterKey *rsa.PrivateKey
var GranteeKey *rsa.PrivateKey

func mustGenerate() *rsa.PrivateKey {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return k
}

func init() {
	GranterKey = mustGenerate()
	GranteeKey = mustGenerate()
}

func TestSatisfiesGrant(t *testing.T) {
	// generate keys
}

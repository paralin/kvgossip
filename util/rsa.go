package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var PemDecodeErr error = errors.New("Pem decoding failed.")
var UnknownPublicKeyErr error = errors.New("Key is not an RSA public key.")

func ParsePublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, PemDecodeErr
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, UnknownPublicKeyErr
	}
}

func ComparePublicKey(k1 *rsa.PublicKey, k2 *rsa.PublicKey) bool {
	return k1.E == k2.E && k1.N.Cmp(k2.N) == 0
}

func ComparePublicKeyIB(k1 *rsa.PublicKey, k2b []byte) bool {
	k2, err := ParsePublicKey(k2b)
	if err != nil {
		return false
	}
	return ComparePublicKey(k1, k2)
}

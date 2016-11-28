package data

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func SignData(typ SignedData_SignedDataType, body []byte, signerKey *rsa.PrivateKey) (*SignedData, error) {
	hashed := sha256.Sum256(body)
	sig, err := rsa.SignPKCS1v15(rand.Reader, signerKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	return &SignedData{
		BodyType:  typ,
		Body:      body,
		Signature: sig,
	}, nil
}

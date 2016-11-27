package grant

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"

	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

var InvalidGrantDataTypeError error = errors.New("Invalid grant data type.")
var GrantRevocationBodyInvalid error = errors.New("Grant revocation body is invalid.")

// Parsed and validated grant data.
type ValidGrantData struct {
	// Original signed grant data
	GrantData *data.SignedData
	// Parsed public key
	PublicKey *rsa.PublicKey
	// Parsed grant
	Grant *Grant
	// Parsed grant revocation
	GrantRevocation *GrantRevocation
	// Revoked grant in the GrantRevocation
	RevokedGrant *Grant
}

// Attempt to parse and validate grant data.
func ValidateGrantData(s *data.SignedData) (*ValidGrantData, error) {
	res := &ValidGrantData{GrantData: s}

	var publicKey []byte
	var err error

	// Attempt to parse the body.
	switch s.BodyType {
	case data.SignedData_SIGNED_GRANT:
		res.Grant = &Grant{}
		err = proto.Unmarshal(s.Body, res.Grant)
		if err == nil {
			err = res.Grant.Validate()
			if err == nil {
				publicKey = res.Grant.IssuerKey
			}
		}
	case data.SignedData_SIGNED_GRANT_REVOCATION:
		res.GrantRevocation = &GrantRevocation{}
		err = proto.Unmarshal(s.Body, res.GrantRevocation)
		if err == nil {
			var revokedGrant *Grant
			revokedGrant, err = res.GrantRevocation.Validate()
			if err == nil {
				res.RevokedGrant = revokedGrant
				publicKey = res.RevokedGrant.IssuerKey
			}
		}
	default:
		err = InvalidGrantDataTypeError
	}

	if err != nil {
		return nil, err
	}

	// Parse the public key
	pkey, err := util.ParsePublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	res.PublicKey = pkey

	// Now, validate the signature of the body.
	hashed := sha256.Sum256(s.Body)
	err = rsa.VerifyPKCS1v15(pkey, crypto.SHA256, hashed[:], s.Signature)
	if err != nil {
		return nil, err
	}

	return res, nil
}

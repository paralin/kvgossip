package grant

import (
	"bytes"
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
var GrantRevocationPublicKeyMismatch error = errors.New("Grant revocation public key does not match original grant key.")

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

	var err error

	// Parse the public key
	pkey, err := util.ParsePublicKey(s.PublicKey)
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

	// Attempt to parse the body.
	switch s.BodyType {
	case data.SignedData_SIGNED_GRANT:
		res.Grant = &Grant{}
		err = proto.Unmarshal(s.Body, res.Grant)
		if err == nil {
			err = res.Grant.Validate()
		}
	case data.SignedData_SIGNED_GRANT_REVOCATION:
		res.GrantRevocation = &GrantRevocation{}
		err = proto.Unmarshal(s.Body, res.GrantRevocation)
		if err == nil {
			var revokedGrant *Grant
			revokedGrant, err = res.GrantRevocation.Validate()
			if err == nil {
				res.RevokedGrant = revokedGrant
				// Verify public keys match
				if bytes.Compare(s.PublicKey, res.GrantRevocation.Grant.PublicKey) != 0 {
					err = GrantRevocationPublicKeyMismatch
				}
			}
		}
	default:
		err = InvalidGrantDataTypeError
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Verify the grant is valid.
func (g *Grant) Validate() error {
	// TODO
	return nil
}

// Verify the grant revocation is valid.
func (g *GrantRevocation) Validate() (*Grant, error) {
	// First, parse the grant it is revoking.
	grant := g.GetGrant()
	if grant == nil {
		return nil, GrantRevocationBodyInvalid
	}

	grantData, err := ValidateGrantData(grant)
	if err != nil {
		return nil, err
	}

	if grantData.Grant == nil {
		return nil, GrantRevocationBodyInvalid
	}

	if g.RevokeTimestamp == 0 || g.RevokeTimestamp <= grantData.Grant.IssueTimestamp {
		return nil, GrantRevocationBodyInvalid
	}

	return grantData.Grant, nil
}

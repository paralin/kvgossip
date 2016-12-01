package grant

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"

	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/key"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

var InvalidGrantDataTypeError error = errors.New("Invalid grant data type.")
var GrantRevocationBodyInvalid error = errors.New("Grant revocation body is invalid.")

// ValidGrantData is parsed and validated grant data.
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

// ValidateGrantData attempts to parse and validate grant data.
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

// ValidGrants returns a slice of only the valid intermediates from a chain.
func (c *GrantAuthorizationPool) ValidGrants(cull bool, revocationChecker RevocationChecker) (valid []*ValidGrantData, revocations []*data.SignedData, invalid []*data.SignedData) {
	valid = []*ValidGrantData{}
	revocations = []*data.SignedData{}
	invalid = []*data.SignedData{}
	for _, gd := range c.GetSignedGrants() {
		revocation := revocationChecker.GetRevocation(gd.Body)
		if revocation != nil {
			invalid = append(invalid, gd)
			revocations = append(revocations, revocation)
			continue
		}
		vd, err := ValidateGrantData(gd)
		if err != nil || vd.Grant == nil {
			invalid = append(invalid, gd)
			continue
		}
		valid = append(valid, vd)
	}
	return
}

// SatisfiesGrant checks if a grant could have issued another grant.
func (c *Grant) SatisfiesGrant(gra *Grant) bool {
	return c.SubgrantAllowed &&
		bytes.Compare(gra.IssuerKey, c.IssueeKey) == 0
}

// Verify the grant is valid.
func (g *Grant) Validate() error {
	// TODO check other things... issuer / issuee keys, for example.
	if err := key.ValidatePattern(g.KeyRegex); err != nil {
		return err
	}
	if bytes.Compare(g.IssueeKey, g.IssuerKey) == 0 {
		return errors.New("IssueeKey cannot be the same as IssuerKey, this does nothing.")
	}
	if _, err := util.ParsePublicKey(g.IssueeKey); err != nil {
		return err
	}
	if _, err := util.ParsePublicKey(g.IssuerKey); err != nil {
		return err
	}
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

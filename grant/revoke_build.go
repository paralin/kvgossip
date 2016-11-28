package grant

import (
	"crypto/rsa"

	"errors"
	"time"

	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

// BuildNewGrantRevocation builds a new grant revocation given a existing grant.
func BuildNewGrantRevocation(revokedData *data.SignedData) (*GrantRevocation, error) {
	revo, err := ValidateGrantData(revokedData)
	if err != nil {
		return nil, err
	}
	revoked := revo.Grant
	if revoked == nil {
		return nil, errors.New("Signed data does not contain a grant.")
	}

	now := time.Now()
	grantTime := util.NumberToTime(revoked.IssueTimestamp)

	// If for some reason we are before the grant time, fudge it a bit.
	if now.Before(grantTime) {
		now = grantTime.Add(time.Duration(10) * time.Second)
	}

	return &GrantRevocation{
		Grant:           revokedData,
		RevokeTimestamp: util.TimeToNumber(now),
	}, nil
}

// SignGrantRevocation mints the grant revocation using the issuerKey.
func SignGrantRevocation(grantRevoke *GrantRevocation, issuerKey *rsa.PrivateKey) (*data.SignedData, error) {
	origGrant, err := grantRevoke.Validate()
	if err != nil {
		return nil, err
	}
	if !util.ComparePublicKeyIB(&issuerKey.PublicKey, origGrant.IssuerKey) {
		return nil, errors.New("Signer key and issuer key do not match.")
	}

	body, err := proto.Marshal(grantRevoke)
	if err != nil {
		return nil, err
	}

	return data.SignData(data.SignedData_SIGNED_GRANT_REVOCATION, body, issuerKey)
}

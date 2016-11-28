package grant

import (
	"crypto/rsa"

	"errors"
	"time"

	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/key"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

// BuildNewGrant builds a new grant given a issuer and issuee.
func BuildNewGrant(issuerKey *rsa.PrivateKey,
	issueeKey *rsa.PublicKey,
	pattern string,
	subgrantAllowed bool) (*Grant, error) {

	issuer, err := util.MarshalPublicKey(&issuerKey.PublicKey)
	if err != nil {
		return nil, err
	}

	issuee, err := util.MarshalPublicKey(issueeKey)
	if err != nil {
		return nil, err
	}

	err = key.ValidatePattern(pattern)
	if err != nil {
		return nil, err
	}

	return &Grant{
		KeyRegex:        pattern,
		SubgrantAllowed: subgrantAllowed,
		IssueTimestamp:  util.TimeToNumber(time.Now()),
		IssueeKey:       issuee,
		IssuerKey:       issuer,
	}, nil
}

// SignGrant mints the grant using the issuerKey.
func SignGrant(grant *Grant, issuerKey *rsa.PrivateKey) (*data.SignedData, error) {
	if err := grant.Validate(); err != nil {
		return nil, err
	}
	if !util.ComparePublicKeyIB(&issuerKey.PublicKey, grant.IssuerKey) {
		return nil, errors.New("Signer key and issuer key do not match.")
	}

	body, err := proto.Marshal(grant)
	if err != nil {
		return nil, err
	}

	return data.SignData(data.SignedData_SIGNED_GRANT, body, issuerKey)
}

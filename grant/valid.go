package grant

import (
	"bytes"
	"github.com/fuserobotics/kvgossip/key"
)

// ValidGrants returns a slice of only the valid intermediates from a chain.
func (c *GrantAuthorizationPool) ValidGrants() []*ValidGrantData {
	res := []*ValidGrantData{}
	for _, gd := range c.GetSignedGrants() {
		vd, err := ValidateGrantData(gd)
		if err != nil || vd.Grant == nil {
			continue
		}
		res = append(res, vd)
	}
	return res
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

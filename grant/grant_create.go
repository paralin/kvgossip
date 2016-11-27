package grant

import (
	"crypto/rsa"
	"errors"
	//  note: requires pending pr #2
	// "github.com/fuserobotics/kvgossip/util"
	// "github.com/twmb/algoimpl/go/graph"
)

var InvalidArgumentError error = errors.New("Arguments are invalid.")

// Returns a slice of only the valid intermediates from a chain.
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

// Verify attempts to find chains of grants from the list of signed grants
// which enable the requested action to the performed. The chains will have
// the root key as the first element, always, and the actor key as the last element.
//
// WARNING: this doesn't do any revocation checking. You should run the resultant
// list of chains through checks for revocation before trusting any of the chains.
func VerifyGrantAuthorization(actor, root *rsa.PublicKey, pool *GrantAuthorizationPool) (chains [][]*ValidGrantData, err error) {
	if root == nil || pool == nil || actor == nil {
		return nil, InvalidArgumentError
	}

	// res := [][]*ValidGrantData{}
	// grants := pool.ValidGrants()
	// grants = append([]*ValidGrantData{{PublicKey: root}}, grants...)

	/*
		recurseFindRoot(targetGrant)

		// check for grant given from root key.
		if util.ComparePublicKey(root, targetGrant.PublicKey) {
			res = append(res, []*ValidGrantData{targetGrant})
			// Since the parent is the root key, none of the intermediates can work.
			return res, nil
		}

		// find the

		g := graph.New(graph.Directed)
		rootNode := g.MakeNode()
		*rootNode.Value = root

		for _, r := range intermediates {
			root.N.c
		}
	*/

	return nil, nil
}

package tx

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"

	"errors"
	"time"

	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/key"
	"github.com/fuserobotics/kvgossip/util"
)

// Context for a verify grant authorization request.
// Finds chains from root -> actor
type verifyGrantAuthorization struct {
	Root *rsa.PublicKey
	Pool *grant.GrantAuthorizationPool

	// Set of grants, value is visited.
	VisitedGrants map[*grant.ValidGrantData]bool

	// Set of resultant paths
	Result [][]*grant.ValidGrantData
	// Current path
	Chain []*grant.ValidGrantData
}

// attempt to find the next link in the chain
func (vga *verifyGrantAuthorization) evaluate(curr *grant.ValidGrantData, target *Transaction) {
	// Mark this as visited, defer unmarking it
	vga.VisitedGrants[curr] = true
	vga.Chain = append(vga.Chain, curr)
	defer func() {
		vga.VisitedGrants[curr] = false
		vga.Chain = vga.Chain[:len(vga.Chain)-1]
	}()

	// If we fulfill the target, we have found a chain.
	// We fufill the target if we are the root key and the target is signed by the root key,
	// OR we are NOT the root key and the grant satisfies the target.
	if (curr.Grant == nil &&
		util.ComparePublicKeyIB(curr.PublicKey, target.Verification.SignerPublicKey)) ||
		(curr.Grant != nil &&
			target.SatisfiedBy(curr.Grant)) {
		path := []*grant.ValidGrantData{}
		copy(vga.Chain, path)
		vga.Result = append(vga.Result, path)
		return
	}

	for gra, visited := range vga.VisitedGrants {
		if visited {
			continue
		}

		// Root key will have Grant = nil
		// The next step in the chain will have:
		// issuer_key = root, or current satisfies target
		// Current satisfies target = issuer + issuee matches, regex allows it, subgrant_allowed = true
		if (curr.Grant == nil && util.ComparePublicKey(curr.PublicKey, gra.PublicKey)) ||
			curr.Grant.SatisfiesGrant(gra.Grant) {
			vga.evaluate(gra, target)
		}
	}
}

type VerifyGrantAuthorizationResult struct {
	Chains        [][]*grant.ValidGrantData
	ValidGrants   []*grant.ValidGrantData
	Revocations   []*data.SignedData
	InvalidGrants []*data.SignedData
}

// Verify attempts to find chains of grants from the list of signed grants
// which enable the requested action to the performed. The chains will have
// the root key as the first element, always, and the actor key as the last element.
func VerifyGrantAuthorization(target *Transaction, root *rsa.PublicKey, pool *grant.GrantAuthorizationPool, revocationChecker grant.RevocationChecker) *VerifyGrantAuthorizationResult {
	if root == nil || pool == nil || target == nil {
		return nil
	}

	valid, revocations, invalid := pool.ValidGrants(true, revocationChecker)
	grants := valid
	rootGrant := &grant.ValidGrantData{PublicKey: root}

	vga := &verifyGrantAuthorization{
		Root:          root,
		Chain:         []*grant.ValidGrantData{},
		Pool:          pool,
		VisitedGrants: make(map[*grant.ValidGrantData]bool),
		Result:        [][]*grant.ValidGrantData{},
	}

	vga.VisitedGrants[rootGrant] = false
	for _, gr := range grants {
		vga.VisitedGrants[gr] = false
	}

	vga.evaluate(rootGrant, target)

	return &VerifyGrantAuthorizationResult{
		Chains:        vga.Result,
		InvalidGrants: invalid,
		ValidGrants:   valid,
		Revocations:   revocations,
	}
}

// SatisfiedBy checks if a grant could have issued a transaction.
func (trx *Transaction) SatisfiedBy(gra *grant.Grant) bool {
	return key.KeyPatternContains(gra.KeyRegex, trx.Key)
}

// Initial check of a transaction message.
func (trx *Transaction) Validate() error {
	if trx.Key == "" {
		return errors.New("Key must be set on transaction.")
	}
	if trx.TransactionType != Transaction_TRANSACTION_SET {
		return errors.New("Only set transactions are supported.")
	}
	if trx.Verification == nil {
		return errors.New("Transaction verification is required.")
	}
	if err := trx.Verification.Validate(); err != nil {
		return err
	}
	// Verify signature matches.
	pubKey, err := util.ParsePublicKey(trx.Verification.SignerPublicKey)
	if err != nil {
		return err
	}
	hashBuf := make([]byte, len(trx.Value)+8)
	copy(hashBuf, trx.Value)
	binary.LittleEndian.PutUint64(hashBuf[len(trx.Value):], trx.Verification.Timestamp)
	hash := sha256.Sum256(trx.Value)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], trx.Verification.ValueSignature)
}

func (trx *TransactionVerification) Validate() error {
	if trx.Grant == nil || len(trx.Grant.SignedGrants) == 0 {
		return errors.New("Transaction must have a grant pool.")
	}
	stamp := util.NumberToTime(int64(trx.Timestamp))
	if stamp.After(time.Now()) {
		return errors.New("Timestamp is in the future.")
	}
	return nil
}

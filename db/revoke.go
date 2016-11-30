package db

import (
	"bytes"
	"errors"

	"github.com/boltdb/bolt"
	dn "github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/grant"
	txna "github.com/fuserobotics/kvgossip/tx"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

var NoValidRevocationErr error = errors.New("Revocation data did not contain a valid revocation.")

// Check for a revocation. grantData is the encoded body of the signed grant.
func (kvg *KVGossipDB) GetRevocation(grantData []byte) (dr *dn.SignedData) {
	key := util.HexSha256(grantData)
	err := kvg.DB.View(func(tx *bolt.Tx) error {
		bkt := kvg.GetRevocationBucket(tx)
		data := bkt.Get([]byte(key))
		if len(data) == 0 {
			return nil
		}
		dr = &dn.SignedData{}
		return proto.Unmarshal(data, dr)
	})
	if err != nil {
		return nil
	}
	return
}

type applyRevocationChecker struct {
	Revocation   *dn.SignedData
	RevokedGrant []byte
}

func (arc *applyRevocationChecker) GetRevocation(grantData []byte) *dn.SignedData {
	if bytes.Compare(grantData, arc.RevokedGrant) == 0 {
		return arc.Revocation
	}
	return nil
}

func (kvg *KVGossipDB) ApplyRevocation(sd *dn.SignedData) error {
	if sd.BodyType != dn.SignedData_SIGNED_GRANT_REVOCATION {
		return NoValidRevocationErr
	}
	vgd, err := grant.ValidateGrantData(sd)
	if err != nil {
		return err
	}
	if vgd.GrantRevocation == nil {
		return NoValidRevocationErr
	}
	rev := vgd.GrantRevocation
	// Check if we have already applied this revocation.
	if kvg.GetRevocation(rev.Grant.Body) != nil {
		return nil
	}

	return kvg.DB.Update(func(tx *bolt.Tx) error {
		// Iterate over existing metadata.
		checker := &applyRevocationChecker{
			Revocation:   sd,
			RevokedGrant: rev.Grant.Body,
		}
		numDeleted := 0
		err := kvg.DB.View(func(vtx *bolt.Tx) error {
			return kvg.ForeachKeyVerification(vtx, func(k string, v *txna.TransactionVerification) error {
				valid, _, _ := v.Grant.ValidGrants(false, checker)
				if len(valid) == 0 {
					log.Warnf("Deleting key %s due to new revocation.", k)
					kvg.PurgeKey(tx, k)
					numDeleted++
				}
				return nil
			})
		})
		if err != nil {
			return err
		}

		if numDeleted > 0 {
			if err := kvg.UpdateOverallHash(tx); err != nil {
				return err
			}
		}

		// Put the new revocation.
		bkt := kvg.GetRevocationBucket(tx)
		key := util.HexSha256(rev.Grant.Body)
		bd, err := proto.Marshal(sd)
		if err != nil {
			return err
		}
		return bkt.Put([]byte(key), bd)
	})
}

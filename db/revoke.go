package db

import (
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

// Check for a revocation. grantData is the encoded body of the signed grant.
func (kvg *KVGossipDB) GetRevocation(grantData []byte) *grant.GrantRevocation {
	key := util.Base64Sha256(grantData)
	var gr *grant.GrantRevocation
	err := kvg.DB.View(func(tx *bolt.Tx) error {
		bkt := kvg.GetRevocationBucket(tx)
		data := bkt.Get([]byte(key))
		if len(data) == 0 {
			return nil
		}
		gr = &grant.GrantRevocation{}
		return proto.Unmarshal(data, gr)
	})
	if err != nil {
		return nil
	}
	return gr
}

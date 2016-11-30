package db

import (
	"github.com/boltdb/bolt"
	dn "github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/golang/protobuf/proto"
)

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

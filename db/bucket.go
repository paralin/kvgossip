package db

import (
	"github.com/boltdb/bolt"
)

func (kvg *KVGossipDB) GetDataBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("data"))
}

func (kvg *KVGossipDB) GetMetaBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("meta"))
}

func (kvg *KVGossipDB) GetRevocationBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("revocations"))
}

func (kvg *KVGossipDB) GetDataHashBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("datahashes"))
}

func (kvg *KVGossipDB) GetGlobalBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("global"))
}

func (kvg *KVGossipDB) GetGrantBucket(tx *bolt.Tx) *bolt.Bucket {
	return GetOrEnsureBucket(tx, []byte("grants"))
}

func (kvg *KVGossipDB) ensureBuckets() error {
	return kvg.DB.Update(func(tx *bolt.Tx) error {
		kvg.GetDataBucket(tx)
		kvg.GetMetaBucket(tx)
		kvg.GetRevocationBucket(tx)
		kvg.GetDataHashBucket(tx)
		kvg.GetGlobalBucket(tx)
		kvg.GetGrantBucket(tx)
		return nil
	})
}

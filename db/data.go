package db

import (
	"github.com/boltdb/bolt"
	kvgtx "github.com/fuserobotics/kvgossip/tx"
	"github.com/golang/protobuf/proto"
)

// Check for data at key.
func (kvg *KVGossipDB) GetKeyData(tx *bolt.Tx, key string) []byte {
	bkt := kvg.GetDataBucket(tx)
	return bkt.Get([]byte(key))
}

func (kvg *KVGossipDB) UpdateKeyData(tx *bolt.Tx, key string, value []byte) error {
	bkt := kvg.GetDataBucket(tx)
	return bkt.Put([]byte(key), value)
}

// Attempt to pull a full transaction for a key.
func (kvg *KVGossipDB) GetTransaction(tx *bolt.Tx, key string) *kvgtx.Transaction {
	meta := kvg.GetKeyVerification(tx, key)
	if meta == nil {
		return nil
	}
	return &kvgtx.Transaction{
		Key:             key,
		TransactionType: kvgtx.Transaction_TRANSACTION_SET,
		Value:           kvg.GetKeyData(tx, key),
		Verification:    meta,
	}
}

// Set the verification for a key.
func (kvg *KVGossipDB) UpdateKeyVerification(tx *bolt.Tx, key string, verification *kvgtx.TransactionVerification) error {
	vd, err := proto.Marshal(verification)
	if err != nil {
		return err
	}
	bkt := kvg.GetMetaBucket(tx)
	return bkt.Put([]byte(key), vd)
}

// Get the verification for a key.
func (kvg *KVGossipDB) GetKeyVerification(tx *bolt.Tx, key string) *kvgtx.TransactionVerification {
	bkt := kvg.GetMetaBucket(tx)
	data := bkt.Get([]byte(key))
	if len(data) == 0 {
		return nil
	}
	res := &kvgtx.TransactionVerification{}
	if err := proto.Unmarshal(data, res); err != nil {
		return nil
	}
	return res
}

// Get the verification for all keys.
func (kvg *KVGossipDB) ForeachKeyVerification(tx *bolt.Tx, fee func(k string, v *kvgtx.TransactionVerification) error) error {
	bkt := kvg.GetMetaBucket(tx)
	return bkt.ForEach(func(k, v []byte) error {
		res := &kvgtx.TransactionVerification{}
		if err := proto.Unmarshal(v, res); err != nil {
			return nil
		}
		return fee(string(k), res)
	})
}

func (kvg *KVGossipDB) ForeachKeyHash(tx *bolt.Tx, fee func(k string, v []byte) error) error {
	bkt := kvg.GetDataHashBucket(tx)
	return bkt.ForEach(func(k, v []byte) error {
		return fee(string(k), v)
	})
}

func (kvg *KVGossipDB) PurgeKey(tx *bolt.Tx, key string) error {
	// Delete metadata.
	buckets := []*bolt.Bucket{
		kvg.GetDataBucket(tx),
		kvg.GetMetaBucket(tx),
		kvg.GetDataHashBucket(tx),
	}
	for _, bkt := range buckets {
		bkt.Delete([]byte(key))
	}
	return nil
}

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

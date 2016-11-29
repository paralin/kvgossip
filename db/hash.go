package db

import (
	"crypto/sha256"
	"errors"

	"github.com/boltdb/bolt"
)

var TreeHashKeyName []byte = []byte("treeHash")

// Update the key hash for a key.
func (kvg *KVGossipDB) UpdateKeyHash(tx *bolt.Tx, key string, keyData []byte) error {
	hash := sha256.Sum256(keyData)
	bkt := kvg.GetDataHashBucket(tx)
	return bkt.Put([]byte(key), hash[:])
}

func (kvg *KVGossipDB) UpdateOverallHash(tx *bolt.Tx) error {
	if !tx.Writable() {
		return errors.New("Transaction must be writable.")
	}

	bkt := kvg.GetDataHashBucket(tx)
	numKeys := bkt.Stats().KeyN
	buf := make([]byte, numKeys*32)

	i := 0
	c := bkt.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		for bi, byt := range v {
			// we expect 32 byte hashes, assert this tho
			if bi == 32 {
				break
			}
			// cast the byte to an int, xor with index, cast to byte
			buf[(i*32)+bi] = byte((int(byt) ^ i) % 255)
		}
		i++
	}

	overallHash := sha256.Sum256(buf)
	gbkt := kvg.GetGlobalBucket(tx)
	kvg.TreeHash = overallHash[:]
	return gbkt.Put(TreeHashKeyName, kvg.TreeHash)
}

func (kvg *KVGossipDB) ensureTreeHash() {
	kvg.DB.Update(func(tx *bolt.Tx) error {
		gbkt := kvg.GetGlobalBucket(tx)
		data := gbkt.Get(TreeHashKeyName)
		if len(data) == 0 {
			kvg.UpdateOverallHash(tx)
		} else {
			kvg.TreeHash = make([]byte, len(data))
			copy(kvg.TreeHash, data)
		}
		return nil
	})
}

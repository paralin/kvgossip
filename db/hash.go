package db

import (
	"bytes"
	"crypto/sha256"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/util"
)

var TreeHashKeyName []byte = []byte("treeHash")

func (kvg *KVGossipDB) GetKeyHash(tx *bolt.Tx, key string) []byte {
	bkt := kvg.GetDataHashBucket(tx)
	data := bkt.Get([]byte(key))
	res := make([]byte, len(data))
	copy(res, data)
	return data
}

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
	// we have to add 1 for the new key
	numKeys++
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
	if len(kvg.TreeHash) != 0 && bytes.Compare(overallHash[:], kvg.TreeHash) != 0 {
		log.Infof("Overall tree hash updated -> %s", util.HashToString(overallHash[:]))
	}

	gbkt := kvg.GetGlobalBucket(tx)
	kvg.TreeHash = overallHash[:]
	kvg.TreeHashChanged <- kvg.TreeHash
	return gbkt.Put(TreeHashKeyName, kvg.TreeHash)
}

func (kvg *KVGossipDB) ensureTreeHash() {
	kvg.DB.Update(func(tx *bolt.Tx) error {
		kvg.UpdateOverallHash(tx)
		return nil
	})
}

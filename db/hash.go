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

// var EmptyHash []byte = hashKeyData([]byte{})

func (kvg *KVGossipDB) GetKeyHash(tx *bolt.Tx, key string) []byte {
	bkt := kvg.GetDataHashBucket(tx)
	data := bkt.Get([]byte(key))
	res := make([]byte, len(data))
	copy(res, data)
	return data
}

func hashKeyData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Update the key hash for a key.
func (kvg *KVGossipDB) UpdateKeyHash(tx *bolt.Tx, key string, keyData []byte) ([]byte, error) {
	bkt := kvg.GetDataHashBucket(tx)

	var hashSlice []byte
	if len(keyData) > 0 {
		hashSlice = hashKeyData(keyData)
	}
	return hashSlice, bkt.Put([]byte(key), hashSlice)
}

func (kvg *KVGossipDB) GetOverallHash() []byte {
	var data []byte
	kvg.DB.View(func(tx *bolt.Tx) error {
		bkt := kvg.GetGlobalBucket(tx)
		data = bkt.Get(TreeHashKeyName)
		return nil
	})
	return data
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

	oldHash := kvg.GetOverallHash()
	overallHash := sha256.Sum256(buf)
	if len(oldHash) != 0 && bytes.Compare(overallHash[:], oldHash) != 0 {
		log.Infof("Overall tree hash updated -> %s", util.HashToString(overallHash[:]))
	}

	gbkt := kvg.GetGlobalBucket(tx)
	kvg.TreeHashChanged <- overallHash[:]
	return gbkt.Put(TreeHashKeyName, overallHash[:])
}

func (kvg *KVGossipDB) ensureTreeHash() {
	kvg.DB.Update(func(tx *bolt.Tx) error {
		kvg.UpdateOverallHash(tx)
		return nil
	})
}

package db

import (
	"github.com/boltdb/bolt"
)

func GetOrEnsureBucket(tx *bolt.Tx, key []byte) *bolt.Bucket {
	if tx.Writable() {
		b, _ := tx.CreateBucketIfNotExists(key)
		return b
	}
	return tx.Bucket(key)
}

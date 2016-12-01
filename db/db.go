package db

import (
	"time"

	"github.com/boltdb/bolt"
)

// BoltDB backed database for KVGossip.
type KVGossipDB struct {
	DB *bolt.DB
	// TreeHash        []byte
	TreeHashChanged chan []byte
}

func OpenDB(dbPath string) (*KVGossipDB, error) {
	res := &KVGossipDB{}
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	res.DB = db
	res.TreeHashChanged = make(chan []byte, 10)
	res.ensureBuckets()
	res.ensureTreeHash()
	return res, nil
}

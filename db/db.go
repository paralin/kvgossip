package db

import (
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// BoltDB backed database for KVGossip.
type KVGossipDB struct {
	DB              *bolt.DB
	TreeHashChanged chan []byte

	keyChanged         chan *KeySubscriptionEvent
	keySubscriptions   map[string][]*KeySubscription
	keySubscriptionMtx sync.Mutex

	closeOnce          sync.Once
	closeSubscriptions chan bool
	lastCloseError     error

	applyMutex sync.Mutex
}

func OpenDB(dbPath string) (*KVGossipDB, error) {
	res := &KVGossipDB{}
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	res.DB = db
	res.TreeHashChanged = make(chan []byte, 10)
	res.keyChanged = make(chan *KeySubscriptionEvent, 10)
	res.keySubscriptions = make(map[string][]*KeySubscription)
	res.closeSubscriptions = make(chan bool, 1)
	res.ensureBuckets()
	res.ensureTreeHash()

	res.DB.Update(func(tx *bolt.Tx) error {
		return res.UpdateOverallHash(tx)
	})

	go res.handleSubscriptions()

	return res, nil
}

func (db *KVGossipDB) Close() error {
	var err error
	db.closeOnce.Do(func() {
		db.closeSubscriptions <- true
		err = db.DB.Close()
	})
	if err == nil {
		err = db.lastCloseError
	}
	db.lastCloseError = err
	return err
}

package db

import (
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/tx"
)

// BoltDB backed database for KVGossip.
type KVGossipDB struct {
	DB *bolt.DB
	// TreeHash        []byte
	TreeHashChanged chan []byte
	KeyChanged      chan *tx.Transaction
	keyChangedChans []chan<- *tx.Transaction
	keyChangedMtx   sync.Mutex
	closeChan       chan bool
}

func OpenDB(dbPath string) (*KVGossipDB, error) {
	res := &KVGossipDB{}
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	res.DB = db
	res.TreeHashChanged = make(chan []byte, 10)
	res.KeyChanged = make(chan *tx.Transaction, 100)
	res.closeChan = make(chan bool, 1)
	res.ensureBuckets()
	res.ensureTreeHash()
	go res.transactionStreamHandler()
	return res, nil
}

func (db *KVGossipDB) TransactionStreamSubscribe(ch chan<- *tx.Transaction) {
	db.keyChangedMtx.Lock()
	defer db.keyChangedMtx.Unlock()

	db.keyChangedChans = append(db.keyChangedChans, ch)
}

func (db *KVGossipDB) TransactionStreamUnsubscribe(ch chan<- *tx.Transaction) {
	db.keyChangedMtx.Lock()
	defer db.keyChangedMtx.Unlock()

	for i, chi := range db.keyChangedChans {
		if chi == ch {
			db.keyChangedChans = append(db.keyChangedChans[:i], db.keyChangedChans[i+1:]...)
			return
		}
	}
}

func (db *KVGossipDB) transactionStreamHandler() {
	kcc := db.KeyChanged
	for trans := range kcc {
		db.keyChangedMtx.Lock()
		for _, ch := range db.keyChangedChans {
			select {
			case ch <- trans:
			default:
			}
		}
		db.keyChangedMtx.Unlock()
	}
}

func (db *KVGossipDB) Close() error {
	if db.KeyChanged != nil {
		close(db.KeyChanged)
		db.KeyChanged = nil
	}
	for _, kcc := range db.keyChangedChans {
		close(kcc)
	}
	db.keyChangedChans = nil
	select {
	case db.closeChan <- true:
	default:
	}
	return db.DB.Close()
}

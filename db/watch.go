package db

import (
	"sync"

	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/tx"
)

type KeySubscription struct {
	db       *KVGossipDB
	key      string
	disposed bool

	lastVal   *tx.Transaction
	subMtx    sync.Mutex
	unsubOnce sync.Once
	chans     []chan<- *tx.Transaction
}

func (ks *KeySubscription) Unsubscribe() {
	ks.unsubOnce.Do(func() {
		ks.subMtx.Lock()
		ks.disposed = true
		ks.db.removeSubscription(ks)
		for _, ch := range ks.chans {
			close(ch)
		}
		ks.chans = nil
		ks.subMtx.Unlock()
	})
}

func (ks *KeySubscription) Value() *tx.Transaction {
	ks.subMtx.Lock()
	defer ks.subMtx.Unlock()

	return ks.lastVal
}

func (ks *KeySubscription) Changes(ch chan<- *tx.Transaction) {
	ks.subMtx.Lock()
	defer ks.subMtx.Unlock()

	if ks.disposed {
		close(ch)
		return
	}

	ks.chans = append(ks.chans, ch)
	select {
	case ch <- ks.lastVal:
	default:
	}
}

func (db *KVGossipDB) SubscribeKey(key string) *KeySubscription {
	var lastVal *tx.Transaction
	db.DB.View(func(t *bolt.Tx) error {
		lastVal = db.GetTransaction(t, key)
		return nil
	})
	ks := &KeySubscription{
		db:      db,
		key:     key,
		lastVal: lastVal,
	}
	db.addSubscription(ks)
	return ks
}

func (ks *KeySubscription) next(trans *tx.Transaction) {
	ks.subMtx.Lock()
	for _, ch := range ks.chans {
		select {
		case ch <- trans:
		default:
		}
	}
	ks.lastVal = trans
	ks.subMtx.Unlock()
}

func (db *KVGossipDB) addSubscription(ks *KeySubscription) {
	db.keySubscriptionMtx.Lock()
	db.keySubscriptions[ks.key] = append(db.keySubscriptions[ks.key], ks)
	db.keySubscriptionMtx.Unlock()
}

func (db *KVGossipDB) removeSubscription(ks *KeySubscription) {
	db.keySubscriptionMtx.Lock()
	arr := db.keySubscriptions[ks.key]
	for i, v := range arr {
		if v == ks {
			arr[i] = arr[len(arr)-1]
			arr = arr[:len(arr)-1]
			break
		}
	}
	db.keySubscriptions[ks.key] = arr
	db.keySubscriptionMtx.Unlock()
}

func (db *KVGossipDB) handleSubscriptions() {
	for {
		select {
		case <-db.closeSubscriptions:
			return
		case trans := <-db.keyChanged:
			db.keySubscriptionMtx.Lock()
			arr := db.keySubscriptions[trans.Key]
			for _, n := range arr {
				n.next(trans)
			}
			db.keySubscriptionMtx.Unlock()
		}
	}
}

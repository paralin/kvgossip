package db

import (
	"sync"

	"github.com/boltdb/bolt"
	kkey "github.com/fuserobotics/kvgossip/key"
	"github.com/fuserobotics/kvgossip/tx"
)

type KeySubscriptionEvent struct {
	Key         string
	Transaction *tx.Transaction
	UpdatedHash []byte
}

type KeySubscription struct {
	db         *KVGossipDB
	key        string
	keyPattern []string
	disposed   bool

	lastVal   *KeySubscriptionEvent
	subMtx    sync.Mutex
	unsubOnce sync.Once
	chans     []chan<- *KeySubscriptionEvent
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

func (ks *KeySubscription) Value() *KeySubscriptionEvent {
	if ks.key == "list" {
		return nil
	}

	ks.subMtx.Lock()
	defer ks.subMtx.Unlock()

	return ks.lastVal
}

func (ks *KeySubscription) Changes(ch chan<- *KeySubscriptionEvent) {
	ks.subMtx.Lock()
	defer ks.subMtx.Unlock()

	if ks.disposed {
		close(ch)
		return
	}

	ks.chans = append(ks.chans, ch)
	if ks.key != "list" {
		select {
		case ch <- ks.lastVal:
		default:
		}
	}
}

func (db *KVGossipDB) SubscribeKey(key string) *KeySubscription {
	if key == "list" {
		return nil
	}

	lastVal := &KeySubscriptionEvent{}
	db.DB.View(func(t *bolt.Tx) error {
		lastVal.Key = lastVal.Transaction.Key
		lastVal.Transaction = db.GetTransaction(t, key)
		lastVal.UpdatedHash = db.GetKeyHash(t, key)
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

func (db *KVGossipDB) SubscribeKeyPattern(keyPattern []string) *KeySubscription {
	ks := &KeySubscription{
		db:         db,
		key:        "list",
		keyPattern: keyPattern,
	}
	db.addSubscription(ks)
	return ks
}

func (ks *KeySubscription) next(trans *KeySubscriptionEvent) {
	ks.subMtx.Lock()
	for _, pat := range ks.keyPattern {
		if !kkey.KeyPatternContains(pat, trans.Transaction.Key) {
			return
		}
	}
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
			larr := db.keySubscriptions["list"]
			for _, n := range arr {
				n.next(trans)
			}
			for _, n := range larr {
				n.next(trans)
			}
			db.keySubscriptionMtx.Unlock()
		}
	}
}

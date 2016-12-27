package db

import (
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/tx"
)

func (db *KVGossipDB) ApplyTransaction(trans *tx.Transaction) error {
	db.applyMutex.Lock()
	defer db.applyMutex.Unlock()

	// Put all the valid grants
	for _, grant := range trans.Verification.Grant.SignedGrants {
		if err := db.PutGrant(grant); err != nil {
			return err
		}
	}
	err := db.DB.Update(func(tx *bolt.Tx) error {
		// Update key hash
		if err := db.UpdateKeyHash(tx, trans.Key, trans.Value); err != nil {
			return err
		}
		// Update value
		if err := db.UpdateKeyData(tx, trans.Key, trans.Value); err != nil {
			return err
		}
		// Update verification
		if err := db.UpdateKeyVerification(tx, trans.Key, trans.Verification); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = db.DB.Update(func(tx *bolt.Tx) error {
		return db.UpdateOverallHash(tx)
	})
	if err != nil {
		return err
	}
	db.keyChanged <- trans
	return nil
}

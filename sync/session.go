package sync

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/tx"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/fuserobotics/kvgossip/version"
	"golang.org/x/net/context"
)

var SessionMessageTimeout time.Duration = time.Duration(5) * time.Second
var TimeoutErr error = errors.New("No messages before timeout.")
var DuplicateSessionErr error = errors.New("Duplicate sync session.")

const MaxSyncSpinCount int = 3

type SyncSessionStream interface {
	Context() context.Context
	Send(*SyncSessionMessage) error
	Recv() (*SyncSessionMessage, error)
}

type SyncSessionState struct {
	ReceivedGlobalHash bool

	RemoteGlobalHash []byte
	RemoteKeyHashes  map[string][]byte
	// Number of times we've tried the sync process
	SyncSpinCount int
}

// An instance of a sync session
type SyncSession struct {
	// The database
	DB *db.KVGossipDB
	// Root key
	RootKey *rsa.PublicKey
	// Dedupe
	Dedupe *SyncSessionDedupe
	// Did we initiate the session
	Initiator bool
	// State
	State SyncSessionState
	// When the session ended
	Ended chan bool
	// Timeout ticker
	Timeout *time.Timer
	// Error to return
	Error error
	// Remote message channel
	RecvChan chan *SyncSessionMessage
	// Stream
	Stream SyncSessionStream
	// Stuff to call on termination
	Cleanup []func()
	// DB read transaction
	ReadTransaction *bolt.Tx
	// Disconnect now
	DisconnectNow bool
}

func NewSyncSession(d *db.KVGossipDB, dd *SyncSessionDedupe, initiator bool, rootKey *rsa.PublicKey) *SyncSession {
	return &SyncSession{
		DB:        d,
		Initiator: initiator,
		Dedupe:    dd,
		Ended:     make(chan bool, 1),
		RecvChan:  make(chan *SyncSessionMessage),
		RootKey:   rootKey,
	}
}

func (ss *SyncSession) waitForResponseWithTimeout() (*SyncSessionMessage, error) {
	ss.Timeout.Reset(SessionMessageTimeout)
	select {
	case <-ss.Timeout.C:
		return nil, TimeoutErr
	case msg, ok := <-ss.RecvChan:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	}
}

func (ss *SyncSession) runSyncSession() error {
	if err := ss.sendSyncGlobalHash(); err != nil {
		return err
	}
	// Wait for the global hash.
	msg, err := ss.waitForResponseWithTimeout()
	if err != nil {
		return err
	}
	if err := ss.assertGlobalHash(msg); err != nil || ss.DisconnectNow {
		return err
	}

	// Iterate through our local keys and attempt to compare them.
	err = ss.DB.ForeachKeyVerification(ss.ReadTransaction, func(k string, v *tx.TransactionVerification) error {
		err := ss.Stream.Send(&SyncSessionMessage{
			SyncKeyHash: &SyncKeyHash{
				Signature: v.ValueSignature,
				Key:       k,
				Timestamp: v.Timestamp,
			},
		})
		if err != nil {
			return err
		}
		msg, err := ss.waitForResponseWithTimeout()
		if err != nil {
			return err
		}
		if msg.SyncKeyHash != nil {
			return nil
		} else if msg.SyncKey != nil {
			if msg.SyncKey.RequestKey != k {
				return errors.New("Request key did not match last sent key hash.")
			}
			if msg.SyncKey.Transaction == nil {
				return ss.sendKeyTransaction(k)
			} else {
				trans := msg.SyncKey.Transaction
				err = trans.Validate()
				if err != nil {
					return err
				}
				if trans.Key != k {
					return errors.New("Key mismatch in SyncKey body.")
				}

				if trans.Verification.Timestamp < v.Timestamp {
					return errors.New("Peer offered key with older timestamp than ours.")
				}

				syncKeyResult, err := ss.handleIncomingTransaction(trans, msg.SyncKey)
				if err != nil {
					return err
				}
				return ss.Stream.Send(&SyncSessionMessage{
					SyncKeyResult: syncKeyResult,
				})
			}
		} else {
			return errors.New("Expected SyncKey or SyncKeyHash response.")
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (ss *SyncSession) handleIncomingTransaction(trans *tx.Transaction, sk *SyncKey) (*SyncKeyResult, error) {
	// validate
	res := tx.VerifyGrantAuthorization(trans, ss.RootKey, sk.Transaction.Verification.Grant, ss.DB)
	syncKeyResult := &SyncKeyResult{
		Revocations: res.Revocations,
		UpdatedKey:  trans.Key,
	}
	if len(res.Chains) > 0 {
		log.Debugf("Received valid new value for key %s, timestamp %v.", trans.Key, util.NumberToTime(int64(trans.Verification.Timestamp)))
		if err := ss.DB.ApplyTransaction(trans); err != nil {
			return nil, err
		}
	} else {
		log.Debugf("For key %s, peer sent valid transaction with no valid grants (%d local revocations).", trans.Key, len(res.Revocations))
	}
	return syncKeyResult, nil
}

func (ss *SyncSession) sendKeyTransaction(key string) error {
	log.Debugf("Sending data for key %s to remote peer.", key)
	tx := ss.ReadTransaction
	trx := ss.DB.GetTransaction(tx, key)
	if trx == nil {
		return errors.New("Unable to pull that transaction from the db.")
	}
	err := ss.Stream.Send(&SyncSessionMessage{
		SyncKey: &SyncKey{
			RequestKey:  key,
			Transaction: trx,
		},
	})
	if err != nil {
		return err
	}
	msg, err := ss.waitForResponseWithTimeout()
	if err != nil {
		return err
	}
	if msg.SyncKeyResult == nil {
		return errors.New("Expected SyncKeyResult after SyncKey.")
	}
	if msg.SyncKeyResult.UpdatedKey != trx.Key {
		return errors.New("Key mismatch in SyncKeyResult.")
	}
	numRev := len(msg.SyncKeyResult.Revocations)
	for i, revocation := range msg.SyncKeyResult.Revocations {
		log.Debug("Applying revocation %d/%d from peer...", i+1, numRev)
		if err := ss.DB.ApplyRevocation(revocation); err != nil {
			return err
		}
	}
	return nil
}

// Gets a new fresh read transaction.
func (ss *SyncSession) buildReadTransaction() (*bolt.Tx, error) {
	if ss.ReadTransaction != nil {
		ss.ReadTransaction.Commit()
		ss.ReadTransaction = nil
	}
	tx, err := ss.DB.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	ss.Cleanup = append(ss.Cleanup, func() {
		if ss.ReadTransaction == tx {
			tx.Commit()
			ss.ReadTransaction = nil
		}
	})
	ss.ReadTransaction = tx
	return tx, nil
}

func (ss *SyncSession) getReadTransaction() (*bolt.Tx, error) {
	if ss.ReadTransaction != nil {
		return ss.ReadTransaction, nil
	}
	return ss.buildReadTransaction()
}

func (ss *SyncSession) handleGlobalHash(msg *SyncGlobalHash) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	// We re-received a global hash
	if ss.State.ReceivedGlobalHash {
		if ss.Initiator {
			return errors.New("Did not expect to receive a global hash twice.")
		} else {
			log.Debug("Remote re-started sync.")
			ss.State.SyncSpinCount++
			ss.State.ReceivedGlobalHash = false
			if ss.State.SyncSpinCount > MaxSyncSpinCount {
				return errors.New("Too many sync attempts.")
			}
		}
	}
	// we force a new read transaction
	_, err := ss.buildReadTransaction()
	if err != nil {
		return err
	}
	ss.State.RemoteKeyHashes = make(map[string][]byte)
	ss.State.RemoteGlobalHash = msg.GlobalTreeHash
	if !ss.State.ReceivedGlobalHash {
		nonce := msg.HostNonce
		if !ss.Dedupe.TryRegisterSession(nonce, ss) {
			return DuplicateSessionErr
		}
		ss.Cleanup = append(ss.Cleanup, func() {
			ss.Dedupe.UnregisterSession(nonce)
		})
	}
	ss.State.ReceivedGlobalHash = true
	if !ss.Initiator {
		if err := ss.sendSyncGlobalHash(); err != nil {
			return err
		}
	}
	if bytes.Compare(ss.State.RemoteGlobalHash, ss.DB.TreeHash) == 0 {
		log.Debug("Remote hash matches local hash, disconnecting.")
		ss.DisconnectNow = true
	} else {
		log.Debug("Remote hash %s != local %s, attempting sync.",
			util.HashToString(ss.State.RemoteGlobalHash),
			util.HashToString(ss.DB.TreeHash))
	}
	return nil
}

func (ss *SyncSession) assertGlobalHash(msg *SyncSessionMessage) error {
	if msg.SyncGlobalHash != nil {
		return ss.handleGlobalHash(msg.SyncGlobalHash)
	} else {
		if !ss.State.ReceivedGlobalHash {
			return errors.New("Remote did not offer global hash before continuing.")
		}
	}
	return nil
}

func (ss *SyncSession) sendSyncGlobalHash() error {
	return ss.Stream.Send(&SyncSessionMessage{
		SyncGlobalHash: &SyncGlobalHash{
			GlobalTreeHash:  ss.DB.TreeHash,
			KvgossipVersion: version.KVGossipVersion,
			HostNonce:       ss.Dedupe.LocalNonce,
		},
	})
}

func (ss *SyncSession) handleMessage(msg *SyncSessionMessage) error {
	if err := ss.assertGlobalHash(msg); err != nil {
		return err
	}
	return nil
}

func (ss *SyncSession) SyncSession(stream SyncSessionStream) error {
	ss.Stream = stream
	log.Debug("Started sync session with peer.")
	ss.Timeout = time.NewTimer(SessionMessageTimeout)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF && ss.Error == nil {
					ss.Error = err
				}
				break
			}
			ss.RecvChan <- msg
		}
		close(ss.RecvChan)
	}()

	defer func() {
		log.Debug("Sync session complete.")
		for _, cleanup := range ss.Cleanup {
			cleanup()
		}
		ss.Ended <- true
	}()

	if ss.Initiator {
		err := ss.runSyncSession()
		if err != nil && ss.Error == nil {
			ss.Error = err
		}
	} else {
		// If we're not the initiator, wait for messages.
	Loop:
		for {
			select {
			case msg, ok := <-ss.RecvChan:
				if !ok {
					break Loop
				}
				ss.Timeout.Reset(SessionMessageTimeout)
				err := ss.handleMessage(msg)
				if ss.Error != nil {
					break Loop
				} else {
					if err != nil {
						ss.Error = err
						break Loop
					}
				}
			case <-ss.Timeout.C:
				log.Warn("Timeout for peer.")
				if ss.Error == nil {
					ss.Error = TimeoutErr
				}
			}
			if ss.DisconnectNow {
				break
			}
		}
	}

	if ss.Error != nil {
		log.Warnf("Session ended with error %v", ss.Error)
	}
	return ss.Error
}

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
	ReceivedGlobalHash          bool
	ReceivedGlobalHashFirstTime bool

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
	err = ss.DB.DB.View(func(readTransaction *bolt.Tx) error {
		return ss.DB.ForeachKeyVerification(readTransaction, func(k string, v *tx.TransactionVerification) error {
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

					syncKeyResult, err := ss.handleIncomingTransaction(msg.SyncKey)
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
	})
	if err != nil {
		return err
	}
	err = ss.Stream.Send(&SyncSessionMessage{
		SyncKeyHash: &SyncKeyHash{
			Key: "",
		},
	})
	if err != nil {
		return err
	}
	// Await incoming new keys.
	for {
		msg, err := ss.waitForResponseWithTimeout()
		if err != nil {
			return err
		}
		// SyncKeyHash indicates end of new keys.
		if msg.SyncKeyHash != nil {
			// Allow re-starting again
			ss.State.ReceivedGlobalHash = false
			break
		}
		sk := msg.SyncKey
		if sk == nil {
			return errors.New("Expected SyncKey or empty SyncKeyHash.")
		}
		if sk.Transaction == nil {
			return errors.New("Expected transactions, not requests in new key phase.")
		}
		res, err := ss.handleIncomingTransaction(sk)
		if err != nil {
			return err
		}
		err = ss.Stream.Send(&SyncSessionMessage{
			SyncKeyResult: res,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ss *SyncSession) handleIncomingTransaction(sk *SyncKey) (*SyncKeyResult, error) {
	trans := sk.Transaction
	// validate
	res := tx.VerifyGrantAuthorization(trans, ss.RootKey, ss.DB)
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
	var trx *tx.Transaction
	ss.DB.DB.View(func(tx *bolt.Tx) error {
		trx = ss.DB.GetTransaction(tx, key)
		return nil
	})
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
	ss.State.RemoteKeyHashes = make(map[string][]byte)
	ss.State.RemoteGlobalHash = msg.GlobalTreeHash
	if !ss.State.ReceivedGlobalHashFirstTime {
		nonce := msg.HostNonce
		if !ss.Dedupe.TryRegisterSession(nonce, ss) {
			return DuplicateSessionErr
		}
		ss.Cleanup = append(ss.Cleanup, func() {
			ss.Dedupe.UnregisterSession(nonce)
		})
		ss.State.ReceivedGlobalHashFirstTime = true
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
		log.Debugf("Remote hash %s != local %s, attempting sync.",
			util.HashToString(ss.State.RemoteGlobalHash),
			util.HashToString(ss.DB.TreeHash))
	}
	return nil
}

func (ss *SyncSession) assertGlobalHash(msg *SyncSessionMessage) error {
	if msg.SyncGlobalHash != nil {
		ss.State.RemoteKeyHashes = make(map[string][]byte)
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
	/* After the initial global hash, this will basically be skipped. */
	if err := ss.assertGlobalHash(msg); err != nil || ss.DisconnectNow || msg.SyncGlobalHash != nil {
		return err
	}

	/* We then expect a series of SyncKeyHash. */
	if msg.SyncKeyHash == nil {
		return errors.New("SyncKeyHash expected.")
	}

	skh := msg.SyncKeyHash
	if len(skh.Key) == 0 {
		// We now need to send keys that we have locally but not remotely.
		err := ss.DB.DB.View(func(readTransaction *bolt.Tx) error {
			return ss.DB.ForeachKeyHash(readTransaction, func(k string, v []byte) error {
				if _, ok := ss.State.RemoteKeyHashes[k]; ok {
					return nil
				}
				return ss.sendKeyTransaction(k)
			})
		})
		if err != nil {
			return err
		}
		// reset
		ss.State.SyncSpinCount++
		ss.State.ReceivedGlobalHash = false
		// indicate to remote that we are ready
		return ss.Stream.Send(&SyncSessionMessage{
			SyncKeyHash: &SyncKeyHash{
				Key: "",
			},
		})
	} else {
		if _, ok := ss.State.RemoteKeyHashes[skh.Key]; ok {
			return errors.New("Received key twice (not allowed).")
		}
		ss.State.RemoteKeyHashes[skh.Key] = skh.Signature
		if len(skh.Signature) != 32 {
			return errors.New("Signature length not 32, invalid.")
		}
		var localSig []byte
		ss.DB.DB.View(func(tx *bolt.Tx) error {
			localSig = ss.DB.GetKeyHash(tx, skh.Key)
			return nil
		})
		// If we don't have the signature locally.
		if localSig == nil || len(localSig) == 0 {
			// Request they send us the key.
			if err := ss.requestRemoteKey(skh.Key); err != nil {
				return err
			}
		} else if bytes.Compare(localSig, skh.Signature) == 0 {
			// We agree, send a agreement.
			return ss.agreeRemoteKey(skh.Key)
		} else {
			// we need to compare timestamps, pull the verification out of the db.
			var tver *tx.TransactionVerification
			ss.DB.DB.View(func(readTransaction *bolt.Tx) error {
				tver = ss.DB.GetKeyVerification(readTransaction, skh.Key)
				return nil
			})
			if tver.Timestamp > skh.Timestamp {
				return ss.sendKeyTransaction(skh.Key)
			} else {
				if err := ss.requestRemoteKey(skh.Key); err != nil {
					return err
				}
			}
		}
		// Wait for the response to the request for data.
		msg, err := ss.waitForResponseWithTimeout()
		if err != nil {
			return err
		}
		if msg.SyncKey == nil || msg.SyncKey.Transaction == nil {
			return errors.New("Expected SyncKey after key request.")
		}
		res, err := ss.handleIncomingTransaction(msg.SyncKey)
		if err != nil {
			return err
		}
		return ss.Stream.Send(&SyncSessionMessage{
			SyncKeyResult: res,
		})
	}
}

func (ss *SyncSession) requestRemoteKey(key string) error {
	log.Debugf("Requesting key %s from peer.", key)
	return ss.Stream.Send(&SyncSessionMessage{
		SyncKey: &SyncKey{
			RequestKey: key,
		},
	})
}

func (ss *SyncSession) agreeRemoteKey(key string) error {
	log.Debugf("Agreeing on key %s from peer.", key)
	return ss.Stream.Send(&SyncSessionMessage{
		SyncKeyHash: &SyncKeyHash{
			Timestamp: 1,
		},
	})
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
		for ss.Error == nil && !ss.DisconnectNow && ss.State.SyncSpinCount < MaxSyncSpinCount {
			err := ss.runSyncSession()
			if err != nil && ss.Error == nil {
				ss.Error = err
			}
			ss.State.SyncSpinCount++
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

package sync

import (
	"bytes"
	"errors"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/fuserobotics/kvgossip/version"
	"golang.org/x/net/context"
)

var SessionMessageTimeout time.Duration = time.Duration(5) * time.Second
var TimeoutErr error = errors.New("No messages before timeout.")
var DuplicateSessionErr error = errors.New("Duplicate sync session.")

type SyncSessionStream interface {
	Context() context.Context
	Send(*SyncSessionMessage) error
	Recv() (*SyncSessionMessage, error)
}

type SyncSessionState struct {
	ReceivedGlobalHash bool

	RemoteGlobalHash []byte
}

// An instance of a sync session
type SyncSession struct {
	// The database
	DB *db.KVGossipDB
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

func NewSyncSession(d *db.KVGossipDB, dd *SyncSessionDedupe, initiator bool) *SyncSession {
	return &SyncSession{
		DB:        d,
		Initiator: initiator,
		Dedupe:    dd,
		Ended:     make(chan bool, 1),
		RecvChan:  make(chan *SyncSessionMessage),
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
	// we force a new read transaction
	_, err := ss.buildReadTransaction()
	if err != nil {
		return err
	}
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

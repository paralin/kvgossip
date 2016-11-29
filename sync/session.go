package sync

import (
	"errors"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
	"golang.org/x/net/context"
)

var SessionMessageTimeout time.Duration = time.Duration(10) * time.Second
var TimeoutErr error = errors.New("No messages before timeout.")

type SyncSessionStream interface {
	Context() context.Context
	Send(*SyncSessionMessage) error
	Recv() (*SyncSessionMessage, error)
}

// An instance of a sync session
type SyncSession struct {
	// The database
	DB *db.KVGossipDB
	// Did we initiate the session
	Initiator bool
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
}

func NewSyncSession(d *db.KVGossipDB, initiator bool) *SyncSession {
	return &SyncSession{
		DB:        d,
		Initiator: initiator,
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
	err := ss.Stream.Send(&SyncSessionMessage{
		SyncGlobalHash: &SyncGlobalHash{
			KvgossipVersion: "test",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ss *SyncSession) handleMessage(msg *SyncSessionMessage) error {
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
				log.Debug("Received message from peer.")
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
		}
	}

	if ss.Error != nil {
		log.Warnf("Session ended with error %v", ss.Error)
	}
	return ss.Error
}

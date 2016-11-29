package sync

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
	"golang.org/x/net/context"
)

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
}

func NewSyncSession(d *db.KVGossipDB, initiator bool) *SyncSession {
	return &SyncSession{
		DB:        d,
		Initiator: initiator,
		Ended:     make(chan bool, 1),
	}
}

func (ss *SyncSession) SyncSession(stream SyncSessionStream) error {
	log.Debug("Started sync session with peer.")
	defer func() {
		log.Debug("Sync session complete.")
		ss.Ended <- true
	}()

	if ss.Initiator {
		stream.Send(&SyncSessionMessage{
			SyncGlobalHash: &SyncGlobalHash{
				GlobalTreeHash:  []byte("hi"),
				KvgossipVersion: "latest",
			},
		})
	} else {
		for {
			_, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}
			log.Debug("Received message from peer.")
		}
	}

	return nil
}

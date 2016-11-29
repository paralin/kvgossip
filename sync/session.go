package sync

import (
	"github.com/fuserobotics/kvgossip/db"
)

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

func (ss *SyncSession) SyncSession(stream SyncService_SyncSessionServer) error {
	// ctx := stream.Context()
	defer func() {
		ss.Ended <- true
	}()

	return nil
}

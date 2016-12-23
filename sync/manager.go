package sync

import (
	"crypto/rsa"
	"fmt"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type queuedSync struct {
	Peer      string
	PeerNonce string
}

// State machine. Attempts to run sync sessions against peers.
type SyncManager struct {
	Dedupe          *SyncSessionDedupe
	SyncServicePort int

	db *db.KVGossipDB

	stopped        bool
	stopChan       chan bool
	serverDisabled bool

	sessions         map[int]*SyncSession
	sessionIdCounter int
	sessionsMtx      sync.RWMutex

	rootKey   *rsa.PublicKey
	syncQueue chan *queuedSync
}

func NewSyncManager(d *db.KVGossipDB, servicePort int, rootKey *rsa.PublicKey) *SyncManager {
	return &SyncManager{
		db:              d,
		rootKey:         rootKey,
		stopped:         true,
		Dedupe:          NewSyncSessionDedupe(),
		stopChan:        make(chan bool, 1),
		SyncServicePort: servicePort,
		sessions:        make(map[int]*SyncSession),
		syncQueue:       make(chan *queuedSync, 50),
	}
}

// Start the sync manager.
func (sm *SyncManager) Start() {
	if !sm.stopped {
		return
	}
	sm.stopped = false
	go sm.syncLoop()
}

func (sm *SyncManager) SetServerDisabled(disabled bool) {
	if !sm.stopped {
		return
	}
	sm.serverDisabled = disabled
}

func (sm *SyncManager) QueueSync(peer, peernonce string) {
	if len(peer) == 0 {
		return
	}
	if len(peernonce) != 0 {
		if sm.Dedupe.HasSession(peernonce) {
			return
		}
	}
	sm.syncQueue <- &queuedSync{
		Peer:      peer,
		PeerNonce: peernonce,
	}
}

func (sm *SyncManager) Connect(peer string) error {
	conn, err := grpc.Dial(peer, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := NewSyncServiceClient(conn)
	ss := NewSyncSession(sm.db, sm.Dedupe, true, sm.rootKey)
	sm.startSyncSession(ss)
	stream, err := client.SyncSession(context.Background())
	if err != nil {
		return err
	}
	return ss.SyncSession(stream)
}

func (sm *SyncManager) startSyncSession(ss *SyncSession) {
	sm.sessionsMtx.Lock()
	sm.sessionIdCounter++
	id := sm.sessionIdCounter
	sm.sessions[id] = ss
	sm.sessionsMtx.Unlock()
	ss.Cleanup = append(ss.Cleanup, func() {
		sm.sessionsMtx.Lock()
		delete(sm.sessions, id)
		sm.sessionsMtx.Unlock()
	})
}

func (sm *SyncManager) syncLoop() {
	log.Debugf("Starting sync manager loop...")
	defer func() {
		log.Debugf("Exiting sync manager loop...")
	}()

	grpcServer := grpc.NewServer()
	RegisterSyncServiceServer(grpcServer, sm)

	if sm.SyncServicePort <= 0 {
		sm.serverDisabled = true
	}
	if !sm.serverDisabled {
		go func() {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", sm.SyncServicePort))
			if err != nil {
				log.Warnf("Unable to start sync service, %v", err)
				sm.serverDisabled = true
				return
			}
			log.Infof("Sync service listening on port %d", sm.SyncServicePort)
			grpcServer.Serve(lis)
		}()
		defer func() {
			grpcServer.Stop()
		}()
	}

	for {
	QueueSelect:
		select {
		case <-sm.stopChan:
			return
		case queued := <-sm.syncQueue:
			if len(queued.PeerNonce) != 0 && sm.Dedupe.HasSession(queued.PeerNonce) {
				break QueueSelect
			}
			log.Debugf("Initiating sync session with %s from queue.", queued.Peer)
			go sm.Connect(queued.Peer)
		}
	}
}

func (sm *SyncManager) SyncSession(stream SyncService_SyncSessionServer) error {
	session := NewSyncSession(sm.db, sm.Dedupe, false, sm.rootKey)
	sm.startSyncSession(session)
	return session.SyncSession(stream)
}

func (sm *SyncManager) Stop() {
	if sm.stopped {
		return
	}
	sm.stopped = true
	sm.stopChan <- true
}

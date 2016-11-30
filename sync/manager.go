package sync

import (
	"crypto/rsa"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// State machine. Attempts to run sync sessions against peers.
type SyncManager struct {
	db               *db.KVGossipDB
	stopped          bool
	stopChan         chan bool
	serverDisabled   bool
	syncServicePort  int
	sessions         map[int]*SyncSession
	sessionIdCounter int
	dedupe           *SyncSessionDedupe
	rootKey          *rsa.PublicKey
}

func NewSyncManager(d *db.KVGossipDB, servicePort int, rootKey *rsa.PublicKey) *SyncManager {
	return &SyncManager{
		db:              d,
		rootKey:         rootKey,
		stopped:         true,
		dedupe:          NewSyncSessionDedupe(),
		stopChan:        make(chan bool, 1),
		syncServicePort: servicePort,
		sessions:        make(map[int]*SyncSession),
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

func (sm *SyncManager) Connect(peer string) error {
	conn, err := grpc.Dial(peer, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	client := NewSyncServiceClient(conn)
	ss := NewSyncSession(sm.db, sm.dedupe, true, sm.rootKey)
	sm.startSyncSession(ss)
	stream, err := client.SyncSession(context.Background())
	if err != nil {
		return err
	}
	return ss.SyncSession(stream)
}

func (sm *SyncManager) startSyncSession(ss *SyncSession) {
	sm.sessionIdCounter++
	id := sm.sessionIdCounter
	sm.sessions[id] = ss
	go func() {
		<-ss.Ended
		delete(sm.sessions, id)
	}()
}

func (sm *SyncManager) syncLoop() {
	log.Debugf("Starting sync manager loop...")
	defer func() {
		log.Debugf("Exiting sync manager loop...")
	}()

	grpcServer := grpc.NewServer()
	RegisterSyncServiceServer(grpcServer, sm)

	if sm.syncServicePort <= 0 {
		sm.serverDisabled = true
	}
	if !sm.serverDisabled {
		go func() {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", sm.syncServicePort))
			if err != nil {
				log.Warnf("Unable to start sync service, %v", err)
				sm.serverDisabled = true
				return
			}
			log.Infof("Sync service listening on port %d", sm.syncServicePort)
			grpcServer.Serve(lis)
		}()
		defer func() {
			grpcServer.Stop()
		}()
	}

	for {
		select {
		case <-sm.stopChan:
			return
		}
	}
}

func (sm *SyncManager) SyncSession(stream SyncService_SyncSessionServer) error {
	session := NewSyncSession(sm.db, sm.dedupe, false, sm.rootKey)
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

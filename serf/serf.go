package serf

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/sync"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/serf/client"
)

type SerfManager struct {
	SyncManager  *sync.SyncManager
	TreeHash     []byte
	TreeHashChan <-chan []byte
	SerfAddress  string

	stopped          bool
	stopChan         chan bool
	serfMessageChan  chan map[string]interface{}
	serfStreamHandle client.StreamHandle
	serfClient       *client.RPCClient
}

func NewSerfManager(sm *sync.SyncManager, serfRpc string) *SerfManager {
	return &SerfManager{
		SyncManager: sm,
		SerfAddress: serfRpc,
		stopped:     true,
		stopChan:    make(chan bool, 1),
	}
}

func (sm *SerfManager) Start() {
	if !sm.stopped {
		return
	}
	sm.stopped = false
	go sm.syncLoop()
}

func (sm *SerfManager) syncLoop() {
	log.Debugf("Starting serf sync loop...")
	defer func() {
		log.Debugf("Exiting serf sync loop...")
		if sm.serfClient != nil {
			sm.serfClient.Close()
			sm.serfClient = nil
		}
	}()

	for {
		if sm.serfClient == nil {
			if sm.initSerfStreamShouldQuit() {
				return
			}
			continue
		}
		select {
		case <-sm.stopChan:
			return
		case th := <-sm.TreeHashChan:
			sm.TreeHash = th
			go sm.broadcastTreeHash()
		case m, ok := <-sm.serfMessageChan:
			if !ok {
				if sm.initSerfStreamShouldQuit() {
					return
				}
				break
			}
			if err := sm.handleSerfMessage(m); err != nil {
				log.Warnf("Unable to handle serf message, %v", err)
			}
		}
	}
}

func (sm *SerfManager) initSerfStreamShouldQuit() bool {
	err := sm.initSerfStream()
	if err != nil {
		select {
		case <-time.After(time.Duration(5) * time.Second):
			return false
		case <-sm.stopChan:
			return true
		}
	}
	return false
}

func (sm *SerfManager) handleSerfMessage(m map[string]interface{}) error {
	payloadInter, ok := m["Payload"]
	if !ok {
		return nil
	}
	payload, ok := payloadInter.([]byte)

	if !ok || m["Name"] != "kvgossip" || m["Event"] != "query" {
		return nil
	}
	idInter, ok := m["ID"]
	if !ok {
		return nil
	}
	id, ok := idInter.(int64)
	if !ok {
		return nil
	}

	mess := &SerfQueryMessage{}
	if err := proto.Unmarshal(payload, mess); err != nil {
		return err
	}
	if mess.HostNonce == sm.SyncManager.Dedupe.LocalNonce {
		return nil
	}
	if mess.TreeHash != nil {
		return sm.handleRemoteTreeHash(uint64(id), mess)
	}
	return nil
}

func (sm *SerfManager) handleRemoteTreeHash(messageId uint64, mess *SerfQueryMessage) error {
	th := mess.TreeHash
	if len(th.TreeHash) != 32 {
		return fmt.Errorf("Ignoring invalid tree hash broadcast with len %d != 32", len(th.TreeHash))
	}
	if bytes.Compare(th.TreeHash, sm.TreeHash) == 0 || sm.serfClient == nil {
		return nil
	}
	resp := &SerfQueryMessage{
		HostNonce: sm.SyncManager.Dedupe.LocalNonce,
		TreeHash: &SerfTreeHashBroadcast{
			SyncPort: uint32(sm.SyncManager.SyncServicePort),
			TreeHash: sm.TreeHash,
		},
	}
	bi, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	sm.serfClient.Respond(messageId, bi)
	return nil
}

func (sm *SerfManager) broadcastTreeHash() {
	if sm.serfClient == nil {
		return
	}
	msg := &SerfQueryMessage{
		TreeHash: &SerfTreeHashBroadcast{
			TreeHash: sm.TreeHash,
		},
		HostNonce: sm.SyncManager.Dedupe.LocalNonce,
	}
	pay, err := proto.Marshal(msg)
	if err != nil {
		log.Warnf("Error marshalling msg: %v", err)
		return
	}
	rch := make(chan client.NodeResponse, 30)
	log.Debug("Initiating tree hash sweep.")
	err = sm.serfClient.Query(&client.QueryParam{
		RequestAck: false,
		Name:       "kvgossip",
		Payload:    pay,
		RespCh:     rch,
		Timeout:    time.Duration(5) * time.Second,
	})
	if err != nil {
		log.Warnf("Error sending query: %v", err)
		return
	}
	members, err := sm.serfClient.Members()
	if err != nil {
		log.Warnf("Fetching members failed, %v", err)
		return
	}
	for m := range rch {
		mess := &SerfQueryMessage{}
		if err := proto.Unmarshal(m.Payload, mess); err != nil || mess.TreeHash == nil {
			log.Warnf("Cannot understand response from %s", m.From)
			continue
		}
		th := mess.TreeHash.TreeHash
		if len(th) != 32 {
			log.Warnf("%s responded with invalid treehash length %d", m.From, len(th))
			continue
		}
		if mess.TreeHash.SyncPort < 2000 {
			log.Warnf("%s responded with invalid syncport %d", m.From, mess.TreeHash.SyncPort)
			continue
		}
		if bytes.Compare(th, sm.TreeHash) != 0 {
			if sm.serfClient == nil {
				return
			}
			for _, memb := range members {
				if memb.Name == m.From {
					log.Debugf("Queuing sync to %s ip %s from query response.", memb.Name, memb.Addr.String())
					sm.SyncManager.QueueSync(fmt.Sprintf("%s:%d", memb.Addr.String(), mess.TreeHash.SyncPort), mess.HostNonce)
					break
				}
			}
		}
	}
	log.Debug("Completed tree hash sweep.")
}

func (sm *SerfManager) initSerfStream() (initError error) {
	defer func() {
		if initError != nil {
			log.Warnf("Unable to init serf query stream, %v", initError)
			if sm.serfClient != nil {
				sm.serfClient.Close()
				sm.serfClient = nil
			}
		}
	}()
	if sm.serfClient != nil {
		sm.serfClient.Close()
		sm.serfClient = nil
	}
	log.Debugf("Connecting to serf at %s...", sm.SerfAddress)
	rc, err := client.NewRPCClient(sm.SerfAddress)
	if err != nil {
		return err
	}
	sm.serfClient = rc
	sm.serfMessageChan = make(chan map[string]interface{})
	log.Debug("Initing serf query stream...")
	sh, err := sm.serfClient.Stream("query:kvgossip", sm.serfMessageChan)
	if err != nil {
		return err
	}
	sm.serfStreamHandle = sh
	return nil
}

func (sm *SerfManager) Stop() {
	if sm.stopped {
		return
	}
	sm.stopped = true
	sm.stopChan <- true
}

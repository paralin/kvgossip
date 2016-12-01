package sync

import (
	"github.com/fuserobotics/kvgossip/util"
	"sync"
)

// Attempts to prevent multiple sessions with a remote host.
type SyncSessionDedupe struct {
	LocalNonce     string
	activeSessions map[string]*SyncSession
	sessionMutex   sync.Mutex
}

func NewSyncSessionDedupe() *SyncSessionDedupe {
	return &SyncSessionDedupe{
		LocalNonce:     util.RandStringRunes(10),
		activeSessions: make(map[string]*SyncSession),
	}
}

func (ss *SyncSessionDedupe) HasSession(nonce string) bool {
	ss.sessionMutex.Lock()
	defer ss.sessionMutex.Unlock()
	_, ok := ss.activeSessions[nonce]
	return ok
}

func (ss *SyncSessionDedupe) TryRegisterSession(key string, sess *SyncSession) bool {
	ss.sessionMutex.Lock()
	defer ss.sessionMutex.Unlock()

	_, ok := ss.activeSessions[key]
	if ok {
		return false
	}
	ss.activeSessions[key] = sess
	return true
}

func (ss *SyncSessionDedupe) UnregisterSession(key string) {
	ss.sessionMutex.Lock()
	defer ss.sessionMutex.Unlock()

	delete(ss.activeSessions, key)
}

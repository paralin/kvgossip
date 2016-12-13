package client

import (
	"sync"

	"github.com/fuserobotics/kvgossip/tx"
)

type KeySubscriptionState struct {
	HasValue bool
	Value    *tx.Transaction
	Dirty    bool
}

type KeySubscription struct {
	Unsubscribe func()

	disposed      bool
	interest      *KeyInterest
	stateChans    []chan *KeySubscriptionState
	stateChanLock sync.Mutex
}

func (ks *KeySubscription) State() *KeySubscriptionState {
	return ks.interest.State()
}

func (ks *KeySubscription) Changes() <-chan *KeySubscriptionState {
	ks.stateChanLock.Lock()
	defer ks.stateChanLock.Unlock()
	ch := make(chan *KeySubscriptionState, 1)
	ks.stateChans = append(ks.stateChans, ch)
	ch <- ks.State()
	return ch
}

func (ks *KeySubscription) nextState(state *KeySubscriptionState) {
	ks.stateChanLock.Lock()
	defer ks.stateChanLock.Unlock()

	for _, cha := range ks.stateChans {
		select {
		case cha <- state:
		default:
		}
	}
}

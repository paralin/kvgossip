package client

import (
	"sync"

	"github.com/fuserobotics/kvgossip/tx"
)

type KeySubscription struct {
	client   *Client
	interest *interest

	disposed         bool
	disposeOnce      sync.Once
	disposeMtx       sync.Mutex
	disposeCallbacks []func(ks *KeySubscription)

	lastState    *KeySubscriptionState
	lastStateMtx sync.RWMutex

	state chan *KeySubscriptionState

	stateSubs    []chan<- *KeySubscriptionState
	stateSubsMtx sync.Mutex
}

type KeySubscriptionState struct {
	/** Do we have a value yet?
	 *  If this is true and Transaction is nil,
	 *  there's no value for that key yet.
	 */
	HasValue bool

	// Dirty indicates the value may have changed.
	Dirty bool

	// Transaction. Do not modify!
	Transaction *tx.Transaction
}

func newKeySubscription(client *Client, interest *interest) *KeySubscription {
	return &KeySubscription{
		client:   client,
		interest: interest,
		state:    make(chan *KeySubscriptionState, 1),
	}
}

func (ks *KeySubscription) next(state *KeySubscriptionState) {
	ks.state <- state
}

func (ks *KeySubscription) updateLoop() {
	for {
		select {
		case state, ok := <-ks.state:
			if !ok {
				return
			}
			ks.disposeMtx.Lock()
			if ks.disposed {
				return
			}

			ks.lastStateMtx.Lock()
			ks.lastState = state
			ks.lastStateMtx.Unlock()

			ks.stateSubsMtx.Lock()
			for _, ch := range ks.stateSubs {
				select {
				case ch <- state:
				default:
				}
			}
			ks.stateSubsMtx.Unlock()

			ks.disposeMtx.Unlock()
		}
	}
}

func (ks *KeySubscription) State() *KeySubscriptionState {
	ks.lastStateMtx.RLock()
	defer ks.lastStateMtx.RUnlock()
	return ks.lastState
}

func (ks *KeySubscription) Changes(ch chan<- *KeySubscriptionState) {
	ks.disposeMtx.Lock()
	defer ks.disposeMtx.Unlock()

	if ks.disposed {
		close(ch)
		return
	}

	ks.stateSubsMtx.Lock()
	select {
	case ch <- ks.State():
	default:
	}
	ks.stateSubs = append(ks.stateSubs, ch)
	ks.stateSubsMtx.Unlock()
}

func (ks *KeySubscription) OnDisposed(cb func(ks *KeySubscription)) {
	ks.disposeMtx.Lock()
	defer ks.disposeMtx.Unlock()

	if ks.disposed {
		cb(ks)
		return
	}

	ks.disposeCallbacks = append(ks.disposeCallbacks, cb)
}

func (ks *KeySubscription) Unsubscribe() {
	ks.disposeOnce.Do(func() {
		ks.disposeMtx.Lock()
		ks.disposed = true
		ks.disposeMtx.Unlock()

		for _, cb := range ks.disposeCallbacks {
			go cb(ks)
		}
		ks.stateSubsMtx.Lock()
		for _, ch := range ks.stateSubs {
			close(ch)
		}
		ks.stateSubs = nil
		ks.stateSubsMtx.Unlock()
		ks.disposeCallbacks = nil
	})
}

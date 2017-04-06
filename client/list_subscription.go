package client

import (
	"sync"
)

type KeyListSubscription struct {
	client   *Client
	interest *listInterest

	disposed         bool
	disposeOnce      sync.Once
	disposeMtx       sync.Mutex
	disposeCallbacks []func(ks *KeyListSubscription)

	eventChan    chan *KeyListSubscriptionEvent
	eventSubs    []chan<- *KeyListSubscriptionEvent
	eventSubsMtx sync.Mutex
}

type KeyListSubscriptionEvent_Action int

const (
	KeyList_Add KeyListSubscriptionEvent_Action = iota
	KeyList_Update
	KeyList_Remove
)

type KeyListSubscriptionEvent struct {
	Action KeyListSubscriptionEvent_Action
	Key    string
	Hash   []byte
}

func newKeyListSubscription(client *Client, interest *listInterest) *KeyListSubscription {
	return &KeyListSubscription{
		client:    client,
		interest:  interest,
		eventChan: make(chan *KeyListSubscriptionEvent, 10),
	}
}

func (ks *KeyListSubscription) next(state *KeyListSubscriptionEvent) {
	ks.eventChan <- state
}

func (ks *KeyListSubscription) updateLoop() {
	for {
		select {
		case event, ok := <-ks.eventChan:
			if !ok {
				return
			}
			ks.disposeMtx.Lock()
			if ks.disposed {
				return
			}

			ks.eventSubsMtx.Lock()
			for _, ch := range ks.eventSubs {
				select {
				case ch <- event:
				default:
				}
			}
			ks.eventSubsMtx.Unlock()

			ks.disposeMtx.Unlock()
		}
	}
}

func (ks *KeyListSubscription) State() map[string]*KeyListEntry {
	ks.interest.stateMtx.Lock()
	defer ks.interest.stateMtx.Unlock()

	// Copy the state, in case they mess with it.
	result := make(map[string]*KeyListEntry)
	for key, v := range ks.interest.state {
		result[key] = v.Copy()
	}

	return result
}

func (ks *KeyListSubscription) Events(ch chan<- *KeyListSubscriptionEvent, includeInitial bool) {
	ks.disposeMtx.Lock()
	defer ks.disposeMtx.Unlock()

	if ks.disposed {
		close(ch)
		return
	}

	if includeInitial {
		ks.interest.stateMtx.Lock()
		for key, val := range ks.interest.state {
			ch <- &KeyListSubscriptionEvent{
				Action: KeyList_Add,
				Key:    key,
				Hash:   val.Hash,
			}
		}
	}

	ks.eventSubsMtx.Lock()
	ks.eventSubs = append(ks.eventSubs, ch)
	ks.eventSubsMtx.Unlock()

	if includeInitial {
		ks.interest.stateMtx.Unlock()
	}
}

func (ks *KeyListSubscription) OnDisposed(cb func(ks *KeyListSubscription)) {
	ks.disposeMtx.Lock()
	defer ks.disposeMtx.Unlock()

	if ks.disposed {
		cb(ks)
		return
	}

	ks.disposeCallbacks = append(ks.disposeCallbacks, cb)
}

func (ks *KeyListSubscription) Unsubscribe() {
	ks.disposeOnce.Do(func() {
		ks.disposeMtx.Lock()
		ks.disposed = true
		ks.disposeMtx.Unlock()

		for _, cb := range ks.disposeCallbacks {
			go cb(ks)
		}
		ks.eventSubsMtx.Lock()
		for _, ch := range ks.eventSubs {
			close(ch)
		}
		ks.eventSubs = nil
		ks.eventSubsMtx.Unlock()
		ks.disposeCallbacks = nil
	})
}

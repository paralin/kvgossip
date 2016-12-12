package client

import (
	"sync"
)

type KeyInterest struct {
	Key           string
	Subscriptions map[int]*KeySubscription

	subscriptionCounter int
	onDisposed          func()
	stateMutex          sync.Mutex
	state               *KeySubscriptionState
	disposed            chan bool
}

func (ki *KeyInterest) AddSubscription() *KeySubscription {
	id := ki.subscriptionCounter
	ki.subscriptionCounter++
	sub := &KeySubscription{interest: ki}
	sub.Unsubscribe = func() {
		if sub.disposed {
			return
		}
		sub.disposed = true
		delete(ki.Subscriptions, id)
		ki.checkDispose()
	}
	ki.Subscriptions[id] = sub
	return sub
}

func (ki *KeyInterest) State() *KeySubscriptionState {
	ki.stateMutex.Lock()
	defer ki.stateMutex.Unlock()
	return ki.state
}

func (ki *KeyInterest) checkDispose() {
	if len(ki.Subscriptions) == 0 {
		ki.Dispose()
	}
}

func (ki *KeyInterest) updateLoop() {
	for {
		select {
		case <-ki.disposed:
			return
		}
	}
}

func (ki *KeyInterest) Dispose() {
	for _, sub := range ki.Subscriptions {
		sub.disposed = true
	}
	ki.Subscriptions = make(map[int]*KeySubscription)
	ki.onDisposed()
	select {
	case ki.disposed <- true:
	default:
	}
}

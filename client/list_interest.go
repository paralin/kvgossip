package client

import (
	"bytes"
	"sync"

	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/oleiade/lane"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type listInterest struct {
	// Wakeup the listInterest when a new conn comes in.
	newConn chan bool

	// Queue of connections to try
	connectionQueue *lane.Queue

	// When the active connection dies
	sessionEnded chan error

	// When we need to dispose the listInterest
	disposeChan chan bool

	// Subscription list
	subscriptions   []*KeyListSubscription
	subscriptionMtx sync.RWMutex

	state    map[string]*KeyListEntry
	stateMtx sync.RWMutex

	// Initial set
	initialSet map[string]bool
}

func newListInterest() *listInterest {
	return &listInterest{
		newConn:         make(chan bool, 1),
		connectionQueue: lane.NewQueue(),
		disposeChan:     make(chan bool, 2),
		state:           make(map[string]*KeyListEntry),
	}
}

func (in *listInterest) addSubscription(ks *KeyListSubscription) {
	in.subscriptionMtx.Lock()
	in.subscriptions = append(in.subscriptions, ks)
	go ks.updateLoop()
	in.subscriptionMtx.Unlock()
}

func (in *listInterest) removeSubscription(ks *KeyListSubscription) {
	in.subscriptionMtx.Lock()
	for i, sub := range in.subscriptions {
		if sub == ks {
			in.subscriptions[i] = in.subscriptions[len(in.subscriptions)-1]
			in.subscriptions[len(in.subscriptions)-1] = nil
			in.subscriptions = in.subscriptions[:len(in.subscriptions)-1]
			break
		}
	}
	in.subscriptionMtx.Unlock()
}

func (in *listInterest) nextEvent(event *KeyListSubscriptionEvent) {
	in.stateMtx.Lock()
	in.subscriptionMtx.RLock()
	defer in.subscriptionMtx.RUnlock()
	defer in.stateMtx.Unlock()

	for _, sub := range in.subscriptions {
		sub.next(event)
	}
}

// Kill the interest by filling the dispose channel.
func (in *listInterest) dispose() {
	for i := 0; i < 2; i++ {
		select {
		case in.disposeChan <- true:
		default:
			return
		}
	}
}

func (i *listInterest) addConnection(c *Connection) {
	if c.Released() {
		return
	}
	i.connectionQueue.Append(c)
	select {
	case i.newConn <- true:
	default:
	}
}

func (i *listInterest) session(conn *Connection) (sessionError error) {
	sessionCtx, cancelFunc := context.WithCancel(context.Background())

	defer func() {
		if sessionError != nil {
			conn.setError(sessionError)
		}
		if !conn.Released() {
			i.connectionQueue.Append(conn)
		}
		i.sessionEnded <- sessionError
		cancelFunc()
	}()

	connReleased := make(chan bool, 1)
	conn.OnRelease(func(c *Connection) {
		connReleased <- true
	})

	stb := conn.stub
	client, err := stb.ListKeys(sessionCtx, &ctl.ListKeysRequest{
		Watch:   true,
		MaxKeys: 0,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	rch := make(chan *ctl.ListKeysResponse)
	rchErr := make(chan error, 1)
	go func() (rchError error) {
		defer func() {
			rchErr <- rchError
		}()

		callCtx := client.Context()
		for {
			resp, err := client.Recv()
			if err != nil {
				return err
			}
			select {
			case rch <- resp:
			case <-callCtx.Done():
				return nil
			}
		}
	}()

	initialSetComplete := false
	i.prepareInitialSet()
	for {
		select {
		case <-i.disposeChan:
			return nil
		case <-connReleased:
			return nil
		case err := <-rchErr:
			return err
		case resp := <-rch:
			if !initialSetComplete {
				if resp.State == ctl.ListKeysResponse_LIST_KEYS_INITIAL_SET {
					delete(i.initialSet, resp.Key)
				} else {
					initialSetComplete = true
					i.finishInitialSet()
				}
			}
			i.handleNextResponse(resp)
		}
	}
}

func (i *listInterest) handleNextResponse(resp *ctl.ListKeysResponse) {
	i.stateMtx.Lock()
	defer i.stateMtx.Unlock()

	var action KeyListSubscriptionEvent_Action

	existing, ok := i.state[resp.Key]
	if len(resp.Hash) == 0 {
		if !ok {
			return
		}
		action = KeyList_Remove
		delete(i.state, resp.Key)
	} else {
		if ok {
			if bytes.Compare(existing.Hash, resp.Hash) == 0 {
				return
			} else {
				action = KeyList_Update
			}
		} else {
			action = KeyList_Add
		}
		i.state[resp.Key] = &KeyListEntry{
			Hash: resp.Hash,
			Key:  resp.Key,
		}
	}

	i.nextEvent(&KeyListSubscriptionEvent{
		Action: action,
		Hash:   resp.Hash,
		Key:    resp.Key,
	})
}

func (i *listInterest) prepareInitialSet() {
	i.stateMtx.Lock()
	i.initialSet = make(map[string]bool)
	for _, ent := range i.state {
		i.initialSet[ent.Key] = true
	}
	i.stateMtx.Unlock()
}

func (i *listInterest) finishInitialSet() {
	i.stateMtx.Lock()
	for key := range i.initialSet {
		i.nextEvent(&KeyListSubscriptionEvent{
			Key:    key,
			Action: KeyList_Remove,
		})
	}
	i.initialSet = nil
	i.stateMtx.Unlock()
}

func (i *listInterest) updateLoop() {
	sessionActive := false
	for {
		if !sessionActive && !i.connectionQueue.Empty() {
			// Try to get a connection
			nextConn := i.connectionQueue.Dequeue().(*Connection)
			if nextConn.Released() {
				continue
			}
			// Spawn the session goroutine
			sessionActive = true
			i.sessionEnded = make(chan error, 1)
			go i.session(nextConn)
		}

		// Sleep until we get a new connection to try
		// Or, if we have an active session, wait for that to die.
		select {
		case <-i.sessionEnded:
			sessionActive = false
			i.sessionEnded = nil
			continue
		case <-i.newConn:
			continue
		case <-i.disposeChan:
			return
		}
	}
}

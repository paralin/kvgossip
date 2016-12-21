package client

import (
	"sync"

	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/tx"
	"github.com/oleiade/lane"
	"golang.org/x/net/context"
)

type interest struct {
	key string

	// Wakeup the interest when a new conn comes in.
	newConn chan bool

	// Queue of connections to try
	connectionQueue *lane.Queue

	// When the active connection dies
	sessionEnded chan error

	// When we need to dispose the interest
	disposeChan chan bool

	// Subscription list
	subscriptions   []*KeySubscription
	subscriptionMtx sync.RWMutex

	state    *KeySubscriptionState
	stateMtx sync.RWMutex
}

func newInterest(key string) *interest {
	return &interest{
		key:             key,
		newConn:         make(chan bool, 1),
		connectionQueue: lane.NewQueue(),
		disposeChan:     make(chan bool, 2),
		state: &KeySubscriptionState{
			Dirty:       true,
			HasValue:    false,
			Transaction: nil,
		},
	}
}

func (in *interest) addSubscription(ks *KeySubscription) {
	in.subscriptionMtx.Lock()
	in.subscriptions = append(in.subscriptions, ks)
	go ks.updateLoop()
	in.subscriptionMtx.Unlock()

	in.stateMtx.RLock()
	ks.next(in.state)
	in.stateMtx.RUnlock()
}

func (in *interest) removeSubscription(ks *KeySubscription) {
	in.subscriptionMtx.Lock()
	for i, sub := range in.subscriptions {
		if sub == ks {
			close(ks.state)
			in.subscriptions[i] = in.subscriptions[len(in.subscriptions)-1]
			in.subscriptions[len(in.subscriptions)-1] = nil
			in.subscriptions = in.subscriptions[:len(in.subscriptions)-1]
			break
		}
	}
	in.subscriptionMtx.Unlock()
}

func (in *interest) nextState(state *KeySubscriptionState) {
	in.stateMtx.Lock()
	in.subscriptionMtx.RLock()
	in.state = state
	for _, sub := range in.subscriptions {
		sub.next(state)
	}
	in.subscriptionMtx.RUnlock()
	in.stateMtx.Unlock()
}

// Kill the interest by filling the dispose channel.
func (in *interest) dispose() {
	for i := 0; i < 2; i++ {
		select {
		case in.disposeChan <- true:
		default:
			return
		}
	}
}

func (i *interest) addConnection(c *Connection) {
	if c.Released() {
		return
	}
	i.connectionQueue.Append(c)
	select {
	case i.newConn <- true:
	default:
	}
}

func (i *interest) fetchValue(ctx context.Context, conn *Connection) error {
	resp, err := conn.stub.GetKey(ctx, &ctl.GetKeyRequest{
		Key: i.key,
	})
	if err != nil {
		return err
	}
	nextState := &KeySubscriptionState{
		HasValue:    true,
		Transaction: resp.Transaction,
	}
	i.nextState(nextState)
	return nil
}

func (i *interest) session(conn *Connection) (sessionError error) {
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
	client, err := stb.SubscribeKeyVer(sessionCtx, &ctl.SubscribeKeyVerRequest{
		Key: i.key,
	})
	if err != nil {
		return err
	}

	rch := make(chan *ctl.SubscribeKeyVerResponse)
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

	for {
		select {
		case <-i.disposeChan:
			return nil
		case <-connReleased:
			return nil
		case err := <-rchErr:
			return err
		case resp := <-rch:
			ns := i.handleNextVerification(resp.Verification)
			if ns.Dirty {
				doneChan := make(chan error, 1)
				go func() {
					doneChan <- i.fetchValue(sessionCtx, conn)
				}()
				select {
				case err := <-doneChan:
					if err != nil {
						return err
					}
				case <-connReleased:
					return nil
				case <-i.disposeChan:
					return nil
				}
			}
		}
	}
}

func (i *interest) handleNextVerification(v *tx.TransactionVerification) (nextState *KeySubscriptionState) {
	nextState = &KeySubscriptionState{
		HasValue: true,
	}

	if v == nil {
		i.nextState(nextState)
		return
	}

	i.stateMtx.RLock()
	lastState := i.state

	// Determine if we need to fetch a value.
	// This is in the following conditions:
	//  NEVER: if v == nil (covered above)
	//  Otherwise...
	//   - did not previously have value
	//  OR
	//   - previously had val & (prev tx == nil or prev tx does not match)
	nextState.Dirty = !lastState.HasValue ||
		(lastState.Transaction == nil ||
			lastState.Transaction.Verification.Timestamp != v.Timestamp)
	nextState.Transaction = lastState.Transaction

	i.stateMtx.RUnlock()
	i.nextState(nextState)
	return
}

func (i *interest) updateLoop() {
	sessionActive := false
	for {
		if !sessionActive {
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
		case <-i.newConn:
		case <-i.disposeChan:
			return
		}
	}
}

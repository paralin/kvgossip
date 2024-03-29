package client

import (
	"sync"

	"github.com/fuserobotics/kvgossip/ctl"
	"google.golang.org/grpc"
)

type Client struct {
	connIdCounter uint32
	connMtx       sync.RWMutex
	connections   map[uint32]*Connection

	interests   map[string]*interest
	interestMtx sync.Mutex

	listInterest    *listInterest
	listInterestMtx sync.Mutex
}

func NewClient() *Client {
	return &Client{
		connections: make(map[uint32]*Connection),
		interests:   make(map[string]*interest),
	}
}

func (c *Client) applyConnection(nconn *Connection) {
	c.connMtx.RLock()
	c.interestMtx.Lock()
	for _, interest := range c.interests {
		interest.addConnection(nconn)
	}
	c.interestMtx.Unlock()
	c.listInterestMtx.Lock()
	if c.listInterest != nil {
		c.listInterest.addConnection(nconn)
	}
	c.listInterestMtx.Unlock()
	c.connMtx.RUnlock()
}

func (c *Client) GetConnections() []ctl.ControlServiceClient {
	c.connMtx.RLock()
	defer c.connMtx.RUnlock()

	var result []ctl.ControlServiceClient
	for _, conn := range c.connections {
		result = append(result, conn.stub)
	}
	return result
}

func (c *Client) AddConnection(conn *grpc.ClientConn) *Connection {
	if conn == nil {
		return nil
	}

	// Build the connection and store it.
	c.connMtx.Lock()
	c.connIdCounter++
	nconn := newConnection(c.connIdCounter, conn)
	c.connections[nconn.id] = nconn
	nconn.OnRelease(func(oconn *Connection) {
		c.connMtx.Lock()
		delete(c.connections, oconn.id)
		c.connMtx.Unlock()
	})
	c.connMtx.Unlock()

	// Distribute the connection to interests
	go c.applyConnection(nconn)

	return nconn
}

func (c *Client) SubscribeKey(key string) *KeySubscription {
	c.interestMtx.Lock()
	defer c.interestMtx.Unlock()

	interest, ok := c.interests[key]
	if !ok {
		interest = newInterest(key)
		go interest.updateLoop()

		c.connMtx.RLock()
		for _, conn := range c.connections {
			interest.addConnection(conn)
		}
		c.connMtx.RUnlock()

		c.interests[key] = interest
	}

	interest.stateMtx.RLock()
	nsub := newKeySubscription(c, interest, interest.state)
	interest.stateMtx.RUnlock()

	interest.addSubscription(nsub)
	nsub.OnDisposed(func(*KeySubscription) {
		c.interestMtx.Lock()
		interest.removeSubscription(nsub)
		if len(interest.subscriptions) == 0 {
			delete(c.interests, interest.key)
			interest.dispose()
		}
		c.interestMtx.Unlock()
	})
	return nsub
}

func (c *Client) SubscribeKeyList() *KeyListSubscription {
	c.listInterestMtx.Lock()
	defer c.listInterestMtx.Unlock()

	listInterest := c.listInterest
	if listInterest == nil {
		listInterest = newListInterest()
		c.listInterest = listInterest
		go listInterest.updateLoop()

		c.connMtx.RLock()
		for _, conn := range c.connections {
			listInterest.addConnection(conn)
		}
		c.connMtx.RUnlock()
	}

	listInterest.stateMtx.RLock()
	nsub := newKeyListSubscription(c, listInterest)
	listInterest.stateMtx.RUnlock()

	listInterest.addSubscription(nsub)
	nsub.OnDisposed(func(*KeyListSubscription) {
		c.listInterestMtx.Lock()
		listInterest.removeSubscription(nsub)
		if len(listInterest.subscriptions) == 0 {
			listInterest.dispose()
			c.listInterest = nil
		}
		c.listInterestMtx.Unlock()
	})
	return nsub
}

package client

import (
	"sync"

	"google.golang.org/grpc"
)

type Client struct {
	connIdCounter uint32
	connMtx       sync.RWMutex
	connections   map[uint32]*Connection

	interests   map[string]*interest
	interestMtx sync.Mutex
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
	c.connMtx.RUnlock()
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
	}

	nsub := newKeySubscription(c, interest)
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

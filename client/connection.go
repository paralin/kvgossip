package client

import (
	"sync"

	"github.com/fuserobotics/kvgossip/ctl"
	"google.golang.org/grpc"
)

type Connection struct {
	id   uint32
	conn *grpc.ClientConn
	stub ctl.ControlServiceClient

	disposeOnce      sync.Once
	releaseCallbacks []func(c *Connection)
	releaseMtx       sync.RWMutex
	released         bool
	releaseError     error
}

func newConnection(id uint32, conn *grpc.ClientConn) *Connection {
	return &Connection{
		id:   id,
		conn: conn,
		stub: ctl.NewControlServiceClient(conn),
	}
}

func (c *Connection) Released() bool {
	c.releaseMtx.RLock()
	defer c.releaseMtx.RUnlock()
	return c.released
}

func (c *Connection) OnRelease(cb func(c *Connection)) {
	c.releaseMtx.Lock()
	defer c.releaseMtx.Unlock()

	if c.released {
		go cb(c)
	} else {
		c.releaseCallbacks = append(c.releaseCallbacks, cb)
	}
}

func (c *Connection) Error() error {
	c.releaseMtx.Lock()
	defer c.releaseMtx.Unlock()

	return c.releaseError
}

// Set the error, if it's a connection related error release the conn.
func (c *Connection) setError(err error) {
	if err == nil {
		return
	}

	c.releaseMtx.Lock()
	c.releaseError = err
	c.releaseMtx.Unlock()

	c.Release()
}

// Release the connection.
func (c *Connection) Release() {
	c.disposeOnce.Do(func() {
		c.releaseMtx.Lock()
		c.released = true
		// After this point no funcs will be added to callbacks, so unlock
		c.releaseMtx.Unlock()

		for _, cb := range c.releaseCallbacks {
			go cb(c)
		}
		c.releaseCallbacks = nil
	})
}

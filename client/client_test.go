package client

import (
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestAddConnection(t *testing.T) {
	client := NewClient()
	mockConn := &grpc.ClientConn{}
	nconn := client.AddConnection(mockConn)
	wasCalled := make(chan bool, 1)
	nconn.OnRelease(func(c *Connection) {
		select {
		case wasCalled <- true:
		default:
			t.Fatal("Was called called multiple times.")
		}
	})
	select {
	case <-wasCalled:
		t.Fatal("Release was called too early.")
	default:
	}
	nconn.Release()
	select {
	case <-wasCalled:
	case <-time.After(time.Duration(1) * time.Second):
		t.Fatal("Release callback was not called.")
	}
	if len(client.connections) > 0 {
		t.Fatal("Connection was not removed when released.")
	}
}

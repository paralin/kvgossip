package client

import (
	// log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/ctl"
	"golang.org/x/net/context"
)

// A remote key monitoring instance.
type KeyInterestRemote struct {
	Client ctl.ControlServiceClient
	Key    string

	disposeChan chan bool
	disposed    bool
}

func (kir *KeyInterestRemote) attemptStream() error {
	strm, err := kir.Client.SubscribeKeyVer(context.Background(), &ctl.SubscribeKeyVerRequest{
		Key: kir.Key,
	})
	if err != nil {
		return err
	}

	resp, err := strm.Recv()
	if err != nil {
		return err
	}

	_ = resp

	return nil
}

func (kir *KeyInterestRemote) updateLoop() {
	for !kir.disposed {
		kir.attemptStream()
	}
}

func (kir *KeyInterestRemote) Dispose() {
	kir.disposeChan <- true
	kir.disposed = true
}

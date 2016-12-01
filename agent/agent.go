package agent

import (
	"crypto/rsa"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/serf"
	"github.com/fuserobotics/kvgossip/sync"
)

type Agent struct {
	DB                *db.KVGossipDB
	SyncManager       *sync.SyncManager
	SerfManager       *serf.SerfManager
	ControlServer     *ctl.CtlServer
	ControlServerPort string
	SyncServicePort   int
	RootKey           *rsa.PublicKey
}

func NewAgent(dbPath string, syncServicePort int, controlServicePort string, rootKey *rsa.PublicKey, serfRpcAddr string) (*Agent, error) {
	res := &Agent{SyncServicePort: syncServicePort, ControlServerPort: controlServicePort}
	d, err := db.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}
	res.DB = d
	res.RootKey = rootKey
	res.SyncManager = sync.NewSyncManager(d, syncServicePort, rootKey)
	res.ControlServer = ctl.NewCtlServer(d, rootKey)
	if len(serfRpcAddr) == 0 {
		log.Info("Disabling serf (no address given).")
	} else {
		res.SerfManager = serf.NewSerfManager(res.SyncManager, serfRpcAddr)
	}
	return res, nil
}

func (a *Agent) Run() error {
	log.Info("DB ok, agent starting up...")
	a.SyncManager.Start()
	if err := a.ControlServer.Start(a.ControlServerPort); err != nil {
		return err
	}
	a.SerfManager.TreeHashChan = a.DB.TreeHashChanged
	a.SerfManager.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Info("Ctrl-c caught, shutting down...")
	a.SerfManager.Stop()
	a.SyncManager.Stop()
	a.ControlServer.Stop()

	return nil
}

func (a *Agent) SyncOnce(peer string) error {
	log.Infof("Attempting sync with %s...", peer)
	a.SyncManager.Start()
	defer func() {
		a.SyncManager.Stop()
	}()
	return a.SyncManager.Connect(peer)
}

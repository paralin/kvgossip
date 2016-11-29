package agent

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/sync"
)

type Agent struct {
	DB              *db.KVGossipDB
	SyncManager     *sync.SyncManager
	SyncServicePort int
}

func NewAgent(dbPath string, syncServicePort int) (*Agent, error) {
	res := &Agent{SyncServicePort: syncServicePort}
	d, err := db.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}
	res.DB = d
	res.SyncManager = sync.NewSyncManager(d, syncServicePort)
	return res, nil
}

func (a *Agent) Run() error {
	log.Info("DB ok, agent starting up...")
	a.SyncManager.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Info("Ctrl-c caught, shutting down...")
	a.SyncManager.Stop()

	return nil
}

func (a *Agent) SyncOnce(peer string) error {
	log.Infof("Attempting sync with %s...", peer)
	return nil
}

package agent

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/db"
)

type Agent struct {
	DB *db.KVGossipDB
}

func NewAgent(dbPath string) (*Agent, error) {
	res := &Agent{}
	d, err := db.OpenDB(dbPath)
	if err == nil {
		return nil, err
	}
	res.DB = d
	return res, nil
}

func (a *Agent) Run() error {
	log.Info("DB ok, agent starting up...")
	return nil
}

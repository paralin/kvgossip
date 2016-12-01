package main

import (
	"crypto/rsa"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/agent"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/urfave/cli"
)

var AgentFlags struct {
	DbPath             string
	SyncServicePort    int
	ControlServicePort string
	RootKeyPath        string
	RootKey            *rsa.PublicKey
	SerfRpcAddr        string
}

func loadRootKey() error {
	data, err := ioutil.ReadFile(AgentFlags.RootKeyPath)
	if err != nil {
		return err
	}
	pk, err := util.ParsePublicKey(data)
	if err != nil {
		return err
	}
	AgentFlags.RootKey = pk
	return nil
}

func buildAgent() (*agent.Agent, error) {
	log.Infof("Loading root key %s...", AgentFlags.RootKeyPath)
	if err := loadRootKey(); err != nil {
		return nil, err
	}
	log.Infof("Attempting to open DB %s...", AgentFlags.DbPath)
	ag, err := agent.NewAgent(AgentFlags.DbPath, AgentFlags.SyncServicePort, AgentFlags.ControlServicePort, AgentFlags.RootKey, AgentFlags.SerfRpcAddr)
	if err != nil {
		log.Errorf("Error opening db: %v", err)
	}
	return ag, err
}

func runAgent(c *cli.Context) error {
	ag, err := buildAgent()
	if err != nil {
		return err
	}
	return ag.Run()
}

var AgentCommand cli.Command = cli.Command{
	Name:   "agent",
	Usage:  "run the kvgossip agent",
	Action: runAgent,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:        "port, p",
			Usage:       "Start sync service on `PORT`.",
			Value:       9021,
			Destination: &AgentFlags.SyncServicePort,
		},
		cli.StringFlag{
			Name:        "ctlport, c",
			Usage:       "Start control service on `PORT`.",
			Value:       "localhost:9022",
			Destination: &AgentFlags.ControlServicePort,
		},
		cli.StringFlag{
			Name:        "serfrpc, r",
			Usage:       "Connect to Serf RPC at `ADDR`.",
			Value:       "127.0.0.1:7373",
			Destination: &AgentFlags.SerfRpcAddr,
		},
	},
}

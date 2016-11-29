package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/agent"
	"github.com/urfave/cli"
)

var AgentFlags struct {
	DbPath          string
	SyncServicePort int
}

func buildAgent() (*agent.Agent, error) {
	log.Infof("Attempting to open DB %s...", AgentFlags.DbPath)
	ag, err := agent.NewAgent(AgentFlags.DbPath, AgentFlags.SyncServicePort)
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
	},
}

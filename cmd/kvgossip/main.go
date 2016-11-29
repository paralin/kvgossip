package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/agent"
	"github.com/urfave/cli"
)

var AgentFlags struct {
	DbPath string
}

func main() {
	app := cli.NewApp()
	app.Name = "kvgossip"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:  "agent",
			Usage: "run the kvgossip agent",
			Action: func(c *cli.Context) error {
				log.Infof("Attempting to open DB %s...", AgentFlags.DbPath)
				ag, err := agent.NewAgent(AgentFlags.DbPath)
				if err != nil {
					log.Fatalf("Error opening db: %v", err)
				}
				return ag.Run()
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "dbpath, d",
					Usage:       "Use database at `PATH`.",
					Value:       "./kvgossip.db",
					Destination: &AgentFlags.DbPath,
				},
			},
		},
	}
	app.Run(os.Args)
}

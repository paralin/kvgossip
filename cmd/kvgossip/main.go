package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "kvgossip"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		AgentCommand,
		ManualSyncCommand,
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "dbpath, d",
			Usage:       "Use database at `PATH`.",
			Value:       "./kvgossip.db",
			Destination: &AgentFlags.DbPath,
		},
	}
	app.Run(os.Args)
}

package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
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

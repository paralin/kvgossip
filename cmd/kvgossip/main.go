package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	ver "github.com/fuserobotics/kvgossip/version"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "kvgossip"
	app.Version = ver.KVGossipVersion
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		AgentCommand,
		ManualSyncCommand,
		KeyGenCommand,
		ControlCommand,
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "dbpath, d",
			Usage:       "Use database at `PATH`.",
			Value:       "./kvgossip.db",
			Destination: &AgentFlags.DbPath,
		},
		cli.StringFlag{
			Name:        "rootkey, r",
			Usage:       "Use root key at `PATH`.",
			Value:       "./root_key.pem",
			Destination: &AgentFlags.RootKeyPath,
		},
	}
	app.Run(os.Args)
}

package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/cmd/kvgossip/cmds"
	ver "github.com/fuserobotics/kvgossip/version"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "kvgossip"
	app.Version = ver.KVGossipVersion
	app.EnableBashCompletion = true
	app.Commands = cmds.RootCommands
	app.Flags = cmds.RootFlags
	app.Run(os.Args)
}

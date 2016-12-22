package cmds

import (
	"github.com/urfave/cli"
)

var RootCommands = []cli.Command{
	AgentCommand,
	ManualSyncCommand,
	KeyGenCommand,
	ControlCommand,
}

var RootFlags = []cli.Flag{
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

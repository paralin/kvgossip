package main

import (
	"github.com/urfave/cli"
)

var ManualSyncFlags struct {
	Peer string
}

func runManualSync(c *cli.Context) error {
	ag, err := buildAgent()
	if err != nil {
		return err
	}
	return ag.SyncOnce(ManualSyncFlags.Peer)
}

var ManualSyncCommand cli.Command = cli.Command{
	Name:   "sync",
	Usage:  "Sync once with a peer and exit.",
	Action: runManualSync,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "peer",
			Usage:       "Sync with peer at `ADDRESS`.",
			Value:       "localhost:9021",
			Destination: &ManualSyncFlags.Peer,
		},
	},
}

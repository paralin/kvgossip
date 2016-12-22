package cmds

import (
	log "github.com/Sirupsen/logrus"
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
	err = ag.SyncOnce(ManualSyncFlags.Peer)
	if err != nil {
		log.Errorf("Error syncing: %v", err)
	}
	return err
}

var ManualSyncCommand cli.Command = cli.Command{
	Name:   "sync",
	Usage:  "Sync once with a peer and exit.",
	Action: runManualSync,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "peer",
			Usage:       "Sync with peer at `ADDRESS`.",
			Value:       "127.0.0.1:9021",
			Destination: &ManualSyncFlags.Peer,
		},
	},
}

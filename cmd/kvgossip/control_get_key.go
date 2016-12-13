package main

import (
	"errors"
	"os"

	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var GetKeyFlags struct {
	Key string
}

var ControlGetKeyCommand cli.Command = cli.Command{
	Name:  "getkey",
	Usage: "Get a key.",
	Action: func(c *cli.Context) error {
		if GetKeyFlags.Key == "" {
			return errors.New("Key must be specified.")
		}

		res, err := ControlFlags.Client.GetKey(context.Background(), &ctl.GetKeyRequest{Key: GetKeyFlags.Key})
		if err != nil {
			return err
		}

		if res.Transaction == nil {
			return errors.New("Server does not know about key.")
		}
		os.Stdout.Write(res.Transaction.Value)
		return nil
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "key",
			Usage:       "Key to get.",
			Destination: &GetKeyFlags.Key,
		},
	},
}

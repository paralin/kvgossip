package main

import (
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var ControlListGrantsCommand cli.Command = cli.Command{
	Name:  "listgrants",
	Usage: "List all grants.",
	Action: func(c *cli.Context) error {
		res, err := ControlFlags.Client.GetGrants(context.Background(), &ctl.GetGrantsRequest{})
		if err != nil {
			return err
		}

		return printGrantList(res.Grants)
	},
}

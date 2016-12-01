package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/util"
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

		log.Info("List of grants:")
		for _, gra := range res.Grants {
			vgd, err := grant.ValidateGrantData(gra)
			if err != nil {
				return err
			}
			hash := util.HexSha256(gra.Body)
			log.Infof("%s: %s", hash[:8], vgd.Grant.KeyRegex)
		}

		return nil
	},
}

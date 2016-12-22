package cmds

import (
	"bytes"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var RevokeGrantArgs struct {
	GrantHash string
}

func printGrantList(grants []*data.SignedData) error {
	log.Info("List of grants:")
	for _, gra := range grants {
		vgd, err := grant.ValidateGrantData(gra)
		if err != nil {
			return err
		}
		msg := ""
		ourKey, _ := util.MarshalPublicKey(&ControlFlags.Key.PublicKey)
		if bytes.Compare(vgd.Grant.IssuerKey, ourKey) == 0 {
			msg = "(issued by you)"
		} else if bytes.Compare(vgd.Grant.IssueeKey, ourKey) == 0 {
			msg = "(issued to you)"
		}
		hash := util.HexSha256(gra.Body)
		log.Infof("%s: %s %s", hash[:8], vgd.Grant.KeyRegex, msg)
	}
	return nil
}

var ControlRevokeGrantCommand cli.Command = cli.Command{
	Name:  "revokegrant",
	Usage: "Revoke a grant.",
	Action: func(c *cli.Context) error {
		if len(RevokeGrantArgs.GrantHash) == 0 {
			return errors.New("Grant hash must be specified.")
		}
		res, err := ControlFlags.Client.GetGrants(context.Background(), &ctl.GetGrantsRequest{})
		if err != nil {
			return err
		}

		printGrantList(res.Grants)

		var toRevoke *data.SignedData
		for _, gr := range res.Grants {
			hash := util.HexSha256(gr.Body)
			if bytes.Compare([]byte(hash)[:len(RevokeGrantArgs.GrantHash)], []byte(RevokeGrantArgs.GrantHash)) == 0 {
				toRevoke = gr
				break
			}
		}

		if toRevoke == nil {
			return errors.New("Didn't find in the list.")
		}

		rev, err := grant.BuildNewGrantRevocation(toRevoke)
		if err != nil {
			return err
		}

		nsd, err := grant.SignGrantRevocation(rev, ControlFlags.Key)
		if err != nil {
			return err
		}

		log.Debug("Signed grant revocation, putting it now.")
		_, err = ControlFlags.Client.PutRevocation(context.Background(), &ctl.PutRevocationRequest{Revocation: nsd})
		return err
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Destination: &RevokeGrantArgs.GrantHash,
			Name:        "grant",
			Usage:       "Grant hash to revoke.",
		},
	},
}

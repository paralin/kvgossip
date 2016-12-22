package cmds

import (
	"crypto/rsa"
	"errors"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/key"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"io/ioutil"
)

var NewGrantFlags struct {
	KeyFile       string
	Key           *rsa.PublicKey
	GrantKey      string
	AllowSubgrant bool
}

func loadGrantKey(c *cli.Context) error {
	data, err := ioutil.ReadFile(NewGrantFlags.KeyFile)
	if err != nil {
		return err
	}
	pk, err := util.ParsePublicKey(data)
	if err != nil {
		return err
	}
	NewGrantFlags.Key = pk
	return nil
}

var ControlBuildGrantCommand cli.Command = cli.Command{
	Name:  "newgrant",
	Usage: "Create a new grant.",
	Action: func(c *cli.Context) error {
		if NewGrantFlags.GrantKey == "" {
			return errors.New("Grant key pattern must be specified.")
		}
		if err := key.ValidatePattern(NewGrantFlags.GrantKey); err != nil {
			return err
		}

		// Build the new grant.
		ng, err := grant.BuildNewGrant(ControlFlags.Key, NewGrantFlags.Key, NewGrantFlags.GrantKey, NewGrantFlags.AllowSubgrant)
		if err != nil {
			return err
		}

		// Sign it
		sg, err := grant.SignGrant(ng, ControlFlags.Key)
		if err != nil {
			return err
		}

		// Push it.
		_, err = ControlFlags.Client.PutGrant(context.Background(), &ctl.PutGrantRequest{
			Pool: &grant.GrantAuthorizationPool{
				SignedGrants: []*data.SignedData{
					sg,
				},
			},
		})
		return err
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "target",
			Usage:       "Public key of the recipient entity.",
			Value:       "target_pub.pem",
			Destination: &NewGrantFlags.KeyFile,
		},
		cli.StringFlag{
			Name:        "grantpattern",
			Usage:       "Key pattern we are making a grant for.",
			Destination: &NewGrantFlags.GrantKey,
		},
		cli.BoolFlag{
			Destination: &NewGrantFlags.AllowSubgrant,
			Name:        "subgrant",
			Usage:       "Allow subgrants?",
		},
	},
	Before:      loadGrantKey,
	Subcommands: []cli.Command{},
}

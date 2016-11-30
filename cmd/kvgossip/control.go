package main

import (
	"crypto/rsa"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"io/ioutil"
)

var ControlFlags struct {
	KeyFile    string
	Key        *rsa.PrivateKey
	Service    string
	Connection *grpc.ClientConn
	Client     ctl.ControlServiceClient
}

func loadControlStuff(c *cli.Context) error {
	data, err := ioutil.ReadFile(ControlFlags.KeyFile)
	if err != nil {
		return err
	}
	pk, err := util.ParsePrivateKey(data)
	if err != nil {
		return err
	}
	ControlFlags.Key = pk

	// attempt to dial the service
	conn, err := grpc.Dial(ControlFlags.Service, grpc.WithInsecure())
	if err != nil {
		return err
	}

	ControlFlags.Connection = conn
	ControlFlags.Client = ctl.NewControlServiceClient(conn)

	return nil
}

var ControlCommand cli.Command = cli.Command{
	Name:  "control",
	Usage: "Control the local agent.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "private",
			Usage:       "Private key of the entity acting.",
			Value:       "entity.pem",
			Destination: &ControlFlags.KeyFile,
		},
		cli.StringFlag{
			Name:        "service",
			Usage:       "Connect to the service at `SERVICE`.",
			Value:       "localhost:9022",
			Destination: &ControlFlags.Service,
		},
	},
	Before: loadControlStuff,
	Subcommands: []cli.Command{
		ControlBuildGrantCommand,
		ControlSetKeyCommand,
	},
	After: func(c *cli.Context) error {
		if ControlFlags.Connection != nil {
			ControlFlags.Connection.Close()
			ControlFlags.Connection = nil
		}
		return nil
	},
}

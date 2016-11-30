package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/util"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"io/ioutil"
	"time"
)

var SetKeyFlags struct {
	Key       string
	ValueFile string
}

var ControlSetKeyCommand cli.Command = cli.Command{
	Name:  "setkey",
	Usage: "Set a key.",
	Action: func(c *cli.Context) error {
		if SetKeyFlags.Key == "" {
			return errors.New("Key must be specified.")
		}

		mpk, err := util.MarshalPublicKey(&ControlFlags.Key.PublicKey)
		if err != nil {
			return err
		}

		res, err := ControlFlags.Client.GetGrantPool(context.Background(), &ctl.GetGrantPoolRequest{
			EntityPublicKey: mpk,
			Key:             SetKeyFlags.Key,
		})
		if err != nil {
			return err
		}

		ngrants := len(res.Transaction.Verification.Grant.SignedGrants)
		log.Infof("Requested grant pool, got %d grants, %d relevant revocations, %d invalid.",
			ngrants,
			len(res.Revocations),
			len(res.Invalid))

		if ngrants == 0 {
			return errors.New("No valid grants for that key, you can't access it.")
		}

		data, err := ioutil.ReadFile(SetKeyFlags.ValueFile)
		if err != nil {
			return err
		}

		hashed := sha256.Sum256(data)
		sig, err := rsa.SignPKCS1v15(rand.Reader, ControlFlags.Key, crypto.SHA256, hashed[:])
		if err != nil {
			return err
		}

		trans := res.Transaction
		trans.Value = data
		trans.Verification.Timestamp = uint64(util.TimeToNumber(time.Now()))
		trans.Verification.ValueSignature = sig

		log.Infof("Built transaction successfully.")
		_, err = ControlFlags.Client.PutTransaction(context.Background(), &ctl.PutTransactionRequest{
			Transaction: trans,
		})
		return err
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "key",
			Usage:       "Key to set.",
			Destination: &SetKeyFlags.Key,
		},
		cli.StringFlag{
			Name:        "value",
			Usage:       "File containing a value.",
			Destination: &SetKeyFlags.ValueFile,
		},
	},
}

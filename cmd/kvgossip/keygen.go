package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/urfave/cli"
)

var KeyGenFlags struct {
	KeyFile    string
	KeyFilePub string
}

var KeyGenCommand cli.Command = cli.Command{
	Name:   "keygen",
	Usage:  "Create a private key.",
	Action: runKeyGen,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "public",
			Usage:       "Output public key file.",
			Value:       "root_key.pem",
			Destination: &KeyGenFlags.KeyFilePub,
		},
		cli.StringFlag{
			Name:        "priv",
			Usage:       "Output private key file.",
			Value:       "root.pem",
			Destination: &KeyGenFlags.KeyFile,
		},
	},
}

func runKeyGen(c *cli.Context) error {
	pkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	pkeydata := x509.MarshalPKCS1PrivateKey(pkey)
	blk := pem.Block{
		Bytes: pkeydata,
		Type:  "PRIVATE KEY",
	}
	pkeyout := pem.EncodeToMemory(&blk)
	if err := ioutil.WriteFile(KeyGenFlags.KeyFile, pkeyout, 0666); err != nil {
		return err
	}

	pubkeydata, err := x509.MarshalPKIXPublicKey(&pkey.PublicKey)
	if err != nil {
		return err
	}
	blk = pem.Block{
		Bytes: pubkeydata,
		Type:  "PUBLIC KEY",
	}
	pubkeyout := pem.EncodeToMemory(&blk)
	return ioutil.WriteFile(KeyGenFlags.KeyFilePub, pubkeyout, 0666)
}

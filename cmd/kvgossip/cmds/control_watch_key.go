package cmds

import (
	"errors"
	"fmt"

	"github.com/fuserobotics/kvgossip/client"
	"github.com/urfave/cli"
)

var WatchKeyFlags struct {
	Key string
}

var ControlWatchKeyCommand cli.Command = cli.Command{
	Name:  "watchkey",
	Usage: "Watch a key.",
	Action: func(c *cli.Context) error {
		if WatchKeyFlags.Key == "" {
			return errors.New("Key must be specified.")
		}

		conn := ControlFlags.Connection
		nt := client.NewClient()
		connHandle := nt.AddConnection(conn)
		defer connHandle.Release()
		releasedChan := make(chan bool, 1)
		connHandle.OnRelease(func(c *client.Connection) {
			select {
			case releasedChan <- true:
			default:
			}
		})

		subHandle := nt.SubscribeKey(WatchKeyFlags.Key)
		subHandle.OnDisposed(func(ks *client.KeySubscription) {
			select {
			case releasedChan <- true:
			default:
			}
		})
		ch := make(chan *client.KeySubscriptionState, 10)
		subHandle.Changes(ch)
		defer subHandle.Unsubscribe()

	UpdateLoop:
		for {
			select {
			case state, ok := <-ch:
				if !ok {
					break UpdateLoop
				}
				fmt.Printf("%#v\n", state)
			case <-releasedChan:
				break UpdateLoop
			}
		}
		return connHandle.Error()
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "key",
			Usage:       "Key to watch.",
			Destination: &WatchKeyFlags.Key,
		},
	},
}

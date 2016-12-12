package client

type Client struct {
	interests map[string]*KeyInterest
}

func NewClient() *Client {
	return &Client{
		interests: make(map[string]*KeyInterest),
	}
}

func (cli *Client) SubscribeKey(key string) *KeySubscription {
	interest, ok := cli.interests[key]
	if !ok {
		interest = &KeyInterest{
			Key:           key,
			Subscriptions: make(map[int]*KeySubscription),
			onDisposed: func() {
				delete(cli.interests, key)
			},
			disposed: make(chan bool, 1),
		}
		cli.interests[key] = interest
		go interest.updateLoop()
	}
	return interest.AddSubscription()
}

package client

type KeyListEntry struct {
	Key  string
	Hash []byte
}

func (kl *KeyListEntry) Copy() *KeyListEntry {
	return &KeyListEntry{
		Key:  kl.Key,
		Hash: kl.Hash,
	}
}

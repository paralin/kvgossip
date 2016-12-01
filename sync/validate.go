package sync

import (
	"errors"
	"fmt"
)

func (sgh *SyncGlobalHash) Validate() error {
	if sgh.KvgossipVersion == "" {
		return errors.New("Version string required.")
	}
	if len(sgh.GlobalTreeHash) != 32 {
		return errors.New("Sha256 global tree hash required.")
	}
	if len(sgh.HostNonce) != 10 {
		return fmt.Errorf("Host nonce length 30 required (got %d '%s').", len(sgh.HostNonce), sgh.HostNonce)
	}
	return nil
}

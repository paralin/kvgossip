package sync

import "errors"

func (sgh *SyncGlobalHash) Validate() error {
	if sgh.KvgossipVersion == "" {
		return errors.New("Version string required.")
	}
	if len(sgh.GlobalTreeHash) != 32 {
		return errors.New("Sha256 global tree hash required.")
	}
	if len(sgh.HostNonce) != 30 {
		return errors.New("Host nonce length 30 required.")
	}
	return nil
}

package ctl

import (
	"errors"
)

func (req *BuildTransactionRequest) Validate() error {
	if req.Key == "" {
		return errors.New("Must specify key.")
	}
	return nil
}

func (req *PutGrantRequest) Validate() error {
	if req.Pool == nil || len(req.Pool.SignedGrants) == 0 {
		return errors.New("Pool of new grants required.")
	}
	return nil
}

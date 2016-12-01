package ctl

import (
	"crypto/rsa"
	"errors"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/tx"
	"github.com/fuserobotics/kvgossip/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CtlServer struct {
	DB      *db.KVGossipDB
	RootKey *rsa.PublicKey

	server *grpc.Server
}

func NewCtlServer(d *db.KVGossipDB, rootKey *rsa.PublicKey) *CtlServer {
	return &CtlServer{
		DB:      d,
		RootKey: rootKey,
	}
}

func (ct *CtlServer) GetGrants(ctx context.Context, req *GetGrantsRequest) (*GetGrantsResponse, error) {
	// Pull the entire list of grants out of the DB.
	grants := ct.DB.GetAllGrants()
	return &GetGrantsResponse{
		Grants: grants,
	}, nil
}

// Request a pool of grants that would satisfy a request.
func (ct *CtlServer) BuildTransaction(ctx context.Context, req *BuildTransactionRequest) (*BuildTransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// Assert the public key is valid.
	_, err := util.ParsePublicKey(req.EntityPublicKey)
	if err != nil {
		return nil, err
	}
	// Pull the entire list of grants out of the DB.
	grants := ct.DB.GetAllGrants()
	// Compute chains
	trans := &tx.Transaction{
		Key:             req.Key,
		TransactionType: tx.Transaction_TRANSACTION_SET,
		Verification: &tx.TransactionVerification{
			Grant: &grant.GrantAuthorizationPool{
				SignedGrants: grants,
			},
			SignerPublicKey: req.EntityPublicKey,
			Timestamp:       uint64(util.TimeToNumber(time.Now())),
		},
	}
	res := tx.VerifyGrantAuthorization(trans, ct.RootKey, ct.DB)
	// build list for final pool
	finalPool := []*data.SignedData{}
	for _, chain := range res.Chains {
		for _, grd := range chain {
			finalPool = append(finalPool, grd.GrantData)
		}
	}
	finalPool = data.DedupeSignedData(finalPool)
	trans.Verification.Grant.SignedGrants = finalPool
	log.Debugf("Generated pool of %d grants for requested transaction against %s.", len(finalPool), req.Key)
	return &BuildTransactionResponse{
		Invalid:     res.InvalidGrants,
		Revocations: res.Revocations,
		Transaction: trans,
	}, nil
}

func (ct *CtlServer) PutGrant(ctx context.Context, req *PutGrantRequest) (*PutGrantResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	log.Debugf("Handling %d incoming grants.", len(req.Pool.SignedGrants))
	valid, revocations, _ := req.Pool.ValidGrants(false, ct.DB)
	for _, val := range valid {
		if err := ct.DB.PutGrant(val.GrantData); err != nil {
			return nil, err
		}
	}
	return &PutGrantResponse{
		Revocations: revocations,
	}, nil
}

func (ct *CtlServer) PutTransaction(ctx context.Context, req *PutTransactionRequest) (*PutTransactionResponse, error) {
	if req.Transaction == nil {
		return nil, errors.New("Transaction cannot be nil.")
	}
	if err := req.Transaction.Validate(); err != nil {
		return nil, err
	}
	res := tx.VerifyGrantAuthorization(req.Transaction, ct.RootKey, ct.DB)
	log.Debugf("Received new value on control channel for key %s timestamp %v.",
		req.Transaction.Key,
		util.NumberToTime(int64(req.Transaction.Verification.Timestamp)))
	if len(res.Chains) == 0 {
		return nil, errors.New("No valid grants for that transaction.")
	}
	if err := ct.DB.ApplyTransaction(req.Transaction); err != nil {
		return nil, err
	}
	return &PutTransactionResponse{}, nil
}

func (ct *CtlServer) GetKey(ctx context.Context, req *GetKeyRequest) (*GetKeyResponse, error) {
	if req.Key == "" {
		return nil, errors.New("Must specify key.")
	}
	res := &GetKeyResponse{}
	err := ct.DB.DB.View(func(tx *bolt.Tx) error {
		res.Transaction = ct.DB.GetTransaction(tx, req.Key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ct *CtlServer) Start(listen string) error {
	if ct.server != nil {
		return nil
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	RegisterControlServiceServer(grpcServer, ct)
	log.Infof("Control server listening on %s.", listen)
	ct.server = grpcServer
	go grpcServer.Serve(lis)
	return nil
}

func (ct *CtlServer) Stop() {
	if ct.server == nil {
		return
	}
	ct.server.Stop()
	ct.server = nil
}
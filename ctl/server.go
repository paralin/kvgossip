package ctl

import (
	"crypto/rsa"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
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

// Request a pool of grants that would satisfy a request.
func (ct *CtlServer) GetGrantPool(ctx context.Context, req *GetGrantPoolRequest) (*GetGrantPoolResponse, error) {
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
	return &GetGrantPoolResponse{
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

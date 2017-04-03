package ctlserver

import (
	"crypto/rsa"
	"errors"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	. "github.com/fuserobotics/kvgossip/ctl"
	"github.com/fuserobotics/kvgossip/data"
	"github.com/fuserobotics/kvgossip/db"
	"github.com/fuserobotics/kvgossip/filter"
	"github.com/fuserobotics/kvgossip/grant"
	"github.com/fuserobotics/kvgossip/tx"
	"github.com/fuserobotics/kvgossip/util"
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

func (ct *CtlServer) PutRevocation(ctx context.Context, req *PutRevocationRequest) (*PutRevocationResponse, error) {
	if req.Revocation == nil {
		return nil, errors.New("Cannot put a nil revocation.")
	}
	vgd, err := grant.ValidateGrantData(req.Revocation)
	if err != nil {
		return nil, err
	}
	if vgd.GrantRevocation == nil || vgd.RevokedGrant == nil {
		return nil, errors.New("Data did not contain a revocation.")
	}
	if err := ct.DB.ApplyRevocation(req.Revocation, ct.RootKey); err != nil {
		return nil, err
	}
	return &PutRevocationResponse{}, nil
}

func (ct *CtlServer) SubscribeKeyVer(req *SubscribeKeyVerRequest, stream ControlService_SubscribeKeyVerServer) error {
	if req.Key == "" {
		return errors.New("Key is required.")
	}

	ks := ct.DB.SubscribeKey(req.Key)
	defer ks.Unsubscribe()

	ch := make(chan *tx.Transaction, 5)
	ks.Changes(ch)

	ctx := stream.Context()

	for {
		select {
		case trans, ok := <-ch:
			if !ok {
				return nil
			}
			var verif *tx.TransactionVerification
			if trans != nil {
				verif = trans.Verification
			}
			err := stream.Send(&SubscribeKeyVerResponse{
				Verification: verif,
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (ct *CtlServer) ListKeys(req *ListKeysRequest, stream ControlService_ListKeysServer) error {
	done := stream.Context().Done()
	hasFilter := len(req.Filter) > 0
	noErrorErr := errors.New("no error")

	nk := uint32(0)
	err := ct.DB.DB.View(func(t *bolt.Tx) error {
		return ct.DB.ForeachKeyHash(t, func(k string, v []byte) error {
			if req.MaxKeys > 0 && nk >= req.MaxKeys {
				return noErrorErr
			}

			select {
			case <-done:
				return noErrorErr
			default:
			}

			if hasFilter && !filter.MatchesFilters(k, req.Filter) {
				return nil
			}

			nk++
			return stream.Send(&ListKeysResponse{
				Hash:  v,
				Key:   k,
				State: ListKeysResponse_LIST_KEYS_INITIAL_SET,
			})
		})
	})

	if err == noErrorErr {
		err = nil
	}
	if err == nil && req.Watch {
		err = stream.Send(&ListKeysResponse{
			State: ListKeysResponse_LIST_KEYS_TAIL,
		})
	}
	if err != nil || !req.Watch {
		return err
	}

	lsub := ct.DB.SubscribeKeyPattern(req.Filter)
	defer lsub.Unsubscribe()
	ch := make(chan *tx.Transaction, 10)
	lsub.Changes(ch)

	// Note: tailed keys will NOT have hash as an optimization.
	for {
		select {
		case <-done:
			return nil
		case next := <-ch:
			if err := stream.Send(&ListKeysResponse{
				Key:   next.Key,
				State: ListKeysResponse_LIST_KEYS_TAIL,
			}); err != nil {
				return err
			}
		}
	}
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

package tss

import (
	"context"
	"fmt"
	"time"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/vald/tss/rpc"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/axelarnetwork/utils/log"
)

// Mgr represents an object that manages all communication with the external tss process
type Mgr struct {
	multiSigClient rpc.MultiSigClient
	cliCtx         sdkClient.Context
	principalAddr  string
	keys           map[string][][]byte
	Timeout        time.Duration
	broadcaster    broadcast.Broadcaster
	cdc            *codec.LegacyAmino
}

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration) (*grpc.ClientConn, error) {
	tofndServerAddress := host + ":" + port
	log.Infof("initiate connection to tofnd gRPC server: %s", tofndServerAddress)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, tofndServerAddress, grpc.WithInsecure(), grpc.WithBlock())
}

// NewMgr returns a new tss manager instance
func NewMgr(multiSigClient rpc.MultiSigClient, cliCtx sdkClient.Context, timeout time.Duration, principalAddr string, broadcaster broadcast.Broadcaster, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		multiSigClient: multiSigClient,
		cliCtx:         cliCtx,
		Timeout:        timeout,
		principalAddr:  principalAddr,
		keys:           make(map[string][][]byte),
		broadcaster:    broadcaster,
		cdc:            cdc,
	}
}

// ProcessHeartBeatEvent broadcasts the heartbeat
func (mgr *Mgr) ProcessHeartBeatEvent(_ tmEvents.Event) error {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	defer cancel()

	// tofnd health check
	// this just checks if tofnd is up and running, so a dummy ID is fine
	request := &tofnd.KeyPresenceRequest{
		KeyUid: "dummyID",
		PubKey: []byte{},
	}

	response, err := mgr.multiSigClient.KeyPresence(grpcCtx, request)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to invoke KeyPresence grpc")
	}

	switch response.Response {
	case tofnd.RESPONSE_UNSPECIFIED, tofnd.RESPONSE_FAIL:
		return sdkerrors.Wrap(err, "tofnd not set up correctly")
	case tofnd.RESPONSE_PRESENT, tofnd.RESPONSE_ABSENT:
		// tofnd is working properly
		break
	default:
		return sdkerrors.Wrap(err, "unknown tofnd response")
	}

	tssMsg := tss.NewHeartBeatRequest(mgr.cliCtx.FromAddress)

	logger := log.With("listener", "tss")
	logger.Info(fmt.Sprintf("operator %s sending heartbeat", mgr.principalAddr))
	if _, err := mgr.broadcaster.Broadcast(context.TODO(), tssMsg); err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing heartbeat msg")
	}

	logger.Info(fmt.Sprintf("no keygen/signing issues reported for operator %s", mgr.principalAddr))

	return nil
}

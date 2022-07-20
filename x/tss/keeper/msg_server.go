package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.TSSKeeper
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	multisig    types.MultiSigKeeper
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.TSSKeeper, s types.Snapshotter, staker types.StakingKeeper, multisig types.MultiSigKeeper) types.MsgServiceServer {
	return msgServer{
		TSSKeeper:   keeper,
		snapshotter: s,
		staker:      staker,
		multisig:    multisig,
	}
}

func (s msgServer) HeartBeat(c context.Context, req *types.HeartBeatRequest) (*types.HeartBeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	participant := s.snapshotter.GetOperator(ctx, req.Sender)
	if participant.Empty() {
		return nil, fmt.Errorf("sender %s is not a registered proxy", req.Sender.String())
	}

	// this could happen after register proxy but before create validator
	if s.staker.Validator(ctx, participant) == nil {
		return nil, fmt.Errorf("%s is not a validator", participant)
	}

	for _, keyID := range req.KeyIDs {
		_, ok := s.multisig.GetKey(ctx, multisig.KeyID(keyID))
		if !ok {
			return nil, fmt.Errorf("operator %s sent heartbeat for unknown key ID %s", participant.String(), keyID)
		}
	}

	return &types.HeartBeatResponse{KeygenIllegibility: snapshot.None, SigningIllegibility: snapshot.None}, nil
}

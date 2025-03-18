package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	multisig    types.MultiSigKeeper
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, s types.Snapshotter, staker types.StakingKeeper, multisig types.MultiSigKeeper) types.MsgServiceServer {
	return msgServer{
		Keeper:      keeper,
		snapshotter: s,
		staker:      staker,
		multisig:    multisig,
	}
}

func (s msgServer) HeartBeat(c context.Context, req *types.HeartBeatRequest) (*types.HeartBeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid sender: %s", err)
	}

	participant := s.snapshotter.GetOperator(ctx, sender)
	if participant.Empty() {
		return nil, fmt.Errorf("sender %s is not a registered proxy", req.Sender)
	}

	// this could happen after register proxy but before create validator
	if _, err := s.staker.Validator(ctx, participant); err != nil {
		return nil, fmt.Errorf("%s is not a validator", participant)
	}

	if ctx.BlockHeight()-s.GetLastHeartbeatAt(ctx, participant) < s.GetParams(ctx).HeartbeatPeriodInBlocks/2 {
		return nil, fmt.Errorf("too many heartbeats received from operator %s", participant.String())
	}

	if err := s.SetLastHeartbeatAt(ctx, participant); err != nil {
		return nil, err
	}

	for _, keyID := range req.KeyIDs {
		_, ok := s.multisig.GetKey(ctx, multisig.KeyID(keyID))
		if !ok {
			return nil, fmt.Errorf("operator %s sent heartbeat for unknown key ID %s", participant.String(), keyID)
		}
	}

	return &types.HeartBeatResponse{}, nil
}

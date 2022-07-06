package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	snapshotter types.Snapshotter
	staker      types.Staker
	slasher     types.Slasher
}

// NewMsgServerImpl returns an implementation of the MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, snapshotter types.Snapshotter, staker types.Staker, slasher types.Slasher) types.MsgServiceServer {
	return msgServer{
		Keeper:      keeper,
		snapshotter: snapshotter,
		staker:      staker,
		slasher:     slasher,
	}
}

func (s msgServer) StartKeygen(c context.Context, req *types.StartKeygenRequest) (*types.StartKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	candidates := slices.Map(s.staker.GetBondedValidatorsByPower(ctx), stakingtypes.Validator.GetOperator)
	filter := func(v snapshot.ValidatorI) bool {
		if v.IsJailed() {
			return false
		}

		consAdd, err := v.GetConsAddr()
		if err != nil {
			return false
		}

		if s.slasher.IsTombstoned(ctx, consAdd) {
			return false
		}

		_, isActive := s.snapshotter.GetProxy(ctx, v.GetOperator())
		return isActive
	}
	snapshot, err := s.snapshotter.CreateSnapshot(ctx, candidates, filter, snapshot.QuadraticWeightFunc, s.GetParams(ctx).KeygenThreshold)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create snapshot for keygen")
	}

	err = s.CreateKeygenSession(ctx, req.KeyID, snapshot)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to start keygen")
	}

	return &types.StartKeygenResponse{}, nil
}

func (s msgServer) SubmitPubKey(c context.Context, req *types.SubmitPubKeyRequest) (*types.SubmitPubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keygenSession, ok := s.GetKeygenSession(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("keygen session %s not found", req.KeyID)
	}

	participant := s.snapshotter.GetOperator(ctx, req.Sender)
	if participant.Empty() {
		return nil, fmt.Errorf("sender %s is not a registered proxy", req.Sender.String())
	}

	if err := keygenSession.AddKey(ctx.BlockHeight(), participant, req.PubKey); err != nil {
		return nil, sdkerrors.Wrap(err, "unable to add public key for keygen")
	}
	s.setKeygenSession(ctx, keygenSession)

	ctx.EventManager().EmitTypedEvent(types.NewPubKeySubmitted(req.KeyID, participant, req.PubKey))

	if keygenSession.State != exported.Completed {
		return &types.SubmitPubKeyResponse{}, nil
	}

	key, err := keygenSession.Result()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to get keygen result")
	}

	s.SetKey(ctx, key)

	return &types.SubmitPubKeyResponse{}, nil
}

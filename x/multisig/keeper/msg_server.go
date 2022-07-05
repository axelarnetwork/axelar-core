package keeper

import (
	"context"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/slices"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (m msgServer) StartKeygen(c context.Context, request *types.StartKeygenRequest) (*types.StartKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	candidates := slices.Map(m.staker.GetBondedValidatorsByPower(ctx), stakingtypes.Validator.GetOperator)
	filter := func(v exported.ValidatorI) bool {
		if v.IsJailed() {
			return false
		}

		consAdd, err := v.GetConsAddr()
		if err != nil {
			return false
		}

		if m.slasher.IsTombstoned(ctx, consAdd) {
			return false
		}

		_, isActive := m.snapshotter.GetProxy(ctx, v.GetOperator())
		if !isActive {
			return false
		}

		return true
	}
	snapshot, err := m.snapshotter.CreateSnapshot(ctx, candidates, filter, exported.QuadraticWeightFunc, m.GetParams(ctx).KeygenThreshold)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create snapshot for keygen")
	}

	err = m.CreateKeygenSession(ctx, request.KeyID, snapshot)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to start keygen")
	}

	return &types.StartKeygenResponse{}, nil
}

func (m msgServer) SubmitPubKey(ctx context.Context, request *types.SubmitPubKeyRequest) (*types.SubmitPubKeyResponse, error) {
	// TODO implement me
	panic("implement me")
}

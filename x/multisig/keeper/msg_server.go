package keeper

import (
	"context"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	snapshotter Snapshotter
	staker      types.Staker
}

// NewMsgServer returns an implementation of the MsgServiceServer interface
// for the provided Keeper.
func NewMsgServer(keeper Keeper, snapshotter Snapshotter, staker types.Staker) types.MsgServiceServer {
	return msgServer{
		Keeper:      keeper,
		snapshotter: snapshotter,
		staker:      staker,
	}
}

func (s msgServer) StartKeygen(c context.Context, req *types.StartKeygenRequest) (*types.StartKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	snap, err := s.snapshotter.CreateSnapshot(ctx, s.getParams(ctx).KeygenThreshold)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to create snapshot for keygen")
	}

	err = s.createKeygenSession(ctx, req.KeyID, snap)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unable to start keygen")
	}

	return &types.StartKeygenResponse{}, nil
}

func (s msgServer) SubmitPubKey(c context.Context, req *types.SubmitPubKeyRequest) (*types.SubmitPubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keygenSession, ok := s.getKeygenSession(ctx, req.KeyID)
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

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewPubKeySubmitted(req.KeyID, participant, req.PubKey)))

	return &types.SubmitPubKeyResponse{}, nil
}

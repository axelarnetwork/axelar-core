package tss

import (
	"fmt"
	"strconv"

	"github.com/axelarnetwork/tssd/convert"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	exported2 "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

func NewHandler(k keeper.Keeper, s types.Staker, v types.Voter) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgKeygenTraffic:
			return handleMsgKeygenTraffic(ctx, k, msg)
		case types.MsgSignTraffic:
			return handleMsgSignTraffic(ctx, k, msg)
		case types.MsgKeygenStart:
			return handleMsgKeygenStart(ctx, k, s, v, msg)
		case types.MsgSignStart:
			return handleMsgSignStart(ctx, k, s, v, msg)
		case types.MsgAssignNextMasterKey:
			return handleMsgAssignNextMasterKey(ctx, k, s, v, msg)
		case types.MsgRotateMasterKey:
			return handleMsgRotateMasterKey(ctx, k, msg)
		case *types.MsgVotePubKey:
			return handleMsgVotePubKey(ctx, k, v, *msg)
		case *types.MsgVoteSig:
			return handleMsgVoteSig(ctx, k, v, *msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgRotateMasterKey(ctx sdk.Context, k keeper.Keeper, msg types.MsgRotateMasterKey) (*sdk.Result, error) {
	if err := k.RotateMasterKey(ctx, msg.Chain); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to rotate master key")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeChain, msg.Chain),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVoteSig(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgVoteSig) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, &msg); err != nil {
		return nil, err
	}

	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeSigPayload, string(msg.SigBytes)),
	)

	if vote := v.Result(ctx, msg.PollMeta); vote != nil {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))

		r, s, err := convert.BytesToSig(msg.SigBytes)
		if err != nil {
			return nil, err
		}
		if err := k.SetSig(ctx, msg.PollMeta.ID, exported2.Signature{R: r, S: s}); err != nil {
			return nil, err
		}
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVotePubKey(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgVotePubKey) (*sdk.Result, error) {
	if err := v.TallyVote(ctx, &msg); err != nil {
		return nil, err
	}

	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeKeyPayload, string(msg.PubKeyBytes)),
	)

	if vote := v.Result(ctx, msg.PollMeta); vote != nil {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))

		switch msg.PollMeta.Type {
		case types.MsgKeygenStart{}.Type():
			k.Logger(ctx).Debug(fmt.Sprintf("public key with ID %s confirmed", msg.PollMeta.ID))
			// Assert: types.MsgVotePubKey.ValidateBasic already checks the conversion, so this cannot fail here
			pubKey, _ := convert.BytesToPubkey(msg.PubKeyBytes)
			k.SetKey(ctx, msg.PollMeta.ID, pubKey)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting message type: %T", msg))
		}
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgAssignNextMasterKey(ctx sdk.Context, k keeper.Keeper, s types.Staker, v types.Voter, msg types.MsgAssignNextMasterKey) (*sdk.Result, error) {
	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, fmt.Errorf("key refresh failed")
	}

	err := k.AssignNextMasterKey(ctx, msg.Chain, snapshot.Height, msg.KeyID)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgKeygenTraffic(ctx sdk.Context, k keeper.Keeper, msg types.MsgKeygenTraffic) (*sdk.Result, error) {
	if err := k.KeygenMsg(ctx, msg); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, msg.Payload.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgKeygenStart(ctx sdk.Context, k keeper.Keeper, s types.Staker, v types.Voter, msg types.MsgKeygenStart) (*sdk.Result, error) {
	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, fmt.Errorf("key refresh failed")
	}

	if msg.Threshold < 1 || msg.Threshold > len(snapshot.Validators) {
		err := fmt.Errorf("invalid threshold: %d, validators: %d", msg.Threshold, len(snapshot.Validators))
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.NewKeyID}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

	pkChan, err := k.StartKeygen(ctx, msg.NewKeyID, msg.Threshold, snapshot.Validators)
	if err != nil {
		return nil, err
	}

	go func() {
		pk, ok := <-pkChan
		if ok {
			bz, err := convert.PubkeyToBytes(pk)
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				return
			}
			if err := v.Vote(ctx, &types.MsgVotePubKey{PollMeta: poll, PubKeyBytes: bz}); err != nil {
				k.Logger(ctx).Error(err.Error())
				return
			}
		}
	}()
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignStart(ctx sdk.Context, k keeper.Keeper, s types.Staker, v types.Voter, msg types.MsgSignStart) (*sdk.Result, error) {
	// TODO for now assume all validators participate
	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, fmt.Errorf("signing failed")
	}
	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.NewSigID}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

	sigChan, err := k.StartSign(ctx, msg, snapshot.Validators)
	if err != nil {
		return nil, err
	}

	go func() {
		sig, ok := <-sigChan
		if ok {
			bz, err := convert.SigToBytes(sig.R.Bytes(), sig.S.Bytes())
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				return
			}
			if err := v.Vote(ctx, &types.MsgVoteSig{PollMeta: poll, SigBytes: bz}); err != nil {
				k.Logger(ctx).Error(err.Error())
				return
			}
		}
	}()
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignTraffic(ctx sdk.Context, k keeper.Keeper, msg types.MsgSignTraffic) (*sdk.Result, error) {
	if err := k.SignMsg(ctx, msg); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, msg.Payload.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

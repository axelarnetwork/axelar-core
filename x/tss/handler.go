package tss

import (
	"fmt"
	"strconv"

	"github.com/axelarnetwork/tssd/convert"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler returns the handler for the tss module
func NewHandler(k keeper.Keeper, s types.Snapshotter, b types.Balancer, v types.Voter) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgKeygenTraffic:
			return handleMsgKeygenTraffic(ctx, k, msg)
		case types.MsgSignTraffic:
			return handleMsgSignTraffic(ctx, k, msg)
		case types.MsgKeygenStart:
			return handleMsgKeygenStart(ctx, k, s, v, msg)
		case types.MsgAssignNextMasterKey:
			return handleMsgAssignNextMasterKey(ctx, k, s, b, msg)
		case types.MsgRotateMasterKey:
			return handleMsgRotateMasterKey(ctx, k, b, msg)
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

func handleMsgRotateMasterKey(ctx sdk.Context, k keeper.Keeper, b types.Balancer, msg types.MsgRotateMasterKey) (*sdk.Result, error) {
	chain, ok := b.GetChain(ctx, msg.Chain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrTss, "unknown chain")
	}

	if err := k.RotateMasterKey(ctx, chain); err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeChain, chain.Name),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVoteSig(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgVoteSig) (*sdk.Result, error) {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeSigPayload, string(msg.SigBytes)),
	)

	if _, ok := k.GetSig(ctx, msg.PollMeta.ID); ok {
		// the signature is already set, no need for further processing of the vote
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	}

	if err := v.TallyVote(ctx, &msg); err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	if result := v.Result(ctx, msg.PollMeta); result != nil {
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))

		switch sigBytes := result.(type) {
		case []byte:
			r, s, err := convert.BytesToSig(sigBytes)
			if err != nil {
				return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
			}
			if err := k.SetSig(ctx, msg.PollMeta.ID, exported.Signature{R: r, S: s}); err != nil {
				return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
			}
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVotePubKey(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgVotePubKey) (*sdk.Result, error) {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeKeyPayload, string(msg.PubKeyBytes)),
	)

	if _, ok := k.GetKey(ctx, msg.PollMeta.ID); ok {
		// the key is already set, no need for further processing of the vote
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	}

	if err := v.TallyVote(ctx, &msg); err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	if result := v.Result(ctx, msg.PollMeta); result != nil {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollDecided, strconv.FormatBool(true)))
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		switch pkBytes := result.(type) {
		case []byte:
			k.Logger(ctx).Debug(fmt.Sprintf("public key with ID %s confirmed", msg.PollMeta.ID))
			pubKey, err := convert.BytesToPubkey(pkBytes)
			if err != nil {
				return nil, sdkerrors.Wrap(types.ErrTss, "could not marshal signature")
			}
			k.SetKey(ctx, msg.PollMeta.ID, pubKey)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgAssignNextMasterKey(ctx sdk.Context, k keeper.Keeper, s types.Snapshotter, b types.Balancer, msg types.MsgAssignNextMasterKey) (*sdk.Result, error) {
	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrTss, "key refresh failed")
	}
	chain, ok := b.GetChain(ctx, msg.Chain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrTss, "unknown chain")
	}
	err := k.AssignNextMasterKey(ctx, chain, snapshot.Height, msg.KeyID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
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
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
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

func handleMsgKeygenStart(ctx sdk.Context, k keeper.Keeper, s types.Snapshotter, v types.Voter, msg types.MsgKeygenStart) (*sdk.Result, error) {
	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrTss, "the system needs to have at least one validator snapshot")
	}

	if msg.Threshold < 1 || msg.Threshold > len(snapshot.Validators) {
		err := fmt.Errorf("invalid threshold: %d, validators: %d", msg.Threshold, len(snapshot.Validators))
		k.Logger(ctx).Error(err.Error())
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	poll := voting.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.NewKeyID}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	pkChan, err := k.StartKeygen(ctx, msg.NewKeyID, msg.Threshold, snapshot)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
	}

	go func() {
		pk, ok := <-pkChan
		if ok {
			bz, err := convert.PubkeyToBytes(pk)
			if err != nil {
				k.Logger(ctx).Error(err.Error())
				return
			}
			v.RecordVote(&types.MsgVotePubKey{PollMeta: poll, PubKeyBytes: bz})
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

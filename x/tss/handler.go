package tss

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// NewHandler returns the handler for the tss module
func NewHandler(k keeper.Keeper, s types.Snapshotter, n types.Nexus, v types.Voter, staker types.StakingKeeper, broadcaster types.Broadcaster) sdk.Handler {
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgKeygenTraffic:
			return handleMsgKeygenTraffic(ctx, k, broadcaster, msg)
		case *types.MsgSignTraffic:
			return handleMsgSignTraffic(ctx, k, broadcaster, msg)
		case *types.MsgKeygenStart:
			return handleMsgKeygenStart(ctx, k, s, staker, v, msg)
		case *types.MsgAssignNextKey:
			return handleMsgAssignNextKey(ctx, k, s, n, msg)
		case *types.MsgRotateKey:
			return handleMsgRotateKey(ctx, k, n, msg)
		case *types.MsgVotePubKey:
			return handleMsgVotePubKey(ctx, k, v, msg)
		case *types.MsgVoteSig:
			return handleMsgVoteSig(ctx, k, v, msg)
		case *types.MsgDeregister:
			return handleMsgDeregister(ctx, k, staker, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
		}
		if res.Log != "" {
			k.Logger(ctx).Debug(res.Log)
		}
		return res, nil
	}
}

func handleMsgRotateKey(ctx sdk.Context, k keeper.Keeper, n types.Nexus, msg *types.MsgRotateKey) (*sdk.Result, error) {
	chain, ok := n.GetChain(ctx, msg.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain")
	}

	if err := k.RotateKey(ctx, chain, msg.KeyRole); err != nil {
		return nil, err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("rotated %s key for chain %s", msg.KeyRole.SimpleString(), chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeChain, chain.Name),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgVoteSig(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg *types.MsgVoteSig) (*sdk.Result, error) {
	if _, ok := k.GetSig(ctx, msg.PollMeta.ID); ok {
		// the signature is already set, no need for further processing of the vote
		return &sdk.Result{Log: fmt.Sprintf("signature %s already verified", msg.PollMeta.ID)}, nil
	}

	if _, err := btcec.ParseDERSignature(msg.SigBytes, btcec.S256()); err != nil {
		return nil, sdkerrors.Wrap(err, "discard vote for invalid signature")
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.PollMeta, msg.SigBytes); err != nil {
		return nil, err
	}

	if result := v.Result(ctx, msg.PollMeta); result != nil {
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
			sdk.NewAttribute(types.AttributeKeyPoll, msg.PollMeta.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(msg.SigBytes)),
		))

		switch sigBytes := result.(type) {
		case []byte:
			k.SetSig(ctx, msg.PollMeta.ID, sigBytes)
			k.Logger(ctx).Info(fmt.Sprintf("signature verification result is %s", result))
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgVotePubKey(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg *types.MsgVotePubKey) (*sdk.Result, error) {
	if _, ok := k.GetKey(ctx, msg.PollMeta.ID); ok {
		// the key is already set, no need for further processing of the vote
		return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.PollMeta, msg.PubKeyBytes); err != nil {
		return nil, err
	}

	if result := v.Result(ctx, msg.PollMeta); result != nil {
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
			sdk.NewAttribute(types.AttributeKeyPoll, msg.PollMeta.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(msg.PubKeyBytes)),
		))
		switch pkBytes := result.(type) {
		case []byte:
			k.Logger(ctx).Debug(fmt.Sprintf("public key with ID %s confirmed", msg.PollMeta.ID))
			btcecPK, err := btcec.ParsePubKey(pkBytes, btcec.S256())
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal public key bytes: [%v]", err)
			}
			pubKey := btcecPK.ToECDSA()
			k.SetKey(ctx, msg.PollMeta.ID, *pubKey)
			k.Logger(ctx).Info(fmt.Sprintf("public key confirmation result is %.10s", result))
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgAssignNextKey(ctx sdk.Context, k keeper.Keeper, s types.Snapshotter, n types.Nexus, msg *types.MsgAssignNextKey) (*sdk.Result, error) {
	chain, ok := n.GetChain(ctx, msg.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain")
	}

	counter, ok := k.GetSnapshotCounterForKeyID(ctx, msg.KeyID)
	if !ok {
		return nil, fmt.Errorf("could not find snapshot counter for given key ID")
	}

	snapshot, ok := s.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not find snapshot for given key ID")
	}

	keyRequirement, found := k.GetKeyRequirement(ctx, chain, msg.KeyRole)
	if !found {
		return nil, fmt.Errorf("%s key is not required for chain %s", msg.KeyRole.SimpleString(), chain.Name)
	}

	if len(snapshot.Validators) < int(keyRequirement.MinValidatorSubsetSize) {
		return nil, fmt.Errorf(
			"expected %s's %s key to be generated with at least %d validators, actual %d",
			chain.Name,
			msg.KeyRole.SimpleString(),
			keyRequirement.MinValidatorSubsetSize,
			len(snapshot.Validators),
		)
	}

	if snapshot.KeyShareDistributionPolicy != keyRequirement.KeyShareDistributionPolicy {
		return nil, fmt.Errorf(
			"expected %s's %s key to have tss shares distributed with policy %s, actual %s",
			chain.Name,
			msg.KeyRole.SimpleString(),
			keyRequirement.KeyShareDistributionPolicy.SimpleString(),
			snapshot.KeyShareDistributionPolicy.SimpleString(),
		)
	}

	err := k.AssignNextKey(ctx, chain, msg.KeyRole, msg.KeyID)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("prepared %s key rotation for chain %s", msg.KeyRole.SimpleString(), chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgKeygenTraffic(ctx sdk.Context, k keeper.Keeper, broadcaster types.Broadcaster, msg *types.MsgKeygenTraffic) (*sdk.Result, error) {
	senderAddress := broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		return nil, fmt.Errorf("invalid message: sender [%s] is not a validator", msg.Sender)
	}

	if !k.DoesValidatorParticipateInKeygen(ctx, msg.SessionID, senderAddress) {
		return nil, fmt.Errorf("invalid message: sender [%.20s] does not participate in keygen [%s] ", senderAddress, msg.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(types.ModuleCdc.MustMarshalJSON(msg.Payload)))))

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgKeygenStart(ctx sdk.Context, k keeper.Keeper, s types.Snapshotter, staker types.StakingKeeper, v types.Voter, msg *types.MsgKeygenStart) (*sdk.Result, error) {
	// record the snapshot of active validators that we'll use for the key
	snapshotConsensusPower, totalConsensusPower, err := s.TakeSnapshot(ctx, msg.SubsetSize, msg.KeyShareDistributionPolicy)
	if err != nil {
		return nil, err
	}

	snapshot, ok := s.GetLatestSnapshot(ctx)
	if !ok {
		return nil, fmt.Errorf("the system needs to have at least one validator snapshot")
	}

	if !k.GetMinKeygenThreshold(ctx).IsMet(snapshotConsensusPower, totalConsensusPower) {
		msg := fmt.Sprintf(
			"Unable to meet min stake threshold required for keygen: active %s out of %s total",
			snapshotConsensusPower.String(),
			totalConsensusPower.String(),
		)
		k.Logger(ctx).Info(msg)

		return nil, fmt.Errorf(msg)
	}

	threshold := k.ComputeCorruptionThreshold(ctx, snapshot.TotalShareCount)
	if threshold < 1 || snapshot.TotalShareCount.Int64() <= threshold {
		return nil, fmt.Errorf("invalid threshold: %d, total power: %d", threshold, snapshot.TotalShareCount.Int64())
	}

	if err := k.StartKeygen(ctx, v, msg.NewKeyID, snapshot); err != nil {
		return nil, err
	}

	var participants []string
	var participantShareCounts []uint32
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyID, msg.NewKeyID),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatInt(threshold, 10)),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participants))),
			sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantShareCounts))),
		),
	)

	k.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", msg.NewKeyID, threshold, msg.KeyShareDistributionPolicy.SimpleString()))

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgSignTraffic(ctx sdk.Context, k keeper.Keeper, broadcaster types.Broadcaster, msg *types.MsgSignTraffic) (*sdk.Result, error) {
	senderAddress := broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		return nil, fmt.Errorf("invalid message: sender [%s] is not a validator", msg.Sender)
	}

	if !k.DoesValidatorParticipateInSign(ctx, msg.SessionID, senderAddress) {
		return nil, fmt.Errorf("invalid message: sender [%.20s] does not participate in sign [%s] ", senderAddress, msg.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(types.ModuleCdc.MustMarshalJSON(msg.Payload)))))

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgDeregister(ctx sdk.Context, k keeper.Keeper, staker types.StakingKeeper, msg *types.MsgDeregister) (*sdk.Result, error) {
	valAddr := sdk.ValAddress(msg.Sender)

	if _, found := staker.GetValidator(ctx, valAddr); !found {
		return nil, fmt.Errorf("sender %s is not a validator and cannot deregister for tss keygen", valAddr.String())
	}

	k.SetValidatorDeregisteredBlockHeight(ctx, valAddr, ctx.BlockHeight())

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

package keeper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	broadcaster types.Broadcaster
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	voter       types.Voter
	nexus       types.Nexus
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, b types.Broadcaster, s types.Snapshotter, staker types.StakingKeeper, v types.Voter, n types.Nexus) types.MsgServiceServer {
	return msgServer{
		Keeper:      keeper,
		broadcaster: b,
		snapshotter: s,
		staker:      staker,
		voter:       v,
		nexus:       n,
	}
}

func (s msgServer) StartKeygen(c context.Context, req *types.StartKeygenRequest) (*types.StartKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// record the snapshot of active validators that we'll use for the key
	snapshotConsensusPower, totalConsensusPower, err := s.snapshotter.TakeSnapshot(ctx, req.SubsetSize, req.KeyShareDistributionPolicy)
	if err != nil {
		return nil, err
	}

	snapshot, ok := s.snapshotter.GetLatestSnapshot(ctx)
	if !ok {
		return nil, fmt.Errorf("the system needs to have at least one validator snapshot")
	}

	if !s.GetMinKeygenThreshold(ctx).IsMet(snapshotConsensusPower, totalConsensusPower) {
		msg := fmt.Sprintf(
			"Unable to meet min stake threshold required for keygen: active %s out of %s total",
			snapshotConsensusPower.String(),
			totalConsensusPower.String(),
		)
		s.Logger(ctx).Info(msg)

		return nil, fmt.Errorf(msg)
	}

	threshold := s.ComputeCorruptionThreshold(ctx, snapshot.TotalShareCount)
	if threshold < 1 || snapshot.TotalShareCount.Int64() <= threshold {
		return nil, fmt.Errorf("invalid threshold: %d, total power: %d", threshold, snapshot.TotalShareCount.Int64())
	}

	if err := s.Keeper.StartKeygen(ctx, s.voter, req.NewKeyID, snapshot); err != nil {
		return nil, err
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyID, req.NewKeyID),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatInt(threshold, 10)),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participants))),
			sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantShareCounts))),
		),
	)

	s.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", req.NewKeyID, threshold, req.KeyShareDistributionPolicy.SimpleString()))

	return &types.StartKeygenResponse{}, nil
}

func (s msgServer) KeygenTraffic(c context.Context, req *types.KeygenTrafficRequest) (*types.KeygenTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.broadcaster.GetPrincipal(ctx, req.Sender)
	if senderAddress.Empty() {
		return nil, fmt.Errorf("invalid message: sender [%s] is not a validator", req.Sender)
	}

	if !s.DoesValidatorParticipateInKeygen(ctx, req.SessionID, senderAddress) {
		return nil, fmt.Errorf("invalid message: sender [%.20s] does not participate in keygen [%s] ", senderAddress, req.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, req.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(types.ModuleCdc.MustMarshalJSON(req.Payload)))))

	return &types.KeygenTrafficResponse{}, nil
}

func (s msgServer) AssignKey(c context.Context, req *types.AssignKeyRequest) (*types.AssignKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain")
	}

	counter, ok := s.GetSnapshotCounterForKeyID(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("could not find snapshot counter for given key ID")
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not find snapshot for given key ID")
	}

	keyRequirement, found := s.GetKeyRequirement(ctx, chain, req.KeyRole)
	if !found {
		return nil, fmt.Errorf("%s key is not required for chain %s", req.KeyRole.SimpleString(), chain.Name)
	}

	if len(snapshot.Validators) < int(keyRequirement.MinValidatorSubsetSize) {
		return nil, fmt.Errorf(
			"expected %s's %s key to be generated with at least %d validators, actual %d",
			chain.Name,
			req.KeyRole.SimpleString(),
			keyRequirement.MinValidatorSubsetSize,
			len(snapshot.Validators),
		)
	}

	if snapshot.KeyShareDistributionPolicy != keyRequirement.KeyShareDistributionPolicy {
		return nil, fmt.Errorf(
			"expected %s's %s key to have tss shares distributed with policy %s, actual %s",
			chain.Name,
			req.KeyRole.SimpleString(),
			keyRequirement.KeyShareDistributionPolicy.SimpleString(),
			snapshot.KeyShareDistributionPolicy.SimpleString(),
		)
	}

	err := s.AssignNextKey(ctx, chain, req.KeyRole, req.KeyID)
	if err != nil {
		return nil, err
	}

	s.Logger(ctx).Debug(fmt.Sprintf("prepared %s key rotation for chain %s", req.KeyRole.SimpleString(), chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
		),
	)

	return &types.AssignKeyResponse{}, nil
}

func (s msgServer) RotateKey(c context.Context, req *types.RotateKeyRequest) (*types.RotateKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain")
	}

	if err := s.Keeper.RotateKey(ctx, chain, req.KeyRole); err != nil {
		return nil, err
	}

	s.Logger(ctx).Debug(fmt.Sprintf("rotated %s key for chain %s", req.KeyRole.SimpleString(), chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeChain, chain.Name),
		),
	)

	return &types.RotateKeyResponse{}, nil
}

func (s msgServer) VotePubKey(c context.Context, req *types.VotePubKeyRequest) (*types.VotePubKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.GetKey(ctx, req.PollMeta.ID); ok {
		// the key is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("public key %s already verified", req.PollMeta.ID))
		return &types.VotePubKeyResponse{}, nil
	}

	if err := s.voter.TallyVote(ctx, req.Sender, req.PollMeta, req.PubKeyBytes); err != nil {
		return nil, err
	}

	if result := s.voter.Result(ctx, req.PollMeta); result != nil {
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
			sdk.NewAttribute(types.AttributeKeyPoll, req.PollMeta.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(req.PubKeyBytes)),
		))
		switch pkBytes := result.(type) {
		case []byte:
			s.Logger(ctx).Debug(fmt.Sprintf("public key with ID %s confirmed", req.PollMeta.ID))
			btcecPK, err := btcec.ParsePubKey(pkBytes, btcec.S256())
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal public key bytes: [%w]", err)
			}
			pubKey := btcecPK.ToECDSA()
			s.SetKey(ctx, req.PollMeta.ID, *pubKey)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	return &types.VotePubKeyResponse{}, nil

}

func (s msgServer) SignTraffic(c context.Context, req *types.SignTrafficRequest) (*types.SignTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.broadcaster.GetPrincipal(ctx, req.Sender)
	if senderAddress.Empty() {
		return nil, fmt.Errorf("invalid message: sender [%s] is not a validator", req.Sender)
	}

	if !s.DoesValidatorParticipateInSign(ctx, req.SessionID, senderAddress) {
		return nil, fmt.Errorf("invalid message: sender [%.20s] does not participate in sign [%s] ", senderAddress, req.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, req.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(types.ModuleCdc.MustMarshalJSON(req.Payload)))))

	return &types.SignTrafficResponse{}, nil
}

func (s msgServer) VoteSig(c context.Context, req *types.VoteSigRequest) (*types.VoteSigResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.GetSig(ctx, req.PollMeta.ID); ok {
		// the signature is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("signature %s already verified", req.PollMeta.ID))
		return &types.VoteSigResponse{}, nil
	}

	if _, err := btcec.ParseDERSignature(req.SigBytes, btcec.S256()); err != nil {
		return nil, sdkerrors.Wrap(err, "discard vote for invalid signature")
	}

	if err := s.voter.TallyVote(ctx, req.Sender, req.PollMeta, req.SigBytes); err != nil {
		return nil, err
	}

	if result := s.voter.Result(ctx, req.PollMeta); result != nil {
		// the result is not necessarily the same as the msg (the vote could have been decided earlier and now a false vote is cast),
		// so use result instead of msg
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
			sdk.NewAttribute(types.AttributeKeyPoll, req.PollMeta.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(req.SigBytes)),
		))

		switch sigBytes := result.(type) {
		case []byte:
			s.SetSig(ctx, req.PollMeta.ID, sigBytes)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized voting result type: %T", result))
		}
	}

	return &types.VoteSigResponse{}, nil
}

func (s msgServer) Deregister(c context.Context, req *types.DeregisterRequest) (*types.DeregisterResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddr := sdk.ValAddress(req.Sender)

	if _, found := s.staker.GetValidator(ctx, valAddr); !found {
		return nil, fmt.Errorf("sender %s is not a validator and cannot deregister for tss keygen", valAddr.String())
	}

	s.SetValidatorDeregisteredBlockHeight(ctx, valAddr, ctx.BlockHeight())

	return &types.DeregisterResponse{}, nil
}

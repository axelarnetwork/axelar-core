package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strconv"

	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	voter       types.Voter
	nexus       types.Nexus
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, s types.Snapshotter, staker types.StakingKeeper, v types.Voter, n types.Nexus) types.MsgServiceServer {
	return msgServer{
		Keeper:      keeper,
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

func (s msgServer) ProcessKeygenTraffic(c context.Context, req *types.ProcessKeygenTrafficRequest) (*types.ProcessKeygenTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.snapshotter.GetPrincipal(ctx, req.Sender)
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

	return &types.ProcessKeygenTrafficResponse{}, nil
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

	key, ok := s.GetNextKey(ctx, chain, req.KeyRole)
	if chain.Name == evm.Ethereum.Name && ok {
		return nil, fmt.Errorf(" prevous %s %s key %s is not rotated", chain.Name, req.KeyRole.SimpleString(), key.ID)
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

	_, found := s.GetCurrentKeyID(ctx, chain, req.KeyRole)
	keyID, nextKeyFound := s.GetNextKeyID(ctx, chain, req.KeyRole)
	if found && nextKeyFound && req.KeyRole == tss.MasterKey && !s.Keeper.IsKeyReady(ctx, keyID) {
		return nil, fmt.Errorf("%s %s key is not ready to rotate", chain.Name, req.KeyRole.SimpleString())
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

	poll := s.voter.GetPoll(ctx, req.PollMeta)
	if poll == nil {
		return nil, fmt.Errorf("poll does not exist or is closed")
	}

	snapshot, found := s.snapshotter.GetSnapshot(ctx, poll.ValidatorSnapshotCounter)
	if !found {
		return nil, fmt.Errorf("no snapshot found for counter %d", poll.ValidatorSnapshotCounter)
	}

	if criminals := req.Result.GetCriminals(); criminals != nil {
		for _, criminal := range criminals.Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if _, found := snapshot.GetValidator(criminalAddress); !found {
				return nil, fmt.Errorf("received criminal %s who is not a participant of producing key %s", criminalAddress.String(), req.PollMeta.ID)
			}
		}
	}

	poll, err := s.voter.TallyVote(ctx, req.Sender, req.PollMeta, req.Result)
	if err != nil {
		return nil, err
	}

	result := poll.GetResult()
	if result == nil && !poll.Failed {
		return &types.VotePubKeyResponse{}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeKeygen,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollMeta.String()),
	)
	defer ctx.EventManager().EmitEvent(event)

	if poll.Failed {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		s.voter.DeletePoll(ctx, req.PollMeta)
		s.deleteSnapshotCounterForKeyID(ctx, req.PollMeta.ID)
		s.deleteKeygenStart(ctx, req.PollMeta.ID)
		s.deleteParticipantsInKeygen(ctx, req.PollMeta.ID)

		return &types.VotePubKeyResponse{}, nil
	}

	switch keygenResult := result.(type) {
	case *tofnd.MessageOut_KeygenResult:
		s.voter.DeletePoll(ctx, req.PollMeta)

		if pubKeyBytes := keygenResult.GetPubkey(); pubKeyBytes != nil {
			btcecPK, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal public key bytes: [%w]", err)
			}

			pubKey := btcecPK.ToECDSA()
			s.SetKey(ctx, req.PollMeta.ID, *pubKey)

			ctx.EventManager().EmitEvent(
				event.AppendAttributes(
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
					sdk.NewAttribute(types.AttributeKeyPayload, keygenResult.String()),
				),
			)
			s.Logger(ctx).Info(fmt.Sprintf("public key confirmation result is %.10s", result))

			return &types.VotePubKeyResponse{}, nil
		}

		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		// TODO: the snapshot itself can be deleted too but we need to be more careful with it
		s.deleteSnapshotCounterForKeyID(ctx, req.PollMeta.ID)
		s.deleteKeygenStart(ctx, req.PollMeta.ID)
		s.deleteParticipantsInKeygen(ctx, req.PollMeta.ID)

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		for _, criminal := range keygenResult.GetCriminals().Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			s.Keeper.PenalizeSignCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for generating key %s verified: %s - %s", req.PollMeta.ID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VotePubKeyResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

func (s msgServer) ProcessSignTraffic(c context.Context, req *types.ProcessSignTrafficRequest) (*types.ProcessSignTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.snapshotter.GetPrincipal(ctx, req.Sender)
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

	return &types.ProcessSignTrafficResponse{}, nil
}

func (s msgServer) VoteSig(c context.Context, req *types.VoteSigRequest) (*types.VoteSigResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.GetSig(ctx, req.PollMeta.ID); ok {
		// the signature is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("signature %s already verified", req.PollMeta.ID))
		return &types.VoteSigResponse{}, nil
	}

	poll := s.voter.GetPoll(ctx, req.PollMeta)
	if poll == nil {
		return nil, fmt.Errorf("poll does not exist or is closed")
	}

	snapshot, found := s.snapshotter.GetSnapshot(ctx, poll.ValidatorSnapshotCounter)
	if !found {
		return nil, fmt.Errorf("no snapshot found for counter %d", poll.ValidatorSnapshotCounter)
	}

	if criminals := req.Result.GetCriminals(); criminals != nil {
		for _, criminal := range criminals.Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if _, found := snapshot.GetValidator(criminalAddress); !found {
				return nil, fmt.Errorf("received criminal %s who is not a participant of producing signature %s", criminalAddress.String(), req.PollMeta.ID)
			}
		}
	}

	poll, err := s.voter.TallyVote(ctx, req.Sender, req.PollMeta, req.Result)
	if err != nil {
		return nil, err
	}

	result := poll.GetResult()
	if result == nil && !poll.Failed {
		return &types.VoteSigResponse{}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeSign,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollMeta.String()),
	)
	defer ctx.EventManager().EmitEvent(event)

	if poll.Failed {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		s.voter.DeletePoll(ctx, req.PollMeta)
		s.deleteKeyIDForSig(ctx, req.PollMeta.ID)

		return &types.VoteSigResponse{}, nil
	}

	switch signResult := result.(type) {
	case *tofnd.MessageOut_SignResult:
		s.voter.DeletePoll(ctx, req.PollMeta)

		if signature := signResult.GetSignature(); signature != nil {
			s.SetSig(ctx, req.PollMeta.ID, signature)

			s.Logger(ctx).Info(fmt.Sprintf("signature for %s verified: %.10s", req.PollMeta.ID, hex.EncodeToString(signature)))
			ctx.EventManager().EmitEvent(
				event.AppendAttributes(
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
					sdk.NewAttribute(types.AttributeKeyPayload, signResult.String()),
				),
			)

			return &types.VoteSigResponse{}, nil
		}

		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		s.deleteKeyIDForSig(ctx, req.PollMeta.ID)
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		for _, criminal := range signResult.GetCriminals().Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			s.Keeper.PenalizeSignCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for signature %s verified: %s - %s", req.PollMeta.ID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VoteSigResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

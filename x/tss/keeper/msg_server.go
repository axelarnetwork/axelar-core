package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	gogoprototypes "github.com/gogo/protobuf/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.TSSKeeper
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	voter       types.Voter
	nexus       types.Nexus
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.TSSKeeper, s types.Snapshotter, staker types.StakingKeeper, v types.Voter, n types.Nexus) types.MsgServiceServer {
	return msgServer{
		TSSKeeper:   keeper,
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

	if err := s.TSSKeeper.StartKeygen(ctx, s.voter, req.NewKeyID, snapshot); err != nil {
		return nil, err
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	threshold, found := s.GetCorruptionThreshold(ctx, req.NewKeyID)
	// if this value is set to false, then something is really wrong, since a successful
	// invocation of StartKeygen should automatically set the corruption threshold for the key ID
	if !found {
		return nil, fmt.Errorf("could not find corruption threshold for key ID %s", req.NewKeyID)
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

	senderAddress := s.snapshotter.GetOperator(ctx, req.Sender)
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

func (s msgServer) RotateKey(c context.Context, req *types.RotateKeyRequest) (*types.RotateKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain")
	}

	_, hasActiveKey := s.TSSKeeper.GetCurrentKeyID(ctx, chain, req.KeyRole)
	if !hasActiveKey {
		if err := s.TSSKeeper.AssignNextKey(ctx, chain, req.KeyRole, req.KeyID); err != nil {
			return nil, err
		}
	}

	_, hasNextKeyAssigned := s.TSSKeeper.GetNextKeyID(ctx, chain, req.KeyRole)
	if !hasNextKeyAssigned {
		return nil, fmt.Errorf("no key assigned for rotation yet")
	}

	if err := s.TSSKeeper.RotateKey(ctx, chain, req.KeyRole); err != nil {
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

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	var voteData codec.ProtoMarshaler
	switch res := req.Result.GetKeygenResultData().(type) {
	case *tofnd.MessageOut_KeygenResult_Criminals:
		voteData = res.Criminals
	case *tofnd.MessageOut_KeygenResult_Data:
		if s.HasRecoveryInfos(ctx, voter, req.PollKey.ID) {
			return nil, fmt.Errorf("voter %s already submitted their recovery infos", voter.String())
		}

		infos := res.Data.GetShareRecoveryInfos()
		if infos == nil {
			return nil, fmt.Errorf("could not obtain recovery info from result")
		}

		counter, ok := s.GetSnapshotCounterForKeyID(ctx, req.PollKey.ID)
		if !ok {
			return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", req.PollKey.ID)
		}
		snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
		if !ok {
			return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
		}

		val, ok := snapshot.GetValidator(voter)
		if !ok {
			return nil, fmt.Errorf("could not find validator %s in snapshot #%d", val.String(), counter)
		}

		// check that the number of shares is the same as the number of recovery info
		if val.ShareCount != int64(len(infos)) {
			return nil, fmt.Errorf("number of shares is not the same as the number of recovery infos"+
				" for validator %s (expected %d, received %d)", voter.String(), val.ShareCount, len(infos))
		}

		s.SetRecoveryInfos(ctx, voter, req.PollKey.ID, infos)

		// TODO: in the near future we need to change the voting value to include both the pubkey and
		// and the public recovery infos of all parties. The way that is currently done, if a single
		// party neglects to vote, it will be impossible later on to recover shares. Therefore, tofnd needs
		// to be updated to provide only the public part of the recovery shares, and voting be done on that
		// data + pubkey. We will also need to provide both the (public) recovery data and the pubkey that
		// was voted on, so that tofnd can re-construct the shares.
		//
		// Check issue #694 on axelar-core repo for details.
		voteData = &gogoprototypes.BytesValue{Value: res.Data.GetPubKey()}

	default:
		return nil, fmt.Errorf("invalid data type")
	}

	if _, ok := s.GetKey(ctx, req.PollKey.ID); ok {
		// the key is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("public key %s already verified", req.PollKey.ID))
		return &types.VotePubKeyResponse{}, nil
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)

	if err := poll.Vote(voter, voteData); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VotePubKeyResponse{}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeKeygen,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollKey.String()),
	)
	defer ctx.EventManager().EmitEvent(event)

	if poll.Is(vote.Failed) {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		s.DeleteSnapshotCounterForKeyID(ctx, req.PollKey.ID)
		s.DeleteKeygenStart(ctx, req.PollKey.ID)
		s.DeleteParticipantsInKeygen(ctx, req.PollKey.ID)
		s.DeleteAllRecoveryInfos(ctx, req.PollKey.ID)

		return &types.VotePubKeyResponse{}, nil
	}

	result := poll.GetResult()
	// result should be either the PubKey or Criminals
	switch keygenResult := result.(type) {
	case *gogoprototypes.BytesValue:

		btcecPK, err := btcec.ParsePubKey(keygenResult.GetValue(), btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal public key bytes: [%w]", err)
		}

		pubKey := btcecPK.ToECDSA()
		s.SetKey(ctx, req.PollKey.ID, *pubKey)

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
				sdk.NewAttribute(types.AttributeKeyPayload, keygenResult.String()),
			),
		)

		return &types.VotePubKeyResponse{}, nil
	case *tofnd.MessageOut_CriminalList:
		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		// TODO: the snapshot itself can be deleted too but we need to be more careful with it
		s.DeleteSnapshotCounterForKeyID(ctx, req.PollKey.ID)
		s.DeleteKeygenStart(ctx, req.PollKey.ID)
		s.DeleteParticipantsInKeygen(ctx, req.PollKey.ID)

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		snapshot, found := s.snapshotter.GetSnapshot(ctx, poll.GetSnapshotSeqNo())
		if !found {
			return nil, fmt.Errorf("no snapshot found for counter %d", poll.GetSnapshotSeqNo())
		}

		for _, criminal := range keygenResult.Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if _, found := snapshot.GetValidator(criminalAddress); !found {
				s.Logger(ctx).Info(fmt.Sprintf("received criminal %s who is not a participant of producing key %s", criminalAddress.String(), req.PollKey.ID))
				continue
			}

			s.TSSKeeper.PenalizeSignCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for generating key %s verified: %s - %s", req.PollKey.ID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VotePubKeyResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

func (s msgServer) ProcessSignTraffic(c context.Context, req *types.ProcessSignTrafficRequest) (*types.ProcessSignTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.snapshotter.GetOperator(ctx, req.Sender)
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

	if _, ok := s.GetSig(ctx, req.PollKey.ID); ok {
		// the signature is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("signature %s already verified", req.PollKey.ID))
		return &types.VoteSigResponse{}, nil
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, req.Result); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteSigResponse{}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeSign,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollKey.String()),
	)
	defer ctx.EventManager().EmitEvent(event)

	if poll.Is(vote.Failed) {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		s.DeleteKeyIDForSig(ctx, req.PollKey.ID)

		return &types.VoteSigResponse{}, nil
	}

	result := poll.GetResult()
	switch signResult := result.(type) {
	case *tofnd.MessageOut_SignResult:

		if signature := signResult.GetSignature(); signature != nil {
			s.SetSig(ctx, req.PollKey.ID, signature)

			s.Logger(ctx).Info(fmt.Sprintf("signature for %s verified: %.10s", req.PollKey.ID, hex.EncodeToString(signature)))
			ctx.EventManager().EmitEvent(
				event.AppendAttributes(
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
					sdk.NewAttribute(types.AttributeKeyPayload, signResult.String()),
				),
			)

			return &types.VoteSigResponse{}, nil
		}

		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		s.DeleteKeyIDForSig(ctx, req.PollKey.ID)
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		snapshot, found := s.snapshotter.GetSnapshot(ctx, poll.GetSnapshotSeqNo())
		if !found {
			return nil, fmt.Errorf("no snapshot found for counter %d", poll.GetSnapshotSeqNo())
		}
		for _, criminal := range signResult.GetCriminals().Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if _, found := snapshot.GetValidator(criminalAddress); !found {
				s.Logger(ctx).Info(fmt.Sprintf("received criminal %s who is not a participant of producing signature %s",
					criminalAddress.String(), req.PollKey.ID))
				continue
			}
			s.TSSKeeper.PenalizeSignCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for signature %s verified: %s - %s", req.PollKey.ID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VoteSigResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

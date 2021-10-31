package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/armon/go-metrics"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.TSSKeeper
	snapshotter types.Snapshotter
	staker      types.StakingKeeper
	voter       types.Voter
	nexus       types.Nexus
	rewarder    types.Rewarder
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.TSSKeeper, s types.Snapshotter, staker types.StakingKeeper, v types.Voter, n types.Nexus, rewarder types.Rewarder) types.MsgServiceServer {
	return msgServer{
		TSSKeeper:   keeper,
		snapshotter: s,
		staker:      staker,
		voter:       v,
		nexus:       n,
		rewarder:    rewarder,
	}
}

func (s msgServer) RegisterExternalKeys(c context.Context, req *types.RegisterExternalKeysRequest) (*types.RegisterExternalKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, fmt.Errorf("unknown chain %s", req.Chain)
	}

	requiredExternalKeyCount := s.GetExternalMultisigThreshold(ctx).Denominator
	if len(req.ExternalKeys) != int(requiredExternalKeyCount) {
		return nil, fmt.Errorf("%d external keys are required for chain %s", requiredExternalKeyCount, chain.Name)
	}

	keyIDs := make([]exported.KeyID, len(req.ExternalKeys))
	for i, externalKey := range req.ExternalKeys {
		if _, ok := s.GetKey(ctx, externalKey.ID); ok {
			return nil, fmt.Errorf("external key ID %s is already used", externalKey.ID)
		}

		pubKey, err := btcec.ParsePubKey(externalKey.PubKey, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("invalid external key received")
		}

		s.SetKey(ctx, externalKey.ID, *pubKey.ToECDSA())
		s.SetKeyInfo(ctx, types.KeyInfo{KeyID: externalKey.ID, KeyRole: exported.ExternalKey})
		keyIDs[i] = externalKey.ID

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeKey,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueAssigned),
			sdk.NewAttribute(types.AttributeChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyRole, exported.ExternalKey.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyKeyID, string(externalKey.ID)),
		))
	}

	s.SetExternalKeyIDs(ctx, chain, keyIDs)

	return &types.RegisterExternalKeysResponse{}, nil
}

func (s msgServer) HeartBeat(c context.Context, req *types.HeartBeatRequest) (*types.HeartBeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddr := s.snapshotter.GetOperator(ctx, req.Sender)
	if valAddr.Empty() {
		return nil, fmt.Errorf("sender [%s] is not a validator", req.Sender)
	}

	s.Logger(ctx).Debug(fmt.Sprintf("updating availability of operator %s (proxy address %s) for keys %v at block %d",
		valAddr.String(), req.Sender, req.KeyIDs, ctx.BlockHeight()))
	s.SetAvailableOperator(ctx, valAddr, req.KeyIDs...)
	s.rewarder.GetPool(ctx, types.ModuleName).ReleaseRewards(valAddr)

	// this explicit type cast is necessary, because snapshot needs to call UnpackInterfaces() on the validator
	// and it is not exposed in the ValidatorI interface
	validator, ok := s.staker.Validator(ctx, valAddr).(stakingtypes.Validator)
	if !ok {
		s.Logger(ctx).Error(fmt.Sprintf("unexpected validator type: expected %T, got %T", stakingtypes.Validator{}, validator))
		return &types.HeartBeatResponse{}, nil
	}

	illegibility, err := s.snapshotter.GetValidatorIllegibility(ctx, &validator)
	if err != nil {
		s.Logger(ctx).Error(err.Error())
		return &types.HeartBeatResponse{}, nil
	}

	response := &types.HeartBeatResponse{
		KeygenIllegibility:  illegibility.FilterIllegibilityForNewKey(),
		SigningIllegibility: illegibility.FilterIllegibilityForSigning(),
	}

	if !response.KeygenIllegibility.Is(snapshot.None) {
		s.Logger(ctx).Error(fmt.Sprintf("validator %s not ready to participate in keygen due to: %s", valAddr.String(), response.KeygenIllegibility.String()))
	}
	if !response.SigningIllegibility.Is(snapshot.None) {
		s.Logger(ctx).Error(fmt.Sprintf("validator %s not ready to participate in signing due to: %s", valAddr.String(), response.SigningIllegibility.String()))
	}

	// metrics for heartbeat
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "heartbeat"},
		1,
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("address", valAddr.String()),
			telemetry.NewLabel("height", strconv.FormatInt(ctx.BlockHeight(), 10)),
			telemetry.NewLabel("key_IDs", strings.Join(exported.KeyIDsToStrings(req.KeyIDs), ",")),
			telemetry.NewLabel("keygen_ineligibilities", response.KeygenIllegibility.String()),
			telemetry.NewLabel("sign_ineligibilities", response.SigningIllegibility.String()),
		})

	return response, nil
}

func (s msgServer) StartKeygen(c context.Context, req *types.StartKeygenRequest) (*types.StartKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyRequirement, ok := s.GetKeyRequirement(ctx, req.KeyInfo.KeyRole, req.KeyInfo.KeyType)
	if !ok {
		return nil, fmt.Errorf("key requirement for key role %s type %s not found", req.KeyInfo.KeyRole.SimpleString(), req.KeyInfo.KeyType.SimpleString())
	}

	// record the snapshot of active validators that we'll use for the key
	snapshot, err := s.snapshotter.TakeSnapshot(ctx, keyRequirement)
	if err != nil {
		return nil, err
	}

	if err := s.TSSKeeper.StartKeygen(ctx, s.voter, req.KeyInfo, snapshot); err != nil {
		return nil, err
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyType, req.KeyInfo.KeyType.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyKeyID, string(req.KeyInfo.KeyID)),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatInt(snapshot.CorruptionThreshold, 10)),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participants))),
			sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantShareCounts))),
			sdk.NewAttribute(types.AttributeKeyTimeout, strconv.FormatInt(keyRequirement.KeygenTimeout, 10)),
		),
	)

	s.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", req.KeyInfo.KeyID, snapshot.CorruptionThreshold, keyRequirement.KeyShareDistributionPolicy.SimpleString()))

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "corruption", "threshold"},
		float32(snapshot.CorruptionThreshold),
		[]metrics.Label{telemetry.NewLabel("keyID", string(req.KeyInfo.KeyID))})

	minKeygenThreshold := keyRequirement.MinKeygenThreshold
	telemetry.SetGauge(float32(minKeygenThreshold.Numerator*100/minKeygenThreshold.Denominator), types.ModuleName, "minimum", "keygen", "threshold")

	// metrics for keygen participation
	ts := ctx.BlockTime().Unix()
	for _, validator := range snapshot.Validators {
		telemetry.SetGaugeWithLabels(
			[]string{types.ModuleName, "keygen", "participation"},
			float32(validator.ShareCount),
			[]metrics.Label{
				telemetry.NewLabel("timestamp", strconv.FormatInt(ts, 10)),
				telemetry.NewLabel("keyID", string(req.KeyInfo.KeyID)),
				telemetry.NewLabel("address", validator.GetSDKValidator().GetOperator().String()),
			})
	}

	return &types.StartKeygenResponse{}, nil
}

func (s msgServer) ProcessKeygenTraffic(c context.Context, req *types.ProcessKeygenTrafficRequest) (*types.ProcessKeygenTrafficResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	senderAddress := s.snapshotter.GetOperator(ctx, req.Sender)
	if senderAddress.Empty() {
		return nil, fmt.Errorf("invalid message: sender [%s] is not a validator", req.Sender)
	}

	keyID := exported.KeyID(req.SessionID)
	if err := keyID.Validate(); err != nil {
		return nil, err
	}
	if !s.DoesValidatorParticipateInKeygen(ctx, keyID, senderAddress) {
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
	if hasActiveKey {
		return nil, fmt.Errorf("manual key rotation is only allowed when no key is active")
	}

	if err := s.TSSKeeper.AssertMatchesRequirements(ctx, s.snapshotter, chain, req.KeyID, req.KeyRole); err != nil {
		return nil, sdkerrors.Wrapf(err, "key %s does not match requirements for chain %s key %s type and key role %s",
			req.KeyID, chain.Name, chain.KeyType.SimpleString(), req.KeyRole.SimpleString())
	}

	if err := s.TSSKeeper.AssignNextKey(ctx, chain, req.KeyRole, req.KeyID); err != nil {
		return nil, err
	}

	if err := s.TSSKeeper.RotateKey(ctx, chain, req.KeyRole); err != nil {
		return nil, err
	}

	s.Logger(ctx).Debug(fmt.Sprintf("rotated %s key for chain %s", req.KeyRole.SimpleString(), chain.Name))

	telemetry.IncrCounter(1, types.ModuleName, strings.ToLower(chain.Name), req.KeyRole.SimpleString(), "key", "rotation", "count")
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, strings.ToLower(chain.Name), req.KeyRole.SimpleString(), "key", "id"},
		0,
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("keyID", string(req.KeyID)),
		})

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
	keyID := exported.KeyID(req.PollKey.ID)
	switch res := req.Result.GetKeygenResultData().(type) {
	case *tofnd.MessageOut_KeygenResult_Criminals:
		voteData = res.Criminals
	case *tofnd.MessageOut_KeygenResult_Data:
		if s.HasPrivateRecoveryInfos(ctx, voter, keyID) {
			return nil, fmt.Errorf("voter %s already submitted their private recovery info", voter.String())
		}

		counter, ok := s.GetSnapshotCounterForKeyID(ctx, keyID)
		if !ok {
			return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", keyID)
		}
		snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
		if !ok {
			return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
		}

		val, ok := snapshot.GetValidator(voter)
		if !ok {
			return nil, fmt.Errorf("could not find validator %s in snapshot #%d", val.String(), counter)
		}

		// get pubkey
		pubKey := res.Data.GetPubKey()
		if pubKey == nil {
			return nil, fmt.Errorf("public key is nil")
		}

		s.SetPrivateRecoveryInfo(ctx, voter, keyID, res.Data.GetPrivateRecoverInfo())

		voteData = &types.KeygenVoteData{
			PubKey:            pubKey,
			GroupRecoveryInfo: res.Data.GetGroupRecoverInfo(),
		}

	default:
		return nil, fmt.Errorf("invalid data type")
	}

	if _, ok := s.GetKey(ctx, keyID); ok {
		// the key is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("public key %s already verified", keyID))
		return &types.VotePubKeyResponse{}, nil
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)

	if err := poll.Vote(voter, voteData); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VotePubKeyResponse{Log: fmt.Sprintf("not enough votes to confirm public key %s yet", keyID)}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeKeygen,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollKey.String()),
	)
	defer ctx.EventManager().EmitEvent(event)

	if poll.Is(vote.Failed) {
		s.Logger(ctx).Info(fmt.Sprintf("voting for key '%s' has failed", keyID))

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		s.DeleteSnapshotCounterForKeyID(ctx, keyID)
		s.DeleteKeygenStart(ctx, keyID)
		s.DeleteParticipantsInKeygen(ctx, keyID)
		s.DeleteAllRecoveryInfos(ctx, keyID)

		return &types.VotePubKeyResponse{}, nil
	}

	s.Logger(ctx).Info(fmt.Sprintf("voting for key '%s' has finished", keyID))

	result := poll.GetResult()
	// result should be either KeygenResult or Criminals
	switch keygenResult := result.(type) {
	case *types.KeygenVoteData:
		s.Logger(ctx).Debug(fmt.Sprintf("processing new key '%s'", keyID))

		btcecPK, err := btcec.ParsePubKey(keygenResult.PubKey, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("could not parse public key bytes: [%w]", err)
		}

		pubKey := btcecPK.ToECDSA()
		s.SetKey(ctx, keyID, *pubKey)

		s.SetGroupRecoveryInfo(ctx, keyID, keygenResult.GroupRecoveryInfo)

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
				sdk.NewAttribute(types.AttributeKeyPayload, keygenResult.String()),
			),
		)

		return &types.VotePubKeyResponse{}, nil
	case *tofnd.MessageOut_CriminalList:
		s.Logger(ctx).Debug(fmt.Sprintf("extracting criminal list for poll %s", keyID))

		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		// TODO: the snapshot itself can be deleted too but we need to be more careful with it
		s.DeleteSnapshotCounterForKeyID(ctx, keyID)
		s.DeleteKeygenStart(ctx, keyID)
		s.DeleteParticipantsInKeygen(ctx, keyID)
		s.DeleteAllRecoveryInfos(ctx, keyID)
		poll.AllowOverride()

		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
		)

		for _, criminal := range keygenResult.Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if err := validateCriminal(criminalAddress, poll); err != nil {
				s.Logger(ctx).Error(err.Error())
				continue
			}

			s.TSSKeeper.PenalizeCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for generating key %s verified: %s - %s", keyID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VotePubKeyResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

func validateCriminal(criminal sdk.ValAddress, poll vote.Poll) error {
	criminalFound := false
	for _, voter := range poll.GetVoters() {
		if criminal.Equals(voter.Validator) {
			criminalFound = true
			break
		}
	}

	if !criminalFound {
		return fmt.Errorf("received criminal %s who is not a voter of poll %s", criminal.String(), poll.GetKey().String())
	}

	return nil
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

	if _, status := s.GetSig(ctx, req.PollKey.ID); status == exported.SigStatus_Signed {
		// the signature is already set, no need for further processing of the vote
		s.Logger(ctx).Debug(fmt.Sprintf("signature %s already verified", req.PollKey.ID))
		return &types.VoteSigResponse{}, nil
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	info, ok := s.GetInfoForSig(ctx, req.PollKey.ID)
	if !ok {
		return nil, fmt.Errorf("sig info does not exist")
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, req.Result); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteSigResponse{Log: fmt.Sprintf("not enough votes to confirm signature %s yet", req.PollKey.ID)}, nil
	}

	event := sdk.NewEvent(
		types.EventTypeSign,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, req.PollKey.String()),
		sdk.NewAttribute(types.AttributeKeySigID, req.PollKey.ID),
		sdk.NewAttribute(types.AttributeKeySigModule, info.RequestModule),
		sdk.NewAttribute(types.AttributeKeyParticipants, string(s.GetSignParticipantsAsJSON(ctx, req.PollKey.ID))),
		sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(s.GetSignParticipantsSharesAsJSON(ctx, req.PollKey.ID))),
	)
	defer func() { ctx.EventManager().EmitEvent(event) }()

	if len(info.Metadata) > 0 {
		// module that requests sign is responsible for marshalling metadata to appropriate encoding (expects JSON)
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeySigData, info.Metadata))
	}

	if poll.Is(vote.Failed) {
		event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject))

		s.DeleteInfoForSig(ctx, req.PollKey.ID)

		return &types.VoteSigResponse{}, nil
	}

	result := poll.GetResult()
	switch signResult := result.(type) {
	case *tofnd.MessageOut_SignResult:

		if signature := signResult.GetSignature(); signature != nil {
			s.SetSig(ctx, req.PollKey.ID, signature)
			s.SetSigStatus(ctx, req.PollKey.ID, exported.SigStatus_Signed)

			s.Logger(ctx).Info(fmt.Sprintf("signature for %s verified: %.10s", req.PollKey.ID, hex.EncodeToString(signature)))
			event = event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided),
				sdk.NewAttribute(types.AttributeKeyPayload, signResult.String()),
			)

			r := s.GetRouter()
			if r.HasRoute(info.RequestModule) {
				handler := r.GetRoute(info.RequestModule)
				err := handler(ctx, info)
				if err != nil {
					s.Logger(ctx).Error(fmt.Sprintf("error while routing signature to module %s: %s", info.RequestModule, err))
				}
			}
			return &types.VoteSigResponse{}, nil
		}

		// TODO: allow vote for timeout only if params.TimeoutInBlocks has passed
		s.DeleteInfoForSig(ctx, req.PollKey.ID)
		s.SetSigStatus(ctx, req.PollKey.ID, exported.SigStatus_Aborted)
		event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject))
		poll.AllowOverride()

		for _, criminal := range signResult.GetCriminals().Criminals {
			criminalAddress, _ := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if err := validateCriminal(criminalAddress, poll); err != nil {
				s.Logger(ctx).Error(err.Error())
				continue
			}

			s.TSSKeeper.PenalizeCriminal(ctx, criminalAddress, criminal.GetCrimeType())

			s.Logger(ctx).Info(fmt.Sprintf("criminal for signature %s verified: %s - %s", req.PollKey.ID, criminal.GetPartyUid(), criminal.CrimeType.String()))
		}

		return &types.VoteSigResponse{}, nil
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
			fmt.Sprintf("unrecognized voting result type: %T", result))
	}
}

func (s msgServer) SubmitMultisigPubKeys(c context.Context, req *types.SubmitMultisigPubKeysRequest) (*types.SubmitMultisigPubKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	validator := s.snapshotter.GetOperator(ctx, req.Sender)
	if validator.Empty() {
		return nil, fmt.Errorf("sender [%s] is not a validator", req.Sender)
	}

	counter, ok := s.GetSnapshotCounterForKeyID(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot counter for key ID %s", req.KeyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("could not obtain snapshot for counter %d", counter)
	}

	val, ok := snapshot.GetValidator(validator)
	if !ok {
		return nil, fmt.Errorf("could not find validator %s in snapshot #%d", val.String(), counter)
	}

	if s.IsMultisigKeygenCompleted(ctx, req.KeyID) {
		return nil, fmt.Errorf("multisig keygen %s has completed", req.KeyID)
	}

	if int64(len(req.PubKeyInfos)) != val.ShareCount {
		return nil, fmt.Errorf("expect %d pub keys, got %d", val.ShareCount, len(req.PubKeyInfos))
	}

	s.Logger(ctx).Debug(fmt.Sprintf("stores %s multisig pubkeys for operator %s (proxy address %s): %v",
		req.KeyID, validator.String(), req.Sender, req.PubKeyInfos))

	var pks [][]byte
	for _, info := range req.PubKeyInfos {
		pubKey, err := btcec.ParsePubKey(info.PubKey, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("could not parse public key bytes: [%w]", err)
		}

		sig, err := btcec.ParseDERSignature(info.Signature, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("could not parse signature bytes: [%w]", err)
		}

		d := sha256.Sum256([]byte(validator.String()))
		if !sig.Verify(d[:], pubKey) {
			return nil, fmt.Errorf("signature is invalid")
		}

		pks = append(pks, info.PubKey)
	}

	ok = s.SubmitPubKeys(ctx, req.KeyID, validator, pks...)
	if !ok {
		s.Logger(ctx).Debug(fmt.Sprintf("duplicate multisig pub key detected for validator %s", validator))
		return &types.SubmitMultisigPubKeysResponse{}, fmt.Errorf("duplicate pub key")
	}

	if s.IsMultisigKeygenCompleted(ctx, req.KeyID) {
		s.Logger(ctx).Debug(fmt.Sprintf("multisig keygen %s completed", req.KeyID))
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeKeygen,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyKeyID, string(req.KeyID)),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDecided)),
		)
	}

	return &types.SubmitMultisigPubKeysResponse{}, nil
}

package tss

import (
	"fmt"
	"strconv"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, keeper keeper.Keeper, voter types.Voter, snapshotter types.Snapshotter) []abci.ValidatorUpdate {
	keygenReqs := keeper.GetAllKeygenRequestsAtCurrentHeight(ctx)
	if len(keygenReqs) > 0 {
		keeper.Logger(ctx).Info(fmt.Sprintf("processing %d keygens at height %d", len(keygenReqs), ctx.BlockHeight()))
	}

	for _, request := range keygenReqs {
		counter := snapshotter.GetLatestCounter(ctx) + 1

		keeper.Logger(ctx).Info(fmt.Sprintf("linking available operations to snapshot #%d", counter))
		keeper.LinkAvailableOperatorsToSnapshot(ctx, request.KeyID, exported.AckType_Keygen, counter)

		err := startKeygen(ctx, keeper, voter, snapshotter, &request)
		if err != nil {
			keeper.Logger(ctx).Error(fmt.Sprintf("error starting keygen: %s", err.Error()))
		}

		keeper.DeleteKeygenStart(ctx, request.KeyID)
		keeper.DeleteAvailableOperators(ctx, request.KeyID, exported.AckType_Keygen)
	}

	signInfos := keeper.GetAllSignInfosAtCurrentHeight(ctx)
	if len(signInfos) > 0 {
		keeper.Logger(ctx).Info(fmt.Sprintf("processing %d signs at height %d", len(keygenReqs), ctx.BlockHeight()))
	}

	for _, info := range signInfos {
		err := startSign(ctx, keeper, voter, snapshotter, info)
		if err != nil {
			keeper.Logger(ctx).Error(fmt.Sprintf("error starting signing: %s", err.Error()))
		}

		keeper.DeleteScheduledSign(ctx, info.SigID)
		keeper.DeleteAvailableOperators(ctx, info.SigID, exported.AckType_Sign)
	}

	return nil
}

// initiates a keygen
func startKeygen(
	ctx sdk.Context,
	keeper types.TSSKeeper,
	voter types.Voter,
	snapshotter types.Snapshotter,
	req *types.StartKeygenRequest,
) error {
	keyRequirement, ok := keeper.GetKeyRequirement(ctx, req.KeyRole)
	if !ok {
		return fmt.Errorf("key requirement for key role %s not found", req.KeyRole.SimpleString())
	}

	// record the snapshot of active validators that we'll use for the key
	snapshot, err := snapshotter.TakeSnapshot(ctx, keyRequirement)
	if err != nil {
		return err
	}

	if err := keeper.StartKeygen(ctx, voter, req.KeyID, req.KeyRole, snapshot); err != nil {
		return err
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
			sdk.NewAttribute(types.AttributeKeyKeyID, req.KeyID),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatInt(snapshot.CorruptionThreshold, 10)),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participants))),
			sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantShareCounts))),
			sdk.NewAttribute(types.AttributeKeyTimeout, strconv.FormatInt(keyRequirement.KeygenTimeout, 10)),
		),
	)

	keeper.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", req.KeyID, snapshot.CorruptionThreshold, keyRequirement.KeyShareDistributionPolicy.SimpleString()))

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "corruption", "threshold"},
		float32(snapshot.CorruptionThreshold),
		[]metrics.Label{telemetry.NewLabel("keyID", req.KeyID)})

	minKeygenThreshold := keyRequirement.MinKeygenThreshold
	telemetry.SetGauge(float32(minKeygenThreshold.Numerator*100/minKeygenThreshold.Denominator), types.ModuleName, "minimum", "keygen", "threshold")

	// metrics for keygen participation
	ts := time.Now().Unix()
	for _, validator := range snapshot.Validators {
		telemetry.SetGaugeWithLabels(
			[]string{types.ModuleName, "keygen", "participation"},
			float32(validator.ShareCount),
			[]metrics.Label{
				telemetry.NewLabel("timestamp", strconv.FormatInt(ts, 10)),
				telemetry.NewLabel("keyID", req.KeyID),
				telemetry.NewLabel("address", validator.GetSDKValidator().GetOperator().String()),
			})
	}

	return nil
}

// starts a tss signing protocol using the specified key for the given chain.
func startSign(
	ctx sdk.Context,
	k types.TSSKeeper,
	voter types.InitPoller,
	snapshotter types.Snapshotter,
	info exported.SignInfo,
) error {
	_, status := k.GetSig(ctx, info.SigID)
	if status != exported.SigStatus_Scheduled {
		return fmt.Errorf("sigID '%s' is not scheduled", info.SigID)
	}

	snap, ok := snapshotter.GetSnapshot(ctx, info.SnapshotCounter)
	if !ok {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf("could not find snapshot with sequence number #%d", info.SnapshotCounter)
	}

	activeShareCount, excluded, err := k.SelectSignParticipants(ctx, snapshotter, info.SigID, snap)
	if err != nil {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return err
	}

	nonParticipantShareCounts := make([]int64, len(excluded))
	nonParticipants := make([]string, len(excluded))
	for i, validator := range excluded {
		nonParticipants[i] = validator.GetSDKValidator().String()
		nonParticipantShareCounts[i] = validator.ShareCount
	}

	event := sdk.NewEvent(types.EventTypeSign,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyKeyID, info.KeyID),
		sdk.NewAttribute(types.AttributeKeySigID, info.SigID),
		sdk.NewAttribute(types.AttributeKeyParticipants, string(k.GetSignParticipantsAsJSON(ctx, info.SigID))),
		sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(k.GetSignParticipantsSharesAsJSON(ctx, info.SigID))),
		sdk.NewAttribute(types.AttributeKeyNonParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipants))),
		sdk.NewAttribute(types.AttributeKeyNonParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantShareCounts))),
		sdk.NewAttribute(types.AttributeKeyPayload, string(info.Msg)))

	didStart := false
	defer func() {
		k.Logger(ctx).Info(fmt.Sprintf("Attempted to start signing sigID %s", info.SigID),
			types.AttributeKeyDidStart, strconv.FormatBool(didStart),
			types.AttributeKeySigID, info.SigID,
			types.AttributeKeyParticipants, string(k.GetSignParticipantsAsJSON(ctx, info.SigID)),
			types.AttributeKeyParticipantShareCounts, string(k.GetSignParticipantsSharesAsJSON(ctx, info.SigID)),
			types.AttributeKeyNonParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipants)),
			types.AttributeKeyNonParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantShareCounts)))

		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyDidStart, strconv.FormatBool(didStart)))
		ctx.EventManager().EmitEvent(event)
	}()

	if activeShareCount.LTE(sdk.NewInt(snap.CorruptionThreshold)) {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)

		return fmt.Errorf(fmt.Sprintf("not enough active validators are online: corruption threshold [%d], online share count [%d], total share count [%d]",
			snap.CorruptionThreshold,
			activeShareCount.Int64(),
			snap.TotalShareCount.Int64(),
		))
	}

	key, ok := k.GetKey(ctx, info.KeyID)
	if !ok {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf("key %s not found", info.KeyID)
	}

	keyRequirement, ok := k.GetKeyRequirement(ctx, key.Role)
	if !ok {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf("key requirement for key role %s not found", key.Role.SimpleString())
	}
	event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyTimeout, strconv.FormatInt(keyRequirement.SignTimeout, 10)))

	pollKey := vote.NewPollKey(types.ModuleName, info.SigID)
	if err := voter.InitializePoll(
		ctx,
		pollKey,
		snap.Counter,
		vote.ExpiryAt(0),
		vote.Threshold(keyRequirement.SignVotingThreshold),
	); err != nil {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("starting sign with corruption threshold [%d], online share count [%d], total share count [%d], excluded [%d] validators",
		snap.CorruptionThreshold,
		activeShareCount.Int64(),
		snap.TotalShareCount.Int64(),
		len(nonParticipants),
	))

	k.SetInfoForSig(ctx, info.SigID, info)
	k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Signing)

	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", info.SigID, info.KeyID, string(info.Msg)))

	// metrics for sign participation
	ts := time.Now().Unix()
	for _, validator := range snap.Validators {
		if !k.DoesValidatorParticipateInSign(ctx, info.SigID, validator.GetSDKValidator().GetOperator()) {
			continue
		}

		telemetry.SetGaugeWithLabels(
			[]string{types.ModuleName, "sign", "participation"},
			float32(validator.ShareCount),
			[]metrics.Label{
				telemetry.NewLabel("timestamp", strconv.FormatInt(ts, 10)),
				telemetry.NewLabel("sigID", info.SigID),
				telemetry.NewLabel("address", validator.GetSDKValidator().GetOperator().String()),
			})
	}

	didStart = true
	return nil
}

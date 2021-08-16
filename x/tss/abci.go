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
		var counter int64 = 0
		snap, found := snapshotter.GetLatestSnapshot(ctx)
		if found {
			counter = snap.Counter + 1
		}

		keeper.Logger(ctx).Info(fmt.Sprintf("linking available operations to snapshot #%d", counter))
		keeper.LinkAvailableOperatorsToSnapshot(ctx, request.NewKeyID, exported.AckType_Keygen, counter)

		err := startKeygen(ctx, keeper, voter, snapshotter, &request)
		if err != nil {
			keeper.Logger(ctx).Error(fmt.Sprintf("error starting keygen: %s", err.Error()))
		}

		keeper.DeleteKeygenStart(ctx, request.NewKeyID)
		keeper.DeleteAvailableOperators(ctx, request.NewKeyID, exported.AckType_Keygen)
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

	// record the snapshot of active validators that we'll use for the key
	snapshotConsensusPower, totalConsensusPower, err := snapshotter.TakeSnapshot(ctx, req.SubsetSize, req.KeyShareDistributionPolicy)
	if err != nil {
		return err
	}

	snapshot, ok := snapshotter.GetLatestSnapshot(ctx)
	if !ok {
		return fmt.Errorf("the system needs to have at least one validator snapshot")
	}

	if !keeper.GetMinKeygenThreshold(ctx).IsMet(snapshotConsensusPower, totalConsensusPower) {
		msg := fmt.Sprintf(
			"Unable to meet min stake threshold required for keygen: active %s out of %s total",
			snapshotConsensusPower.String(),
			totalConsensusPower.String(),
		)
		keeper.Logger(ctx).Info(msg)

		return fmt.Errorf(msg)
	}

	if err := keeper.StartKeygen(ctx, voter, req.NewKeyID, snapshot); err != nil {
		return err
	}

	participants := make([]string, 0, len(snapshot.Validators))
	participantShareCounts := make([]uint32, 0, len(snapshot.Validators))
	for _, validator := range snapshot.Validators {
		participants = append(participants, validator.GetSDKValidator().GetOperator().String())
		participantShareCounts = append(participantShareCounts, uint32(validator.ShareCount))
	}

	threshold, found := keeper.GetCorruptionThreshold(ctx, req.NewKeyID)
	// if this value is set to false, then something is really wrong, since a successful
	// invocation of StartKeygen should automatically set the corruption threshold for the key ID
	if !found {
		return fmt.Errorf("could not find corruption threshold for key ID %s", req.NewKeyID)
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

	keeper.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", req.NewKeyID, threshold, req.KeyShareDistributionPolicy.SimpleString()))

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "corruption", "threshold"},
		float32(threshold),
		[]metrics.Label{telemetry.NewLabel("keyID", req.NewKeyID)})

	t := keeper.GetMinKeygenThreshold(ctx)
	telemetry.SetGauge(float32(t.Numerator*100/t.Denominator), types.ModuleName, "minimum", "keygen", "threshold")

	// metrics for keygen participation
	ts := time.Now().Unix()
	for _, validator := range snapshot.Validators {
		telemetry.SetGaugeWithLabels(
			[]string{types.ModuleName, "keygen", "participation"},
			float32(validator.ShareCount),
			[]metrics.Label{
				telemetry.NewLabel("timestamp", strconv.FormatInt(ts, 10)),
				telemetry.NewLabel("keyID", req.NewKeyID),
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
	status := k.GetSigStatus(ctx, info.SigID)
	if status != exported.SigStatus_Scheduled {
		return fmt.Errorf("sigID '%s' is not scheduled", info.SigID)
	}

	snap, ok := snapshotter.GetSnapshot(ctx, info.SnapshotCounter)
	if !ok {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf("could not find snapshot with sequence number #%d", info.SnapshotCounter)
	}

	// for now we recalculate the threshold
	// might make sense to store it with the snapshot after keygen is done.
	threshold, found := k.GetCorruptionThreshold(ctx, info.KeyID)
	if !found {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf("keyID %s has no corruption threshold defined", info.KeyID)
	}

	k.SetSignParticipants(ctx, info.SigID, snap.Validators)

	if !k.MeetsThreshold(ctx, info.SigID, threshold) {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return fmt.Errorf(fmt.Sprintf("not enough active validators are online: threshold [%d], online share count [%d]",
			threshold, k.GetTotalShareCount(ctx, info.SigID)))
	}

	pollKey := vote.NewPollKey(types.ModuleName, info.SigID)
	if err := voter.InitializePoll(ctx, pollKey, snap.Counter, vote.ExpiryAt(0)); err != nil {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Aborted)
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("starting sign with threshold [%d] (need [%d]), online share count [%d]",
		threshold, threshold+1, k.GetTotalShareCount(ctx, info.SigID)))

	k.SetKeyIDForSig(ctx, info.SigID, info.KeyID)
	k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Signing)

	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", info.SigID, info.KeyID, string(info.Msg)))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyID, info.KeyID),
			sdk.NewAttribute(types.AttributeKeySigID, info.SigID),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(k.GetSignParticipantsAsJSON(ctx, info.SigID))),
			sdk.NewAttribute(types.AttributeKeyPayload, string(info.Msg))))

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

	return nil
}

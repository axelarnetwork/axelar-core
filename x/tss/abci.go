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
		keeper.DeleteAtCurrentHeight(ctx, request.NewKeyID, exported.AckType_Keygen)
		keeper.DeleteAvailableOperators(ctx, request.NewKeyID, exported.AckType_Keygen)

		err := startKeygen(ctx, keeper, voter, snapshotter, &request)
		if err != nil {
			keeper.Logger(ctx).Error(fmt.Sprintf("error starting keygen: %s", err.Error()))
		}
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

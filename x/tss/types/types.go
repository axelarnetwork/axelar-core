package types

import (
	"fmt"
	"strconv"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// HexSignature represents a tss signature as hex encoded bytes for use in responses from the query client.
type HexSignature struct {
	R string `json:"r"`
	S string `json:"s"`
}

// NewHexSignatureFromQuerySigResponse converts a QuerySigResponse to a HexSignature
func NewHexSignatureFromQuerySigResponse(sigResp *QuerySigResponse) HexSignature {
	return HexSignature{
		R: hexutil.Encode(sigResp.Signature.R),
		S: hexutil.Encode(sigResp.Signature.S),
	}
}

// ComputeCorruptionThreshold returns corruption threshold to be used by tss.
// (threshold + 1) shares are required to sign
func ComputeCorruptionThreshold(threshold utils.Threshold, totalShareCount sdk.Int) int64 {
	return totalShareCount.MulRaw(threshold.Numerator).QuoRaw(threshold.Denominator).Int64() - 1
}

// StartKeygen initiates a keygen
func StartKeygen(
	ctx sdk.Context,
	keeper TSSKeeper,
	voter Voter,
	snapshotter Snapshotter,
	req *StartKeygenRequest,
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
		sdk.NewEvent(EventTypeKeygen,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, AttributeValueStart),
			sdk.NewAttribute(AttributeKeyKeyID, req.NewKeyID),
			sdk.NewAttribute(AttributeKeyThreshold, strconv.FormatInt(threshold, 10)),
			sdk.NewAttribute(AttributeKeyParticipants, string(ModuleCdc.LegacyAmino.MustMarshalJSON(participants))),
			sdk.NewAttribute(AttributeKeyParticipantShareCounts, string(ModuleCdc.LegacyAmino.MustMarshalJSON(participantShareCounts))),
		),
	)

	keeper.Logger(ctx).Info(fmt.Sprintf("new Keygen: key_id [%s] threshold [%d] key_share_distribution_policy [%s]", req.NewKeyID, threshold, req.KeyShareDistributionPolicy.SimpleString()))

	return nil
}

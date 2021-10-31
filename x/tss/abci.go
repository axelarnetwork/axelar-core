package tss

import (
	"fmt"
	"strconv"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, keeper keeper.Keeper, voter types.Voter, nexus types.Nexus, snapshotter types.Snapshotter) []abci.ValidatorUpdate {
	emitHeartbeatEvent(ctx, keeper, nexus)
	sequentialSign(ctx, keeper.GetSignQueue(ctx), keeper, snapshotter, voter)
	timeoutMultiSigKeygen(ctx, keeper.GetMultisigKeygenQueue(ctx), keeper)

	return nil
}

func emitHeartbeatEvent(ctx sdk.Context, keeper keeper.Keeper, nexus types.Nexus) {
	if ctx.BlockHeight() > 0 && (ctx.BlockHeight()%keeper.GetHeartbeatPeriodInBlocks(ctx)) == 0 {
		var keyInfos []types.KeyInfo
		for _, chain := range nexus.GetChains(ctx) {
			keyType := chain.KeyType
			for _, role := range exported.GetKeyRoles() {
				if currentKey, ok := keeper.GetCurrentKeyID(ctx, chain, role); ok {
					keyInfos = append(keyInfos, types.KeyInfo{KeyID: currentKey, KeyType: keyType})
					oldKeyIDs, err := keeper.GetOldActiveKeyIDs(ctx, chain, role)
					if err != nil {
						keeper.Logger(ctx).Error(fmt.Sprintf("unable to retrieve old keys for chain %s with role %s: %s",
							chain.Name, role.SimpleString(), err))
						continue
					}

					for _, keyID := range oldKeyIDs {
						keyInfos = append(keyInfos, types.KeyInfo{KeyID: keyID, KeyType: keyType})
					}
				}
			}
		}

		bz := types.ModuleCdc.LegacyAmino.MustMarshalJSON(keyInfos)
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeHeartBeat,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSend),
			sdk.NewAttribute(types.AttributeKeyKeyInfos, string(bz)),
		))
	}

}

// sequentialSign limits tss sign within max signing shares
func sequentialSign(ctx sdk.Context, signQueue utils.SequenceKVQueue, k types.TSSKeeper, s types.Snapshotter, voter types.Voter) {
	i := uint64(0)
	signShares := int64(0)
	var signInfo exported.SignInfo

	defer func() {
		ctx.Logger().Debug(fmt.Sprintf("%d active sign shares, %d signatures in queue", signShares, signQueue.Size()))
	}()

	maxSignShares := k.GetMaxSimultaneousSignShares(ctx)
	for signShares < maxSignShares && signQueue.Peek(i, &signInfo) {
		_, sigStatus := k.GetSig(ctx, signInfo.SigID)
		// no need to check if snapshot exists again, sanity check for that passed at this point
		snap, _ := s.GetSnapshot(ctx, signInfo.SnapshotCounter)

		switch sigStatus {
		case exported.SigStatus_Queued:
			signShares += snap.CorruptionThreshold + 1
			if signShares > maxSignShares {
				return
			}
			emitSignStartEvent(ctx, k, voter, signInfo, snap)
			k.SetInfoForSig(ctx, signInfo.SigID, signInfo)
			k.SetSigStatus(ctx, signInfo.SigID, exported.SigStatus_Signing)
			ctx.Logger().Debug(fmt.Sprintf("starting sign %s", signInfo.SigID))
			i++
		case exported.SigStatus_Signing:
			signShares += snap.CorruptionThreshold + 1
			ctx.Logger().Debug(fmt.Sprintf("signing %s in progress", signInfo.SigID))
			i++
		case exported.SigStatus_Signed, exported.SigStatus_Aborted, exported.SigStatus_Invalid:
			signQueue.Dequeue(i, &signInfo)
			ctx.Logger().Debug(fmt.Sprintf("dequeque %s, sign status %s", signInfo.SigID, sigStatus))
		default:
			panic("invalid sig status type")
		}
	}
}

// request proxies to initiate a tss signing protocol using the specified signing metadata
func emitSignStartEvent(ctx sdk.Context, k types.TSSKeeper, voter types.InitPoller, info exported.SignInfo, snap snapshot.Snapshot) {
	var nonParticipantShareCounts []int64
	var nonParticipants []string

	for _, validator := range snap.Validators {
		if !k.DoesValidatorParticipateInSign(ctx, info.SigID, validator.GetSDKValidator().GetOperator()) {
			nonParticipants = append(nonParticipants, validator.GetSDKValidator().String())
			nonParticipantShareCounts = append(nonParticipantShareCounts, validator.ShareCount)
			continue
		}

		// metrics for sign participation
		telemetry.SetGaugeWithLabels(
			[]string{types.ModuleName, "sign", "participation"},
			float32(validator.ShareCount),
			[]metrics.Label{
				telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
				telemetry.NewLabel("sigID", info.SigID),
				telemetry.NewLabel("address", validator.GetSDKValidator().GetOperator().String()),
			})
	}

	// no need to check if these exists again, sanity checks for that passed at this point
	key, _ := k.GetKey(ctx, info.KeyID)
	keyType := k.GetKeyType(ctx, info.KeyID)
	keyRequirement, _ := k.GetKeyRequirement(ctx, key.Role, keyType)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeSign,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyKeyID, string(info.KeyID)),
		sdk.NewAttribute(types.AttributeKeySigID, info.SigID),
		sdk.NewAttribute(types.AttributeKeyParticipants, string(k.GetSignParticipantsAsJSON(ctx, info.SigID))),
		sdk.NewAttribute(types.AttributeKeyParticipantShareCounts, string(k.GetSignParticipantsSharesAsJSON(ctx, info.SigID))),
		sdk.NewAttribute(types.AttributeKeyNonParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipants))),
		sdk.NewAttribute(types.AttributeKeyNonParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantShareCounts))),
		sdk.NewAttribute(types.AttributeKeyPayload, string(info.Msg)),
		sdk.NewAttribute(types.AttributeKeyTimeout, strconv.FormatInt(keyRequirement.SignTimeout, 10)),
	))

	k.Logger(ctx).Info(fmt.Sprintf("next sign: sig_id [%s] key_id [%s] message [%s]", info.SigID, info.KeyID, string(info.Msg)),
		types.AttributeKeySigID, info.SigID,
		types.AttributeKeyParticipants, string(k.GetSignParticipantsAsJSON(ctx, info.SigID)),
		types.AttributeKeyParticipantShareCounts, string(k.GetSignParticipantsSharesAsJSON(ctx, info.SigID)),
		types.AttributeKeyNonParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipants)),
		types.AttributeKeyNonParticipantShareCounts, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantShareCounts)),
		types.AttributeKeyPayload, string(info.Msg),
		types.AttributeKeyTimeout, strconv.FormatInt(keyRequirement.SignTimeout, 10))
}

// timeoutMultiSigKeygen checks timed out multisig keygen and penalize absent participants
func timeoutMultiSigKeygen(ctx sdk.Context, multiSigKeygenQueue utils.SequenceKVQueue, k types.TSSKeeper) {
	var keyIDStr gogoprototypes.StringValue

	for multiSigKeygenQueue.Peek(0, &keyIDStr) {
		keyID := exported.KeyID(keyIDStr.Value)
		timeoutHeight, ok := k.GetMultisigPubKeyTimeout(ctx, keyID)
		if !ok {
			// should not reach here, the queue is controlled by keeper
			panic(fmt.Sprintf("timeout block height for multisig key %s not found", keyID))
		}
		if timeoutHeight > ctx.BlockHeight() {
			return
		}

		// penalize absent validator
		if !k.IsMultisigKeygenCompleted(ctx, keyID) {
			participants := k.GetParticipantsInKeygen(ctx, keyID)
			for _, participant := range participants {
				if !k.HasValidatorSubmittedMultisigPubKey(ctx, keyID, participant) {
					ctx.Logger().Debug(fmt.Sprintf("missing pub keys from %s absent for multisig key %s", participant, keyID))
					k.PenalizeCriminal(ctx, participant, tofnd.CRIME_TYPE_NON_MALICIOUS)
				}
			}

			k.DeleteSnapshotCounterForKeyID(ctx, keyID)
			k.DeleteParticipantsInKeygen(ctx, keyID)
			k.DeleteMultisigKeygen(ctx, keyID)
			ctx.Logger().Debug(fmt.Sprintf("multisig keygen %s timed out", keyID))
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKeygen,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyKeyID, string(keyID)),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)),
			)
		}

		multiSigKeygenQueue.Dequeue(0, &keyIDStr)
	}
}

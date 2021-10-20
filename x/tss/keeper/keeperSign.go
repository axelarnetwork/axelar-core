package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const signQueueName = "signqueue"

// EnqueueSign enqueue the pending sign info into a queue and returns the position of the added sign info.
// Returns error if queue is full
func (k Keeper) EnqueueSign(ctx sdk.Context, info exported.SignInfo) (int64, error) {
	status := k.getSigStatus(ctx, info.SigID)
	if status == exported.SigStatus_Signed ||
		status == exported.SigStatus_Signing ||
		status == exported.SigStatus_Scheduled ||
		status == exported.SigStatus_Queued {
		return 0, fmt.Errorf("sig ID '%s' has been used before", info.SigID)
	}

	q := k.GetSignQueue(ctx)
	err := q.Enqueue(&info)
	if err == nil {
		k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Queued)
	}
	return int64(q.Size()), err
}

// ScheduleSign sets a sign to start at block currentHeight + AckWindow and emits events
// to ask vald processes about sending their acknowledgments It returns the height at which it was scheduled
func (k Keeper) ScheduleSign(ctx sdk.Context, info exported.SignInfo) int64 {
	k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Scheduled)

	height := k.GetParams(ctx).AckWindowInBlocks + ctx.BlockHeight()

	key := scheduledSignPrefix.AppendStr(strconv.FormatInt(height, 10)).AppendStr(exported.AckType_Sign.String()).AppendStr(info.SigID)
	k.getStore(ctx).Set(key, &info)

	k.emitAckEvent(ctx, types.AttributeValueSign, info.KeyID, info.SigID, height)
	k.Logger(ctx).Info(fmt.Sprintf(
		"scheduling signing for sig ID '%s' and key ID '%s' at block %d (currently at %d)",
		info.SigID, info.KeyID, height, ctx.BlockHeight()))

	return height
}

// GetAllSignInfosAtCurrentHeight returns all keygen requests scheduled for the current height
func (k Keeper) GetAllSignInfosAtCurrentHeight(ctx sdk.Context) []exported.SignInfo {
	prefix := scheduledSignPrefix.AppendStr(strconv.FormatInt(ctx.BlockHeight(), 10)).AppendStr(exported.AckType_Sign.String())
	var infos []exported.SignInfo

	iter := k.getStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var info exported.SignInfo
		iter.UnmarshalValue(&info)
		infos = append(infos, info)
	}

	return infos
}

// DeleteScheduledSign removes a keygen request for the current height
func (k Keeper) DeleteScheduledSign(ctx sdk.Context, sigID string) {
	key := scheduledSignPrefix.AppendStr(strconv.FormatInt(ctx.BlockHeight(), 10)).
		AppendStr(exported.AckType_Sign.String()).AppendStr(sigID)
	k.getStore(ctx).Delete(key)
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus) {
	status := k.getSigStatus(ctx, sigID)
	if status != exported.SigStatus_Signed {
		return exported.Signature{}, status
	}

	bz := k.getStore(ctx).GetRaw(sigPrefix.AppendStr(sigID))
	if bz == nil {
		return exported.Signature{}, exported.SigStatus_Invalid
	}

	btcecSig, err := btcec.ParseDERSignature(bz, btcec.S256())
	if err != nil {
		return exported.Signature{}, exported.SigStatus_Invalid
	}

	return exported.Signature{R: btcecSig.R, S: btcecSig.S}, exported.SigStatus_Signed
}

// SetSig stores the given signature by its ID
func (k Keeper) SetSig(ctx sdk.Context, sigID string, signature []byte) {
	k.getStore(ctx).SetRaw(sigPrefix.AppendStr(sigID), signature)
}

// GetKeyForSigID returns the key that produced the signature corresponding to the given ID
func (k Keeper) GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool) {
	var info exported.SignInfo
	k.getStore(ctx).Get(infoForSigPrefix.AppendStr(sigID), &info)
	return k.GetKey(ctx, info.KeyID)
}

// SetInfoForSig stores key ID for the given sig ID
func (k Keeper) SetInfoForSig(ctx sdk.Context, sigID string, info exported.SignInfo) {
	k.getStore(ctx).Set(infoForSigPrefix.AppendStr(sigID), &info)
}

// GetInfoForSig stores key ID for the given sig ID
func (k Keeper) GetInfoForSig(ctx sdk.Context, sigID string) (exported.SignInfo, bool) {
	var info exported.SignInfo
	ok := k.getStore(ctx).Get(infoForSigPrefix.AppendStr(sigID), &info)
	return info, ok
}

// DeleteInfoForSig deletes the key ID associated with the given signature
func (k Keeper) DeleteInfoForSig(ctx sdk.Context, sigID string) {
	k.getStore(ctx).Delete(infoForSigPrefix.AppendStr(sigID))
}

// SetSigStatus defines the status of some sign sig ID
func (k Keeper) SetSigStatus(ctx sdk.Context, sigID string, status exported.SigStatus) {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, uint32(status))
	k.getStore(ctx).SetRaw(sigStatusPrefix.AppendStr(sigID), bz)
}

// returns the status of a sig ID
func (k Keeper) getSigStatus(ctx sdk.Context, sigID string) exported.SigStatus {
	bz := k.getStore(ctx).GetRaw(sigStatusPrefix.AppendStr(sigID))
	if bz == nil {
		return exported.SigStatus_Unspecified
	}
	return exported.SigStatus(binary.LittleEndian.Uint32(bz))
}

// SelectSignParticipants appoints a subset of the specified validators to participate in sign ID and returns
// the active share count and excluded validators if no error
func (k Keeper) SelectSignParticipants(ctx sdk.Context, snapshotter types.Snapshotter, sigID string, snap snapshot.Snapshot) ([]snapshot.Validator, []snapshot.Validator, error) {
	var activeValidators []snapshot.Validator
	available := k.GetAvailableOperators(ctx, sigID, exported.AckType_Sign, ctx.BlockHeight())
	validatorAvailable := make(map[string]bool)
	for _, validator := range available {
		validatorAvailable[validator.String()] = true
	}

	var excludedValidators []snapshot.Validator
	validators := snap.Validators

	for _, validator := range validators {
		illegibility, err := snapshotter.GetValidatorIllegibility(ctx, validator.GetSDKValidator())
		if err != nil {
			return nil, nil, err
		}

		if illegibility = illegibility.FilterIllegibilityForSigning(); !illegibility.Is(snapshot.None) {
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s from signing %s due to [%s]",
				validator.GetSDKValidator().GetOperator().String(),
				sigID,
				illegibility.String(),
			))
			excludedValidators = append(excludedValidators, validator)
			continue
		}

		if !validatorAvailable[validator.GetSDKValidator().GetOperator().String()] {
			k.Logger(ctx).Debug(fmt.Sprintf("excluding validator %s from signing %s due to [not-available]",
				validator.GetSDKValidator().GetOperator().String(),
				sigID,
			))
			excludedValidators = append(excludedValidators, validator)
			continue
		}

		activeValidators = append(activeValidators, validator)
	}

	selectedSigners := k.optimizedSigningSet(activeValidators, snap.CorruptionThreshold)

	for _, signer := range selectedSigners {
		k.setParticipateInSign(ctx, sigID, signer.GetSDKValidator().GetOperator(), signer.ShareCount)
	}

	return selectedSigners, activeValidators, nil
}

// selects a subset of the given participants whose total number of shares
// represent the top of the list and amount to at least threshold+1.
func (k Keeper) optimizedSigningSet(activeValidators []snapshot.Validator, threshold int64) []snapshot.Validator {
	if len(activeValidators) == 0 {
		return []snapshot.Validator{}
	}

	sorted := make([]snapshot.Validator, len(activeValidators))
	copy(sorted, activeValidators)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].ShareCount > sorted[j].ShareCount
	})

	var index int
	var totalSubsetShares int64
	for ; index < len(sorted) && totalSubsetShares < (threshold+1); index++ {
		totalSubsetShares = totalSubsetShares + sorted[index].ShareCount
	}

	return sorted[:index]
}

func (k Keeper) setParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress, shareCount int64) {
	key := participatePrefix.AppendStr("sign").AppendStr(sigID).AppendStr(validator.String())
	k.getStore(ctx).SetRaw(key, big.NewInt(shareCount).Bytes())
}

// GetSignParticipants returns the list of participants for specified sig ID
func (k Keeper) GetSignParticipants(ctx sdk.Context, sigID string) []string {
	prefix := participatePrefix.AppendStr("sign").AppendStr(sigID)

	iter := k.getStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	participants := make([]string, 0)
	for ; iter.Valid(); iter.Next() {
		participants = append(participants, strings.TrimPrefix(string(iter.Key()), string(prefix.AsKey())+"_"))
	}

	return participants
}

// GetSignParticipantsShares returns the list of participants share counts for specified sig ID
func (k Keeper) GetSignParticipantsShares(ctx sdk.Context, sigID string) []int64 {
	prefix := participatePrefix.AppendStr("sign").AppendStr(sigID)

	iter := k.getStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	shares := make([]int64, 0)
	for ; iter.Valid(); iter.Next() {
		shares = append(shares, big.NewInt(0).SetBytes(iter.Value()).Int64())
	}

	return shares
}

// GetSignParticipantsAsJSON returns the list of participants for specified sig ID in JSON format
func (k Keeper) GetSignParticipantsAsJSON(ctx sdk.Context, sigID string) []byte {
	list, _ := json.Marshal(k.GetSignParticipants(ctx, sigID))
	return list
}

// GetSignParticipantsSharesAsJSON returns the list of participant share counts for specified sig ID in JSON format
func (k Keeper) GetSignParticipantsSharesAsJSON(ctx sdk.Context, sigID string) []byte {
	list, _ := json.Marshal(k.GetSignParticipantsShares(ctx, sigID))
	return list
}

// DoesValidatorParticipateInSign returns true if given validator participates in signing for the given sig ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool {
	return k.getStore(ctx).Has(participatePrefix.AppendStr("sign").AppendStr(sigID).AppendStr(validator.String()))
}

// PenalizeCriminal penalizes the criminal caught during tss protocol according to the given crime type
func (k Keeper) PenalizeCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd.MessageOut_CriminalList_Criminal_CrimeType) {
	switch crimeType {
	// currently we do not distinguish between malicious and non-malicious faults
	case tofnd.CRIME_TYPE_MALICIOUS, tofnd.CRIME_TYPE_NON_MALICIOUS:
		k.setTssSuspendedUntil(ctx, criminal, ctx.BlockHeight()+k.GetParams(ctx).SuspendDurationInBlocks)
	default:
		k.Logger(ctx).Info(fmt.Sprintf("no policy is set to penalize validator %s for crime type %s", criminal.String(), crimeType.String()))
	}
}

// GetSignQueue returns the sign queue
func (k Keeper) GetSignQueue(ctx sdk.Context) utils.SequenceKVQueue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(signQueueName))
	return utils.NewSequenceKVQueue(utils.NewNormalizedStore(store, k.cdc), uint64(k.getMaxSignQueueSize(ctx)), k.Logger(ctx))
}

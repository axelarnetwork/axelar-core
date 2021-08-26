package keeper

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ScheduleSign sets a sign to start at block currentHeight + AckWindow and emits events
// to ask vald processes about sending their acknowledgments It returns the height at which it was scheduled
func (k Keeper) ScheduleSign(ctx sdk.Context, info exported.SignInfo) (int64, error) {
	status := k.getSigStatus(ctx, info.SigID)
	if status == exported.SigStatus_Signed ||
		status == exported.SigStatus_Signing ||
		status == exported.SigStatus_Scheduled {
		return -1, fmt.Errorf("sig ID '%s' has been used before", info.SigID)
	}
	k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Scheduled)

	height := k.GetParams(ctx).AckWindowInBlocks + ctx.BlockHeight()

	key := fmt.Sprintf("%s%d_%s_%s", scheduledSignPrefix, height, exported.AckType_Sign.String(), info.SigID)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(key), bz)

	k.emitAckEvent(ctx, types.AttributeValueSign, info.KeyID, info.SigID, height)
	k.Logger(ctx).Info(fmt.Sprintf(
		"scheduling signing for sig ID '%s' and key ID '%s' at block %d (currently at %d)",
		info.SigID, info.KeyID, height, ctx.BlockHeight()))

	return height, nil
}

// GetAllSignInfosAtCurrentHeight returns all keygen requests scheduled for the current height
func (k Keeper) GetAllSignInfosAtCurrentHeight(ctx sdk.Context) []exported.SignInfo {
	prefix := fmt.Sprintf("%s%d_%s_", scheduledSignPrefix, ctx.BlockHeight(), exported.AckType_Sign.String())
	store := ctx.KVStore(k.storeKey)
	var infos []exported.SignInfo

	iter := sdk.KVStorePrefixIterator(store, []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var info exported.SignInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		infos = append(infos, info)
	}

	return infos
}

// DeleteScheduledSign removes a keygen request for the current height
func (k Keeper) DeleteScheduledSign(ctx sdk.Context, sigID string) {
	key := fmt.Sprintf("%s%d_%s_%s", scheduledSignPrefix, ctx.BlockHeight(), exported.AckType_Sign, sigID)
	ctx.KVStore(k.storeKey).Delete([]byte(key))
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus) {
	status := k.getSigStatus(ctx, sigID)
	if status != exported.SigStatus_Signed {
		return exported.Signature{}, status
	}

	bz := ctx.KVStore(k.storeKey).Get([]byte(sigPrefix + sigID))
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
	ctx.KVStore(k.storeKey).Set([]byte(sigPrefix+sigID), signature)
}

// GetKeyForSigID returns the key that produced the signature corresponding to the given ID
func (k Keeper) GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(infoForSigPrefix + sigID))
	if bz == nil {
		return exported.Key{}, false
	}
	var info exported.SignInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return k.GetKey(ctx, info.KeyID)
}

// SetInfoForSig stores key ID for the given sig ID
func (k Keeper) SetInfoForSig(ctx sdk.Context, sigID string, info exported.SignInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(infoForSigPrefix+sigID), bz)
}

// GetInfoForSig stores key ID for the given sig ID
func (k Keeper) GetInfoForSig(ctx sdk.Context, sigID string) (exported.SignInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(infoForSigPrefix + sigID))
	if bz == nil {
		return exported.SignInfo{}, false
	}
	var info exported.SignInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	return info, true
}

// DeleteInfoForSig deletes the key ID associated with the given signature
func (k Keeper) DeleteInfoForSig(ctx sdk.Context, sigID string) {
	ctx.KVStore(k.storeKey).Delete([]byte(infoForSigPrefix + sigID))
}

// SetSigStatus defines the status of some sign sig ID
func (k Keeper) SetSigStatus(ctx sdk.Context, sigID string, status exported.SigStatus) {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, uint32(status))
	ctx.KVStore(k.storeKey).Set([]byte(sigStatusPrefix+sigID), bz)
}

// returns the status of a sig ID
func (k Keeper) getSigStatus(ctx sdk.Context, sigID string) exported.SigStatus {
	bz := ctx.KVStore(k.storeKey).Get([]byte(sigStatusPrefix + sigID))
	if bz == nil {
		return exported.SigStatus_Unspecified
	}
	return exported.SigStatus(binary.LittleEndian.Uint32(bz))
}

// SelectSignParticipants appoints a subset of the specified validators to participate in sign ID
func (k Keeper) SelectSignParticipants(ctx sdk.Context, snapshotter types.Snapshotter, sigID string, validators []snapshot.Validator) ([]snapshot.Validator, error) {
	activeShareCount := sdk.ZeroInt()
	var activeValidators []snapshot.Validator
	available := k.GetAvailableOperators(ctx, sigID, exported.AckType_Sign, ctx.BlockHeight())
	validatorAvailable := make(map[string]bool)
	for _, validator := range available {
		validatorAvailable[validator.String()] = true
	}

	var excludedValidators []snapshot.Validator

	for _, validator := range validators {
		illegibility, err := snapshotter.GetValidatorIllegibility(ctx, validator.GetSDKValidator())
		if err != nil {
			return nil, err
		}

		if illegibility = illegibility.FilterIllegibilityForSigning(); illegibility != snapshot.None {
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

	for _, v := range activeValidators {
		k.setParticipateInSign(ctx, sigID, v.GetSDKValidator().GetOperator(), v.ShareCount)
		activeShareCount = activeShareCount.AddRaw(v.ShareCount)
	}
	ctx.KVStore(k.storeKey).Set([]byte(participateShareCountPrefix+"sign_"+sigID), activeShareCount.BigInt().Bytes())

	return excludedValidators, nil
}

func (k Keeper) setParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress, shareCount int64) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"sign_"+sigID+validator.String()), big.NewInt(shareCount).Bytes())
}

// MeetsThreshold returns true if the specified signing threshold is met for the given sign ID
func (k Keeper) MeetsThreshold(ctx sdk.Context, sigID string, threshold int64) bool {
	return k.GetTotalShareCount(ctx, sigID) > threshold
}

// GetTotalShareCount returns to total share count for the given key ID
func (k Keeper) GetTotalShareCount(ctx sdk.Context, sigID string) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(participateShareCountPrefix + "sign_" + sigID))
	if bz == nil {
		return -1
	}
	return big.NewInt(0).SetBytes(bz).Int64()

}

// GetSignParticipants returns the list of participants for specified sig ID
func (k Keeper) GetSignParticipants(ctx sdk.Context, sigID string) []string {
	prefix := participatePrefix + "sign_" + sigID
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	participants := make([]string, 0)
	for ; iter.Valid(); iter.Next() {
		participants = append(participants, strings.TrimPrefix(string(iter.Key()), prefix))
	}

	return participants
}

// GetSignParticipantsShares returns the list of participants share counts for specified sig ID
func (k Keeper) GetSignParticipantsShares(ctx sdk.Context, sigID string) []int64 {
	prefix := participatePrefix + "sign_" + sigID
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(prefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	shares := make([]int64, 0)
	for ; iter.Valid(); iter.Next() {
		shares = append(shares, big.NewInt(0).SetBytes(iter.Value()).Int64())
	}

	return shares
}

// GetSignParticipantsAsJSON returns the list of participants for specified sig ID in JSON format
func (k Keeper) GetSignParticipantsAsJSON(ctx sdk.Context, sigID string) []byte {
	return k.cdc.MustMarshalJSON(k.GetSignParticipants(ctx, sigID))
}

// GetSignParticipantsSharesAsJSON returns the list of participant share counts for specified sig ID in JSON format
func (k Keeper) GetSignParticipantsSharesAsJSON(ctx sdk.Context, sigID string) []byte {
	return k.cdc.MustMarshalJSON(k.GetSignParticipantsShares(ctx, sigID))
}

// DoesValidatorParticipateInSign returns true if given validator participates in signing for the given sig ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "sign_" + sigID + validator.String()))
}

// PenalizeCriminal penalizes the criminal caught during tss protocol according to the given crime type
func (k Keeper) PenalizeCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd.MessageOut_CriminalList_Criminal_CrimeType) {
	switch crimeType {
	case tofnd.CRIME_TYPE_MALICIOUS:
		k.setTssSuspendedUntil(ctx, criminal, ctx.BlockHeight()+k.GetParams(ctx).SuspendDurationInBlocks)
	default:
		k.Logger(ctx).Info(fmt.Sprintf("no policy is set to penalize validator %s for crime type %s", criminal.String(), crimeType.String()))
	}
}

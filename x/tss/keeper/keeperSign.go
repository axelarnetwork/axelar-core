package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const signQueueName = "sign_queue"

// StartSign kickstarts signing
func (k Keeper) StartSign(ctx sdk.Context, info exported.SignInfo, snapshotter types.Snapshotter, voter types.InitPoller) error {
	status := k.getSigStatus(ctx, info.SigID)
	if status == exported.SigStatus_Signed ||
		status == exported.SigStatus_Signing ||
		status == exported.SigStatus_Queued {
		return fmt.Errorf("sig ID '%s' has been used before", info.SigID)
	}

	key, ok := k.GetKey(ctx, info.KeyID)
	if !ok {
		return fmt.Errorf("key %s not found", info.KeyID)
	}

	snap, ok := snapshotter.GetSnapshot(ctx, info.SnapshotCounter)
	if !ok {
		return fmt.Errorf("could not find snapshot with sequence number #%d", info.SnapshotCounter)
	}

	participants, active, err := k.selectSignParticipants(ctx, snapshotter, info, snap, key.Type)
	if err != nil {
		return err
	}

	signingShareCount := sdk.ZeroInt()
	for _, p := range participants {
		signingShareCount = signingShareCount.AddRaw(p.ShareCount)
	}

	activeShareCount := sdk.ZeroInt()
	for _, v := range active {
		activeShareCount = activeShareCount.AddRaw(v.ShareCount)
	}

	if signingShareCount.LTE(sdk.NewInt(snap.CorruptionThreshold)) {
		return fmt.Errorf(fmt.Sprintf("not enough active validators are online: corruption threshold [%d], online share count [%d], total share count [%d]",
			snap.CorruptionThreshold,
			activeShareCount.Int64(),
			snap.TotalShareCount.Int64(),
		))
	}

	keyRequirement, ok := k.GetKeyRequirement(ctx, key.Role, key.Type)
	if !ok {
		return fmt.Errorf("key requirement for %s and %s not found", key.Role, key.Type)
	}

	switch key.Type {
	case exported.Multisig:
		// init multisig key info
		multisigSignInfo := types.MultisigInfo{
			ID:        info.SigID,
			Timeout:   ctx.BlockHeight() + keyRequirement.SignTimeout,
			TargetNum: snap.CorruptionThreshold + 1,
		}
		k.SetMultisigSignInfo(ctx, multisigSignInfo)
		// enqueue ongoing multisig sign
		q := k.GetMultisigSignQueue(ctx)
		// TODO: the sign will be queued and may not start right away, it might affect the sign timeout
		if err := q.Enqueue(&gogoprototypes.StringValue{Value: info.SigID}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid key type %s", key.Type)
	}

	q := k.GetSignQueue(ctx)
	err = q.Enqueue(&info)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("enqueued sign with corruption threshold [%d], signing share count [%d], online share count [%d], total share count [%d], excluded [%d] validators",
		snap.CorruptionThreshold,
		signingShareCount.Int64(),
		activeShareCount.Int64(),
		snap.TotalShareCount.Int64(),
		len(snap.Validators)-len(participants),
	))

	k.SetSigStatus(ctx, info.SigID, exported.SigStatus_Queued)
	return nil
}

func (k Keeper) getSignedSigs(ctx sdk.Context) (sigs []exported.Signature) {
	iter := k.getStore(ctx).Iterator(sigPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))
	for ; iter.Valid(); iter.Next() {
		var sig exported.Signature
		iter.UnmarshalValue(&sig)

		if sig.SigStatus != exported.SigStatus_Signed {
			continue
		}

		sigs = append(sigs, sig)
	}

	return sigs
}

func (k Keeper) getSig(ctx sdk.Context, sigID string) (sig exported.Signature, ok bool) {
	return sig, k.getStore(ctx).Get(sigPrefix.AppendStr(sigID), &sig)
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus) {
	sig, ok := k.getSig(ctx, sigID)
	if !ok {
		return sig, exported.SigStatus_Invalid
	}

	return sig, sig.SigStatus
}

// SetSig stores the given signature
func (k Keeper) SetSig(ctx sdk.Context, signature exported.Signature) {
	k.getStore(ctx).Set(sigPrefix.AppendStr(signature.SigID), &signature)
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
	sig, _ := k.getSig(ctx, sigID)
	sig.SigID = sigID
	sig.SigStatus = status

	k.SetSig(ctx, sig)
}

// returns the status of a sig ID
func (k Keeper) getSigStatus(ctx sdk.Context, sigID string) exported.SigStatus {
	sig, ok := k.getSig(ctx, sigID)
	if !ok {
		return exported.SigStatus_Unspecified
	}

	return sig.SigStatus
}

// selectSignParticipants appoints a subset of the specified validators to participate in sign ID and returns
// the active share count and excluded validators if no error
func (k Keeper) selectSignParticipants(ctx sdk.Context, snapshotter types.Snapshotter, info exported.SignInfo, snap snapshot.Snapshot, keyType exported.KeyType) ([]snapshot.Validator, []snapshot.Validator, error) {
	var activeValidators []snapshot.Validator
	for _, validator := range snap.Validators {
		illegibility, err := snapshotter.GetValidatorIllegibility(ctx, validator.GetSDKValidator())
		if err != nil {
			return nil, nil, err
		}

		switch keyType {
		case exported.Threshold:
			if illegibility = illegibility.FilterIllegibilityForTssSigning(); !illegibility.Is(snapshot.None) {
				k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s from signing %s due to [%s]",
					validator.GetSDKValidator().GetOperator().String(),
					info.SigID,
					illegibility.String(),
				))

				continue
			}

			// Check heartbeat
			availableOperators := make(map[string]bool)
			for _, validator := range k.GetAvailableOperators(ctx, info.KeyID) {
				availableOperators[validator.String()] = true
			}

			if !availableOperators[validator.GetSDKValidator().GetOperator().String()] {
				k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s from signing %s due to [not-available]",
					validator.GetSDKValidator().GetOperator().String(),
					info.SigID,
				))

				continue
			}

		case exported.Multisig:
			if illegibility = illegibility.FilterIllegibilityForMultisigSigning(); !illegibility.Is(snapshot.None) {
				k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s from signing %s due to [%s]",
					validator.GetSDKValidator().GetOperator().String(),
					info.SigID,
					illegibility.String(),
				))

				continue
			}
		default:
			return nil, nil, fmt.Errorf("invalid key type %s", keyType.SimpleString())
		}

		activeValidators = append(activeValidators, validator)
	}

	participants := optimizedSigningSet(keyType, activeValidators, snap.CorruptionThreshold)
	for _, participant := range participants {
		k.setParticipateInSign(ctx, info.SigID, participant.GetSDKValidator().GetOperator(), participant.ShareCount)
	}

	return participants, activeValidators, nil
}

// optimizedSigningSet selects a subset of the given participants whose total number of shares
// represent the top of the list and amount to at least threshold+1.
func optimizedSigningSet(keyType exported.KeyType, validators []snapshot.Validator, threshold int64) []snapshot.Validator {
	if len(validators) == 0 {
		return []snapshot.Validator{}
	}

	if keyType == exported.Multisig {
		return validators
	}

	sorted := make([]snapshot.Validator, len(validators))
	copy(sorted, validators)

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
		k.rewarder.GetPool(ctx, types.ModuleName).ClearRewards(criminal)
		k.setSuspendedUntil(ctx, criminal, ctx.BlockHeight()+k.GetParams(ctx).SuspendDurationInBlocks)
	default:
		k.Logger(ctx).Info(fmt.Sprintf("no policy is set to penalize validator %s for crime type %s", criminal.String(), crimeType.String()))
	}
}

// GetSignQueue returns the sign queue
func (k Keeper) GetSignQueue(ctx sdk.Context) utils.SequenceKVQueue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(signQueueName))
	return utils.NewSequenceKVQueue(utils.NewNormalizedStore(store, k.cdc), uint64(k.getMaxSignQueueSize(ctx)), k.Logger(ctx))
}

// SetMultisigSignInfo stores the MultisigInfo for a multisig sign info
func (k Keeper) SetMultisigSignInfo(ctx sdk.Context, info types.MultisigInfo) {
	k.getStore(ctx).Set(multisigSignPrefix.AppendStr(info.ID), &info)
}

// GetMultisigSignInfo returns the MultisigSignInfo
func (k Keeper) GetMultisigSignInfo(ctx sdk.Context, sigID string) (types.MultisigSignInfo, bool) {
	var info types.MultisigInfo
	ok := k.getStore(ctx).Get(multisigSignPrefix.AppendStr(sigID), &info)

	return &info, ok
}

// SubmitSignatures stores signatures a validator has under the given multisig sigID
func (k Keeper) SubmitSignatures(ctx sdk.Context, sigID string, validator sdk.ValAddress, sigs ...[]byte) bool {
	var signInfo types.MultisigInfo
	ok := k.getStore(ctx).Get(multisigSignPrefix.AppendStr(sigID), &signInfo)
	if !ok {
		// the setter is controlled by keeper
		panic(fmt.Sprintf("MultisigSignInfo %s not found", sigID))
	}

	for _, sig := range sigs {
		if signInfo.HasData(sig) {
			return false
		}
	}

	signInfo.AddData(validator, sigs)
	k.SetMultisigSignInfo(ctx, signInfo)

	return true
}

// DeleteMultisigSign deletes the multisig sign info for the given sig ID
func (k Keeper) DeleteMultisigSign(ctx sdk.Context, signID string) {
	k.getStore(ctx).Delete(multisigSignPrefix.AppendStr(signID))
}

// GetMultisigSignQueue returns the multisig sign timeout queue
func (k Keeper) GetMultisigSignQueue(ctx sdk.Context) utils.SequenceKVQueue {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(multisigSignQueue))
	return utils.NewSequenceKVQueue(utils.NewNormalizedStore(store, k.cdc), uint64(k.getMaxSignQueueSize(ctx)), k.Logger(ctx))
}

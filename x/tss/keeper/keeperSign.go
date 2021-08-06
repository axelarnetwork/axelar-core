package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// StartSign starts a tss signing protocol using the specified key for the given chain.
func (k Keeper) StartSign(ctx sdk.Context, voter types.InitPoller, keyID string, sigID string, msg []byte, s snapshot.Snapshot) error {
	if _, ok := k.getKeyIDForSig(ctx, sigID); ok {
		return fmt.Errorf("sigID %s has been used before", sigID)
	}
	k.SetKeyIDForSig(ctx, sigID, keyID)

	// for now we recalculate the threshold
	// might make sense to store it with the snapshot after keygen is done.
	threshold, found := k.GetCorruptionThreshold(ctx, keyID)
	if !found {
		return fmt.Errorf("keyID %s has no corruption threshold defined", keyID)
	}

	var activeValidators []snapshot.Validator
	activeShareCount := sdk.ZeroInt()

	for _, validator := range s.Validators {
		if snapshot.IsValidatorActive(ctx, k.slasher, validator.GetSDKValidator()) && !snapshot.IsValidatorTssSuspended(ctx, k, validator.GetSDKValidator()) {
			activeValidators = append(activeValidators, validator)
			activeShareCount = activeShareCount.AddRaw(validator.ShareCount)
		}
	}

	if activeShareCount.Int64() <= threshold {
		return fmt.Errorf(fmt.Sprintf("not enough active validators are online: threshold [%d], online share count [%d]",
			threshold, activeShareCount.Int64()))
	}

	k.Logger(ctx).Info(fmt.Sprintf("starting sign with threshold [%d] (need [%d]), online share count [%d]",
		threshold, threshold+1, activeShareCount.Int64()))

	// set sign participants
	var participants []string
	for _, v := range activeValidators {
		participants = append(participants, v.GetSDKValidator().GetOperator().String())
		k.setParticipateInSign(ctx, sigID, v.GetSDKValidator().GetOperator())
	}

	pollKey := vote.NewPollKey(types.ModuleName, sigID)
	if err := voter.InitializePoll(ctx, pollKey, s.Counter, vote.ExpiryAt(0)); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("new Sign: sig_id [%s] key_id [%s] message [%s]", sigID, keyID, string(msg)))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyKeyID, keyID),
			sdk.NewAttribute(types.AttributeKeySigID, sigID),
			sdk.NewAttribute(types.AttributeKeyParticipants, string(k.cdc.MustMarshalJSON(participants))),
			sdk.NewAttribute(types.AttributeKeyPayload, string(msg))))

	return nil
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
func (k Keeper) GetSig(ctx sdk.Context, sigID string) (exported.Signature, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(sigPrefix + sigID))
	if bz == nil {
		return exported.Signature{}, false
	}
	btcecSig, err := btcec.ParseDERSignature(bz, btcec.S256())
	if err != nil {
		// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
		panic(err)
	}

	return exported.Signature{R: btcecSig.R, S: btcecSig.S}, true
}

// SetSig stores the given signature by its ID
func (k Keeper) SetSig(ctx sdk.Context, sigID string, signature []byte) {
	ctx.KVStore(k.storeKey).Set([]byte(sigPrefix+sigID), signature)
}

// GetKeyForSigID returns the key that produced the signature corresponding to the given ID
func (k Keeper) GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool) {
	keyID, ok := k.getKeyIDForSig(ctx, sigID)
	if !ok {
		return exported.Key{}, false
	}
	return k.GetKey(ctx, keyID)
}

// SetKeyIDForSig stores key ID for the given sig ID
func (k Keeper) SetKeyIDForSig(ctx sdk.Context, sigID string, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDForSigPrefix+sigID), []byte(keyID))
}

func (k Keeper) getKeyIDForSig(ctx sdk.Context, sigID string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDForSigPrefix + sigID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// DeleteKeyIDForSig deletes the key ID associated with the given signature
func (k Keeper) DeleteKeyIDForSig(ctx sdk.Context, sigID string) {
	ctx.KVStore(k.storeKey).Delete([]byte(keyIDForSigPrefix + sigID))
}

func (k Keeper) setParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"sign_"+sigID+validator.String()), []byte{})
}

// DoesValidatorParticipateInSign returns true if given validator participates in signing for the given sig ID; otherwise, false
func (k Keeper) DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "sign_" + sigID + validator.String()))
}

// PenalizeSignCriminal penalizes the criminal caught during signing according to the given crime type
func (k Keeper) PenalizeSignCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd.MessageOut_CriminalList_Criminal_CrimeType) {
	switch crimeType {
	case tofnd.CRIME_TYPE_MALICIOUS:
		k.setTssSuspendedUntil(ctx, criminal, ctx.BlockHeight()+k.GetParams(ctx).SuspendDurationInBlocks)
	default:
		k.Logger(ctx).Info("no policy is set to penalize validator %s for crime type %s", criminal.String(), crimeType.String())
	}
}

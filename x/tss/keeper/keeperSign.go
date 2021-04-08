package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// StartSign starts a tss signing protocol using the specified key for the given chain.
func (k Keeper) StartSign(ctx sdk.Context, voter types.InitPoller, keyID string, sigID string, msg []byte, s snapshot.Snapshot) error {
	if _, ok := k.getKeyIDForSig(ctx, sigID); ok {
		return fmt.Errorf("sigID %s has been used before", sigID)
	}
	k.setKeyIDForSig(ctx, sigID, keyID)

	// for now we recalculate the threshold
	// might make sense to store it with the snapshot after keygen is done.
	threshold := k.ComputeCorruptionThreshold(ctx, len(s.Validators))

	var activeValidators []snapshot.Validator
	for _, validator := range s.Validators {
		if snapshot.IsValidatorActive(ctx, k.slasher, validator) {
			activeValidators = append(activeValidators, validator)
		}
	}

	if len(activeValidators) <= threshold {
		return fmt.Errorf(fmt.Sprintf("not enough active validators are online: threshold [%d], online [%d]",
			threshold, len(activeValidators)))
	}

	k.Logger(ctx).Info(fmt.Sprintf("starting sign with threshold [%d] (need [%d]), online validators count [%d]",
		threshold, threshold+1, len(activeValidators)))

	// set sign participants
	var participants []string
	for _, v := range activeValidators {
		participants = append(participants, v.GetOperator().String())
		k.setParticipateInSign(ctx, sigID, v.GetOperator())
	}

	poll := vote.NewPollMeta(types.ModuleName, types.EventTypeSign, sigID)
	if err := voter.InitPoll(ctx, poll, s.Counter); err != nil {
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

// SignMsg takes a types.MsgSignTraffic from the chain and relays it to the keygen protocol
func (k Keeper) SignMsg(ctx sdk.Context, msg types.MsgSignTraffic) error {
	senderAddress := k.broadcaster.GetPrincipal(ctx, msg.Sender)
	if senderAddress.Empty() {
		return fmt.Errorf("invalid message: sender [%s] is not a validator", msg.Sender)
	}

	if !k.participatesInSign(ctx, msg.SessionID, senderAddress) {
		return fmt.Errorf("invalid message: sender [%.20s] does not participate in sign [%s] ", senderAddress, msg.SessionID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeSign,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueMsg),
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionID),
			sdk.NewAttribute(sdk.AttributeKeySender, senderAddress.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, string(k.cdc.MustMarshalJSON(msg.Payload)))))

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

func (k Keeper) setKeyIDForSig(ctx sdk.Context, sigID string, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDForSigPrefix+sigID), []byte(keyID))
}

func (k Keeper) getKeyIDForSig(ctx sdk.Context, sigID string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDForSigPrefix + sigID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) setParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(participatePrefix+"sign_"+sigID+validator.String()), []byte{})
}

func (k Keeper) participatesInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(participatePrefix + "sign_" + sigID + validator.String()))
}

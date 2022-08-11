package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/math"
	"github.com/axelarnetwork/utils/slices"
)

// GetKeygenSessionsByExpiry returns all keygen sessions that either expires at
// or goes out of the grace period at the given block height
func (k Keeper) GetKeygenSessionsByExpiry(ctx sdk.Context, expiry int64) []types.KeygenSession {
	var results []types.KeygenSession

	iter := k.getStore(ctx).Iterator(expiryKeygenPrefix.Append(utils.KeyFromInt(expiry)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		keyID := exported.KeyID(iter.Value())
		result, ok := k.getKeygenSession(ctx, keyID)
		if !ok {
			panic(fmt.Errorf("keygen session %s not found", keyID))
		}

		results = append(results, result)
	}

	return results
}

// GetKeygenSession returns a keygen session by key ID if it exists
func (k Keeper) GetKeygenSession(ctx sdk.Context, id exported.KeyID) (types.KeygenSession, bool) {
	return k.getKeygenSession(ctx, id)
}

// GetKey returns the key of the given ID
func (k Keeper) GetKey(ctx sdk.Context, keyID exported.KeyID) (exported.Key, bool) {
	var key types.Key
	ok := k.getStore(ctx).Get(keyPrefix.Append(utils.LowerCaseKey(keyID.String())), &key)
	if !ok {
		return nil, false
	}

	return &key, true
}

// SetKey sets the given key
func (k Keeper) SetKey(ctx sdk.Context, key types.Key) {
	k.setKey(ctx, key)

	participants := key.GetParticipants()
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenCompleted(key.ID)))
	k.Logger(ctx).Info("setting key",
		"key_id", key.ID,
		"participant_count", len(participants),
		"participants", strings.Join(slices.Map(participants, sdk.ValAddress.String), ", "),
		"participants_weight", key.GetParticipantsWeight().String(),
		"bonded_weight", key.Snapshot.BondedWeight.String(),
		"signing_threshold", key.SigningThreshold.String(),
	)
}

// DeleteKeygenSession deletes the keygen session with the given key ID
func (k Keeper) DeleteKeygenSession(ctx sdk.Context, id exported.KeyID) {
	keygen, ok := k.getKeygenSession(ctx, id)
	if !ok {
		return
	}

	k.getStore(ctx).Delete(getKeygenSessionExpiryKey(keygen))
	k.getStore(ctx).Delete(getKeygenSessionKey(id))
}

func (k Keeper) createKeygenSession(ctx sdk.Context, id exported.KeyID, snapshot snapshot.Snapshot) error {
	if _, ok := k.getKeygenSession(ctx, id); ok {
		return fmt.Errorf("key %s already being generated", id)
	}

	if _, ok := k.getKey(ctx, id); ok {
		return fmt.Errorf("key %s already set", id)
	}

	params := k.getParams(ctx)

	expiresAt := ctx.BlockHeight() + params.KeygenTimeout
	keygenSession := types.NewKeygenSession(id, params.KeygenThreshold, params.SigningThreshold, snapshot, expiresAt, params.KeygenGracePeriod)
	if err := keygenSession.ValidateBasic(); err != nil {
		return err
	}

	k.setKeygenSession(ctx, keygenSession)

	participants := snapshot.GetParticipantAddresses()
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenStarted(id, participants)))

	k.Logger(ctx).Info("keygen session started",
		"key_id", id,
		"participant_count", len(participants),
		"participants", strings.Join(slices.Map(participants, sdk.ValAddress.String), ", "),
		"participants_weight", snapshot.GetParticipantsWeight().String(),
		"bonded_weight", snapshot.BondedWeight.String(),
		"keygen_threshold", params.KeygenThreshold.String(),
		"signing_threshold", params.SigningThreshold.String(),
		"expires_at", expiresAt,
	)

	return nil
}

func (k Keeper) setKey(ctx sdk.Context, key types.Key) {
	k.getStore(ctx).Set(keyPrefix.Append(utils.LowerCaseKey(key.ID.String())), &key)
}

func (k Keeper) getKey(ctx sdk.Context, id exported.KeyID) (key types.Key, ok bool) {
	return key, k.getStore(ctx).Get(keyPrefix.Append(utils.LowerCaseKey(id.String())), &key)
}

func (k Keeper) setKeygenSession(ctx sdk.Context, keygen types.KeygenSession) {
	// the deletion is necessary because we may update it to a different location depending on the current state of the session
	k.getStore(ctx).Delete(expiryKeygenPrefix.Append(utils.KeyFromInt(keygen.ExpiresAt)).Append(utils.LowerCaseKey(keygen.GetKeyID().String())))
	k.getStore(ctx).SetRaw(getKeygenSessionExpiryKey(keygen), []byte(keygen.GetKeyID()))

	k.getStore(ctx).Set(getKeygenSessionKey(keygen.GetKeyID()), &keygen)
}

func (k Keeper) getKeygenSession(ctx sdk.Context, id exported.KeyID) (keygen types.KeygenSession, ok bool) {
	return keygen, k.getStore(ctx).Get(getKeygenSessionKey(id), &keygen)
}

func getKeygenSessionExpiryKey(keygen types.KeygenSession) utils.Key {
	expiry := keygen.ExpiresAt
	if keygen.State == exported.Completed {
		expiry = math.Min(keygen.ExpiresAt, keygen.CompletedAt+keygen.GracePeriod+1)
	}

	return expiryKeygenPrefix.Append(utils.KeyFromInt(expiry)).Append(utils.LowerCaseKey(keygen.GetKeyID().String()))
}

func getKeygenSessionKey(id exported.KeyID) utils.Key {
	return keygenPrefix.Append(utils.LowerCaseKey(id.String()))
}

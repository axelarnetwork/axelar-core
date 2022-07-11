package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/math"
	"github.com/axelarnetwork/utils/slices"
)

var (
	keygenPrefix = utils.KeyFromStr("session_keygen")
	keyPrefix    = utils.KeyFromStr("key")
	expiryPrefix = utils.KeyFromStr("expiry")
)

// Keeper provides access to all state changes regarding this module
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
}

// NewKeeper is the constructor for the keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace.WithKeyTable(types.KeyTable()),
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetKeygenSessionsByExpiry returns all keygen sessions that either expires at
// goes out of the grace period at the given block height
func (k Keeper) GetKeygenSessionsByExpiry(ctx sdk.Context, expiry int64) []types.KeygenSession {
	var results []types.KeygenSession

	iter := k.getStore(ctx).Iterator(expiryPrefix.Append(utils.KeyFromInt(expiry)))
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

// SetKey sets the given key
func (k Keeper) SetKey(ctx sdk.Context, key types.Key) {
	k.getStore(ctx).Set(keyPrefix.AppendStr(key.ID.String()), &key)

	participants := key.GetParticipants()
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenCompleted(key.ID)))
	k.Logger(ctx).Info("keygen session completed",
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

func (k Keeper) getParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// createKeygenSession creates a new keygen session with the given key ID and snapshot
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

func (k Keeper) setKeygenSession(ctx sdk.Context, keygen types.KeygenSession) {
	k.getStore(ctx).Delete(expiryPrefix.Append(utils.KeyFromInt(keygen.ExpiresAt)).Append(utils.KeyFromStr(keygen.GetKeyID().String())))
	k.getStore(ctx).SetRaw(getKeygenSessionExpiryKey(keygen), []byte(keygen.GetKeyID()))

	k.getStore(ctx).Set(getKeygenSessionKey(keygen.GetKeyID()), &keygen)
}

// getKeygenSession returns the keygen session with the given key ID
func (k Keeper) getKeygenSession(ctx sdk.Context, id exported.KeyID) (keygen types.KeygenSession, ok bool) {
	return keygen, k.getStore(ctx).Get(getKeygenSessionKey(id), &keygen)
}

// GetKey returns the key with the given key ID
func (k Keeper) getKey(ctx sdk.Context, id exported.KeyID) (key types.Key, ok bool) {
	return key, k.getStore(ctx).Get(keyPrefix.AppendStr(id.String()), &key)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func getKeygenSessionExpiryKey(keygen types.KeygenSession) utils.Key {
	expiry := keygen.ExpiresAt
	if keygen.State == exported.Completed {
		expiry = math.Min(keygen.ExpiresAt, keygen.CompletedAt+keygen.GracePeriod+1)
	}

	return expiryPrefix.Append(utils.KeyFromInt(expiry)).Append(utils.KeyFromStr(keygen.GetKeyID().String()))
}

func getKeygenSessionKey(id exported.KeyID) utils.Key {
	return keygenPrefix.AppendStr(id.String())
}

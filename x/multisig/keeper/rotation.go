package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// GetCurrentKey returns the current key of the given chain
func (k Keeper) GetCurrentKey(ctx sdk.Context, chainName nexus.ChainName) (exported.Key, bool) {
	keyID, ok := k.GetCurrentKeyID(ctx, chainName)
	if !ok {
		return nil, false
	}

	return k.GetKey(ctx, keyID)
}

// GetCurrentKeyID returns the current key ID of the given chain
func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chainName nexus.ChainName) (exported.KeyID, bool) {
	keyEpoch, ok := k.getKeyEpoch(ctx, chainName, k.getKeyRotationCount(ctx, chainName))
	if !ok {
		return "", false
	}

	return keyEpoch.KeyID, true
}

// GetNextKeyID returns the next key ID of the given chain
func (k Keeper) GetNextKeyID(ctx sdk.Context, chainName nexus.ChainName) (exported.KeyID, bool) {
	keyEpoch, ok := k.getKeyEpoch(ctx, chainName, k.getKeyRotationCount(ctx, chainName)+1)
	if !ok {
		return "", false
	}

	return keyEpoch.KeyID, true
}

// AssignKey assigns the given key ID to be the next key of the given chain
func (k Keeper) AssignKey(ctx sdk.Context, chainName nexus.ChainName, keyID exported.KeyID) error {
	key, ok := k.getKey(ctx, keyID)
	if !ok {
		return fmt.Errorf("key %s not found", keyID)
	}
	if key.State != types.Inactive {
		return fmt.Errorf("key %s is already assigned", keyID)
	}

	nextRotationCount := k.getKeyRotationCount(ctx, chainName) + 1
	if _, ok := k.getKeyEpoch(ctx, chainName, nextRotationCount); ok {
		return fmt.Errorf("next key of chain %s already assigned", chainName)
	}

	key.State = types.Assigned
	k.SetKey(ctx, key)
	k.setKeyEpoch(ctx, types.NewKeyEpoch(nextRotationCount, chainName, keyID))

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeyAssigned(chainName, keyID)))
	k.Logger(ctx).Info("new key assigned",
		"chain", chainName,
		"keyID", keyID,
	)

	return nil
}

// RotateKey rotates to the given chain's next key
func (k Keeper) RotateKey(ctx sdk.Context, chainName nexus.ChainName) error {
	nextRotationCount := k.getKeyRotationCount(ctx, chainName) + 1
	keyEpoch, ok := k.getKeyEpoch(ctx, chainName, nextRotationCount)
	if !ok {
		return fmt.Errorf("next key of chain %s not assigned", chainName)
	}

	key := funcs.MustOk(k.getKey(ctx, keyEpoch.GetKeyID()))
	if key.State != types.Assigned {
		panic(fmt.Errorf("key must be assigned when being rotated to"))
	}
	key.State = types.Active

	k.SetKey(ctx, key)
	k.setKeyRotationCount(ctx, chainName, nextRotationCount)

	params := k.getParams(ctx)
	if keyEpoch.Epoch > params.ActiveEpochCount {
		k.deactivateKeyAtEpoch(ctx, chainName, keyEpoch.Epoch-params.ActiveEpochCount)
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeyRotated(chainName, keyEpoch.GetKeyID())))
	k.Logger(ctx).Info("new key rotated",
		"chain", chainName,
		"keyID", keyEpoch.GetKeyID(),
	)

	return nil
}

// GetActiveKeyIDs returns all active keys in reverse temporal order. The first key is the key of the current epoch
func (k Keeper) GetActiveKeyIDs(ctx sdk.Context, chainName nexus.ChainName) []exported.KeyID {
	epochs := k.getStore(ctx).ReverseIterator(keyEpochPrefix.Append(utils.LowerCaseKey(chainName.String())))
	defer utils.CloseLogError(epochs, k.Logger(ctx))

	var keys []exported.KeyID
	for ; epochs.Valid(); epochs.Next() {
		var epoch types.KeyEpoch
		epochs.UnmarshalValue(&epoch)
		key := funcs.MustOk(k.getKey(ctx, epoch.KeyID))

		switch key.State {
		case types.Inactive:
			// assumption: once an epoch is inactive, no older epoch is active so we can return early
			return keys
		case types.Assigned:
			continue
		case types.Active:
			keys = append(keys, key.ID)
		default:
			panic(fmt.Sprintf("unexpected key state %s", key.State.String()))
		}
	}

	// TODO: deactivate old epochs, otherwise this only returns once all epochs are iterated (and returns all keys)
	return keys
}

func (k Keeper) deactivateKeyAtEpoch(ctx sdk.Context, chainName nexus.ChainName, epoch uint64) {
	keyEpoch := funcs.MustOk(k.getKeyEpoch(ctx, chainName, epoch))
	key := funcs.MustOk(k.getKey(ctx, keyEpoch.GetKeyID()))

	key.State = types.Inactive
	k.SetKey(ctx, key)
}

func (k Keeper) getKeyEpoch(ctx sdk.Context, chainName nexus.ChainName, epoch uint64) (keyEpoch types.KeyEpoch, ok bool) {
	return keyEpoch, k.getStore(ctx).Get(keyEpochPrefix.Append(utils.LowerCaseKey(chainName.String())).Append(utils.KeyFromInt(epoch)), &keyEpoch)
}

func (k Keeper) setKeyEpoch(ctx sdk.Context, keyEpoch types.KeyEpoch) {
	k.getStore(ctx).Set(keyEpochPrefix.Append(utils.LowerCaseKey(keyEpoch.Chain.String())).Append(utils.KeyFromInt(keyEpoch.Epoch)), &keyEpoch)
}

func (k Keeper) setKeyRotationCount(ctx sdk.Context, chainName nexus.ChainName, count uint64) {
	k.getStore(ctx).Set(keyRotationCountPrefix.Append(utils.LowerCaseKey(chainName.String())), &gogoprototypes.UInt64Value{Value: count})
}

func (k Keeper) getKeyRotationCount(ctx sdk.Context, chainName nexus.ChainName) uint64 {
	var value gogoprototypes.UInt64Value
	k.getStore(ctx).Get(keyRotationCountPrefix.Append(utils.LowerCaseKey(chainName.String())), &value)

	return value.Value
}

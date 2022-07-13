package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func (k Keeper) GetCurrentKey(ctx sdk.Context, chainName nexus.ChainName) (types.Key, bool) {
	keyID, ok := k.GetCurrentKeyID(ctx, chainName)
	if !ok {
		return types.Key{}, false
	}

	key, ok := k.getKey(ctx, keyID)
	if !ok {
		panic(fmt.Errorf("key %s not found", keyID))
	}

	return key, true
}

func (k Keeper) GetCurrentKeyID(ctx sdk.Context, chainName nexus.ChainName) (exported.KeyID, bool) {
	keyRotation, ok := k.getKeyRotation(ctx, chainName)
	if !ok {
		return "", false
	}

	return k.getKeyID(ctx, keyRotation.Chain, keyRotation.Count)
}

func (k Keeper) AssignKey(ctx sdk.Context, chainName nexus.ChainName, keyID exported.KeyID) error {
	if _, ok := k.getKey(ctx, keyID); !ok {
		return fmt.Errorf("key %s not found", keyID)
	}

	keyRotation, ok := k.getKeyRotation(ctx, chainName)
	if !ok {
		keyRotation = types.NewKeyRotation(chainName)
	}

	if keyRotation.NextAssigned {
		return fmt.Errorf("next key of chain %s already assigned", chainName)
	}

	keyRotation.NextAssigned = true

	k.setKeyRotation(ctx, keyRotation)
	k.setKeyID(ctx, keyRotation.Chain, keyRotation.Count+1, keyID)

	return nil
}

func (k Keeper) RotateKey(ctx sdk.Context, chainName nexus.ChainName) error {
	keyRotation, ok := k.getKeyRotation(ctx, chainName)
	if !ok || !keyRotation.NextAssigned {
		return fmt.Errorf("next key of chain %s not assigned", chainName)
	}

	keyRotation.Count += 1
	keyRotation.NextAssigned = false
	if _, ok := k.getKeyID(ctx, keyRotation.Chain, keyRotation.Count); !ok {
		panic(fmt.Errorf("next key of chain %s not set", chainName))
	}

	k.setKeyRotation(ctx, keyRotation)

	return nil
}

func (k Keeper) getKeyRotation(ctx sdk.Context, chainName nexus.ChainName) (keyRotation types.KeyRotation, ok bool) {
	return keyRotation, k.getStore(ctx).Get(keyRotationPrefix.AppendStr(chainName.String()), &keyRotation)
}

func (k Keeper) setKeyRotation(ctx sdk.Context, keyRotation types.KeyRotation) {
	k.getStore(ctx).Set(keyRotationPrefix.AppendStr(keyRotation.Chain.String()), &keyRotation)
}

func (k Keeper) getKeyID(ctx sdk.Context, chainName nexus.ChainName, rotation uint64) (exported.KeyID, bool) {
	bz := k.getStore(ctx).GetRaw(
		keyIDPrefix.
			AppendStr(chainName.String()).
			Append(utils.KeyFromInt(rotation)),
	)
	if bz == nil {
		return "", false
	}

	return exported.KeyID(bz), true
}

func (k Keeper) setKeyID(ctx sdk.Context, chainName nexus.ChainName, rotation uint64, keyID exported.KeyID) {
	k.getStore(ctx).SetRaw(
		keyIDPrefix.
			AppendStr(chainName.String()).
			Append(utils.KeyFromInt(rotation)),
		[]byte(keyID),
	)
}

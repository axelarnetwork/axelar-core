package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func (k Keeper) getKeyRecoveryInfos(ctx sdk.Context) []types.KeyRecoveryInfo {
	var keyRecoveryInfos []types.KeyRecoveryInfo

	iter := k.getStore(ctx).Iterator(keyRecoveryInfoPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var keyRecoveryInfo types.KeyRecoveryInfo
		iter.UnmarshalValue(&keyRecoveryInfo)

		keyRecoveryInfos = append(keyRecoveryInfos, keyRecoveryInfo)
	}

	return keyRecoveryInfos
}

func (k Keeper) getKeyRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) (keyRecoveryInfo types.KeyRecoveryInfo, ok bool) {
	return keyRecoveryInfo, k.getStore(ctx).Get(keyRecoveryInfoPrefix.AppendStr(string(keyID)), &keyRecoveryInfo)
}

func (k Keeper) setKeyRecoveryInfo(ctx sdk.Context, keyRecoveryInfo types.KeyRecoveryInfo) {
	k.getStore(ctx).Set(keyRecoveryInfoPrefix.AppendStr(string(keyRecoveryInfo.KeyID)), &keyRecoveryInfo)
}

// SetGroupRecoveryInfo sets the group recovery info for the given key ID
func (k Keeper) SetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID, recoveryInfo []byte) {
	keyRecoveryInfo, _ := k.getKeyRecoveryInfo(ctx, keyID)
	keyRecoveryInfo.KeyID = keyID
	keyRecoveryInfo.Public = recoveryInfo

	k.setKeyRecoveryInfo(ctx, keyRecoveryInfo)
}

// GetGroupRecoveryInfo returns the group recovery info for the given key ID
func (k Keeper) GetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) []byte {
	keyRecoveryInfo, ok := k.getKeyRecoveryInfo(ctx, keyID)
	if !ok {
		return nil
	}

	return keyRecoveryInfo.Public
}

// SetPrivateRecoveryInfo sets the private recovery info for the given validator of the given key ID
func (k Keeper) SetPrivateRecoveryInfo(ctx sdk.Context, validator sdk.ValAddress, keyID exported.KeyID, recoveryInfo []byte) {
	keyRecoveryInfo, _ := k.getKeyRecoveryInfo(ctx, keyID)
	keyRecoveryInfo.KeyID = keyID

	if keyRecoveryInfo.Private == nil {
		keyRecoveryInfo.Private = make(map[string][]byte)
	}
	keyRecoveryInfo.Private[validator.String()] = recoveryInfo

	k.setKeyRecoveryInfo(ctx, keyRecoveryInfo)
}

// GetPrivateRecoveryInfo returns the private recovery info for the given validator of the given key ID
func (k Keeper) GetPrivateRecoveryInfo(ctx sdk.Context, validator sdk.ValAddress, keyID exported.KeyID) []byte {
	keyRecoveryInfo, ok := k.getKeyRecoveryInfo(ctx, keyID)
	if !ok {
		return nil
	}

	privateRecoveryInfo, ok := keyRecoveryInfo.Private[validator.String()]
	if !ok {
		return nil
	}

	return privateRecoveryInfo
}

// HasPrivateRecoveryInfo returns true if the private recovery info for the given validator of the given key ID exists
func (k Keeper) HasPrivateRecoveryInfo(ctx sdk.Context, validator sdk.ValAddress, keyID exported.KeyID) bool {
	return k.GetPrivateRecoveryInfo(ctx, validator, keyID) != nil
}

// DeleteKeyRecoveryInfo deletes the key recovery info for the given key ID
func (k Keeper) DeleteKeyRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) {
	k.getStore(ctx).Delete(keyRecoveryInfoPrefix.AppendStr(string(keyID)))
}

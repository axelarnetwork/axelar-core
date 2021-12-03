package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getNonce(ctx sdk.Context) uint64 {
	if bz := k.getStore(ctx).GetRaw(nonceKey); bz != nil {
		return binary.LittleEndian.Uint64(bz)
	}

	return 0
}

func (k Keeper) setNonce(ctx sdk.Context, nonce uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, nonce)

	k.getStore(ctx).SetRaw(nonceKey, bz)
}

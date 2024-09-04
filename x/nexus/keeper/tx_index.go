package keeper

import (
	"crypto/sha256"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.TxIDGenerator = &Keeper{}

func getTxHash(ctx sdk.Context) [32]byte {
	return sha256.Sum256(ctx.TxBytes())
}

func (k Keeper) Next(ctx sdk.Context) ([32]byte, uint64) {
	return getTxHash(ctx), utils.NewCounter[uint64](messageNonceKey, k.getStore(ctx)).Incr(ctx)
}

func (k Keeper) Curr(ctx sdk.Context) ([32]byte, uint64) {
	return getTxHash(ctx), utils.NewCounter[uint64](messageNonceKey, k.getStore(ctx)).Curr(ctx)
}

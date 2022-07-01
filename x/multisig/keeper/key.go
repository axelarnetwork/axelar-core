package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

var _ exported.Key = &key{}

type key struct {
	types.Key
	ctx sdk.Context
	k   Keeper
}

func newKey(ctx sdk.Context, k Keeper, ky types.Key) *key {
	return &key{
		ctx: ctx,
		k:   k,
		Key: ky,
	}
}

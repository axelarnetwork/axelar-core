package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate2To3 returns the handler that performs in-place store migrations
func Migrate2To3(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		params := k.GetParams(ctx)
		defaultParams := types.DefaultParams()

		params.KeygenTimeout = defaultParams.KeygenTimeout
		params.KeygenGracePeriod = defaultParams.KeygenGracePeriod
		params.SigningTimeout = defaultParams.SigningTimeout
		params.SigningGracePeriod = defaultParams.SigningGracePeriod

		k.setParams(ctx, params)
		return nil
	}
}

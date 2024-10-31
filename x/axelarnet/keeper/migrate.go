package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// Migrate6to7 returns the handler that performs in-place store migrations from version 6 to 7
func Migrate6to7(k Keeper, bankK types.BankKeeper, accountK types.AccountKeeper, nexusK types.Nexus, ibcK IBCKeeper) func(ctx sdk.Context) error {

	return func(ctx sdk.Context) error {
		// Failed IBC transfers are held in Axelarnet module account for later retry.
		// This migration escrows tokens back to escrow accounts so that we can use the same code path for retry.
		moduleAccount := accountK.GetModuleAddress(types.ModuleName)

		balances := bankK.SpendableCoins(ctx, moduleAccount)
		for _, coin := range balances {
			asset, err := nexusK.NewLockableAsset(ctx, ibcK, bankK, coin)
			if err != nil {
				k.Logger(ctx).Info(fmt.Sprintf("coin %s is not a lockable asset", coin), "error", err)
				continue
			}

			err = asset.LockFrom(ctx, moduleAccount)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

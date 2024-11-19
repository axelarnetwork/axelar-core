package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Migrate6to7 returns the handler that performs in-place store migrations from version 6 to 7
func Migrate6to7(k Keeper, bankK types.BankKeeper, accountK types.AccountKeeper, nexusK types.Nexus, ibcK IBCKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// Failed IBC transfers are held in Axelarnet module account for later retry.
		// This migration escrows tokens back to escrow accounts so that we can use the same code path for retry.
		err := escrowFundsFromFailedTransfers(ctx, k, bankK, accountK, nexusK, ibcK)
		if err != nil {
			return err
		}

		// All IBC transfer are routed from AxelarIBCAccount after v1.1
		// This migration updates the sender of failed transfers to AxelarIBCAccount for retry
		err = migrateFailedTransfersToAxelarIBCAccount(ctx, k)
		if err != nil {
			return err
		}

		return nil
	}
}

func escrowFundsFromFailedTransfers(ctx sdk.Context, k Keeper, bankK types.BankKeeper, accountK types.AccountKeeper, nexusK types.Nexus, ibcK IBCKeeper) error {
	axelarnetModuleAccount := accountK.GetModuleAddress(types.ModuleName)
	nexusModuleAccount := accountK.GetModuleAddress(nexus.ModuleName)

	balances := bankK.SpendableCoins(ctx, axelarnetModuleAccount)
	for _, coin := range balances {
		asset, err := nexusK.NewLockableAsset(ctx, ibcK, bankK, coin)
		if err != nil {
			k.Logger(ctx).Info(fmt.Sprintf("coin %s is not a lockable asset", coin), "error", err)
			continue
		}

		// Transfer coins from the Axelarnet module account to the Nexus module account for subsequent locking,
		// as the Axelarnet module account is now restricted from sending coins.
		err = bankK.SendCoinsFromModuleToModule(ctx, types.ModuleName, nexus.ModuleName, sdk.NewCoins(asset.GetAsset()))
		if err != nil {
			return err
		}

		err = asset.LockFrom(ctx, nexusModuleAccount)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateFailedTransfersToAxelarIBCAccount(ctx sdk.Context, k Keeper) error {
	transfers := k.getIBCTransfers(ctx)
	for _, transfer := range transfers {
		if transfer.Status != types.TransferFailed {
			continue
		}

		transfer.Sender = types.AxelarIBCAccount
		if err := k.setTransfer(ctx, transfer); err != nil {
			return err
		}
	}

	return nil
}

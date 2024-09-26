package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Migrate7to8 returns the handler that performs in-place store migrations
func Migrate7to8(_ Keeper, bank types.BankKeeper, account types.AccountKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		if err := sendCoinsFromAxelarnetToNexus(ctx, bank, account); err != nil {
			return err
		}

		return nil
	}
}

func sendCoinsFromAxelarnetToNexus(ctx sdk.Context, bank types.BankKeeper, account types.AccountKeeper) error {
	balances := bank.GetAllBalances(ctx, account.GetModuleAddress(axelarnettypes.ModuleName))

	return bank.SendCoinsFromModuleToModule(ctx, axelarnettypes.ModuleName, types.ModuleName, balances)
}

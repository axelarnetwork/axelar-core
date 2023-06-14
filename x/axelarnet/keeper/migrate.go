package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// Migrate5to6 returns the handler that performs in-place store migrations from version 5 to 6
func Migrate5to6(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		addModuleParamCallContractsProposalMinDeposits(ctx, k)

		return nil
	}
}

func addModuleParamCallContractsProposalMinDeposits(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyCallContractsProposalMinDeposits, types.DefaultParams().CallContractsProposalMinDeposits)
}

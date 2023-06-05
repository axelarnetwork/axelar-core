package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Migrate6to7 returns the handler that performs in-place store migrations from version 6 to 7
func Migrate6to7(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		addModuleParamCallContractsProposalMinDeposits(ctx, k)

		return nil
	}
}

func addModuleParamCallContractsProposalMinDeposits(ctx sdk.Context, k Keeper) {
	k.params.Set(ctx, types.KeyCallContractsProposalMinDeposits, types.DefaultParams().CallContractsProposalMinDeposits)
}

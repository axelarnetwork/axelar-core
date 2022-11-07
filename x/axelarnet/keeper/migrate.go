package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate4To5 returns the handler that performs in-place store migrations
// - add ibc_path -> chain_name reverse mapping
func Migrate4To5(k Keeper) func(_ sdk.Context) error {
	return func(ctx sdk.Context) error {
		return migrateIBCPaths(ctx, k)
	}
}

// migrateIBCPaths adds ibc path -> cosmos chain name reverse mapping
func migrateIBCPaths(ctx sdk.Context, k Keeper) error {
	k.Logger(ctx).Info("migrating ibc paths")

	for _, chain := range k.getCosmosChains(ctx) {
		// axelarnet does not have an ibc path
		if chain.Name.Equals(exported.Axelarnet.Name) {
			continue
		}

		if chainName, found := k.GetChainNameByIBCPath(ctx, chain.IBCPath); found {
			return fmt.Errorf("ibc path %s already registered for chain %s when migrating chain %s", chain.IBCPath, chainName, chain.Name)
		}

		k.Logger(ctx).Info(fmt.Sprintf("migrating cosmos chain %s with ibc path %s", chain.Name, chain.IBCPath),
			types.AttributeChain, chain.Name,
			types.AttributeIBCPath, chain.IBCPath,
		)

		if err := k.SetChainByIBCPath(ctx, chain.IBCPath, chain.Name); err != nil {
			return err
		}
	}

	return nil
}

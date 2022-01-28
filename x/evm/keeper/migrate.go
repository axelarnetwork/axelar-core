package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/evm/legacy"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.13 to v0.14. The
// migration includes:
// - Add default absorber bytecode for every existing evm chain
func GetMigrationHandler(k types.BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		bzAbsorber, err := hex.DecodeString(types.Absorber)
		if err != nil {
			return err
		}

		for _, chain := range n.GetChains(ctx) {
			// ignore non-evm chains
			if chain.Module != types.ModuleName {
				continue
			}

			keeper := k.ForChain(chain.Name).(chainKeeper)
			subspace, ok := keeper.getSubspace(ctx)
			if !ok {
				return fmt.Errorf("params for chain %s not set", keeper.GetName())
			}

			var legacyParams legacy.Params
			subspace.GetParamSet(ctx, &legacyParams)

			params := legacyParams.Params
			params.Absorber = bzAbsorber
			subspace.SetParamSet(ctx, &params)
		}

		return nil
	}
}

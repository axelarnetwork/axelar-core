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
		bzGateway, err := hex.DecodeString(types.MultisigGateway)
		if err != nil {
			panic(err)
		}

		bzToken, err := hex.DecodeString(types.Token)
		if err != nil {
			panic(err)
		}

		bzBurnable, err := hex.DecodeString(types.Burnable)
		if err != nil {
			panic(err)
		}

		bzAbsorber, err := hex.DecodeString(types.Absorber)
		if err != nil {
			panic(err)
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
			params.GatewayCode = bzGateway
			params.TokenCode = bzToken
			params.Burnable = bzBurnable
			params.Absorber = bzAbsorber
			subspace.SetParamSet(ctx, &params)
		}

		return nil
	}
}

package keeper

import (
	"encoding/hex"

	"github.com/axelarnetwork/axelar-core/x/evm/legacy"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.13 to v0.14. The
// migration includes:
// - Update contracts' bytecode
// - Set burner code on existing tokens
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

		bzLegacyBurnable, err := hex.DecodeString(legacy.Burnable)
		if err != nil {
			panic(err)
		}

		for _, chain := range n.GetChains(ctx) {
			// ignore non-evm chains
			if chain.Module != types.ModuleName {
				continue
			}

			keeper := k.ForChain(chain.Name)
			params := keeper.GetParams(ctx)

			params.GatewayCode = bzGateway
			params.TokenCode = bzToken
			params.Burnable = bzBurnable

			keeper.SetParams(ctx, params)

			for _, token := range keeper.GetTokens(ctx) {
				token.SaveBurnerCode(bzLegacyBurnable)
			}
		}

		return nil
	}
}

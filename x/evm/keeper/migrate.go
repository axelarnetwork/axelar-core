package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.15 to v0.16. The
// migration includes:
// 	- delete pending gateways of all evm chains
func GetMigrationHandler(k types.BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		for _, chain := range n.GetChains(ctx) {
			if chain.Module != types.ModuleName {
				continue
			}

			keeper := k.ForChain(chain.Name).(chainKeeper)
			if gateway := keeper.getGateway(ctx); gateway.Status == types.GatewayStatusPending {
				keeper.getStore(ctx, keeper.chainLowerKey).Delete(gatewayKey)
			}
		}

		return nil
	}
}

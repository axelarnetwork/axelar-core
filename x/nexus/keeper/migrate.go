package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.11 to v0.13. The
// migration includes:
// - Activate all cosmos chains
// - Delete pending transfers to fee collector and instead add those to transfer fee
func GetMigrationHandler(k Keeper, a types.AxelarnetKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		activateCosmosChains(ctx, k, a)
		addTransferFee(ctx, k, a)

		return nil
	}

}

func activateCosmosChains(ctx sdk.Context, k Keeper, a types.AxelarnetKeeper) {
	for _, chain := range k.GetChains(ctx) {
		if a.IsCosmosChain(ctx, chain.Name) {
			k.ActivateChain(ctx, chain)
		}
	}
}

func addTransferFee(ctx sdk.Context, k Keeper, a types.AxelarnetKeeper) {
	feeCollector, ok := a.GetFeeCollector(ctx)
	if !ok {
		return
	}

	feeCollectorStr := feeCollector.String()

	for _, transfer := range k.getTransfers(ctx) {
		if transfer.State != exported.Pending {
			continue
		}

		recipient := transfer.Recipient

		if recipient.Address != feeCollectorStr {
			continue
		}

		if recipient.Chain.Name != axelarnet.Axelarnet.Name {
			continue
		}
		fmt.Printf("transfer %#v\n", transfer)

		k.deleteTransfer(ctx, transfer)
		k.addTransferFee(ctx, transfer.Asset)
	}
}

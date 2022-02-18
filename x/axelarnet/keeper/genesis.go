package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the axelarnet module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.setParams(ctx, genState.Params)
	if len(genState.CollectorAddress) > 0 {
		if err := k.SetFeeCollector(ctx, genState.CollectorAddress); err != nil {
			panic(err)
		}
	}

	for _, chain := range genState.Chains {
		k.SetCosmosChain(ctx, chain)
	}

	for _, transfer := range genState.PendingTransfers {
		k.SetPendingIBCTransfer(ctx, transfer)
	}
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	collector, _ := k.GetFeeCollector(ctx)

	return types.NewGenesisState(
		k.getParams(ctx),
		collector,
		k.getCosmosChains(ctx),
		k.getPendingIBCTransfers(ctx),
	)
}

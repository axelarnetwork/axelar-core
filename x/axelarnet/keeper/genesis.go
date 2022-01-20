package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the axelarnet module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, nexus types.Nexus, genState *types.GenesisState) {
	k.setParams(ctx, nexus, genState.Params)
	if len(genState.CollectorAddress) > 0 {
		if err := k.SetFeeCollector(ctx, genState.CollectorAddress); err != nil {
			panic(err)
		}
	}

	for _, chain := range genState.Chains {
		k.SetCosmosChain(ctx, chain)

		if err := k.RegisterIBCPath(ctx, chain.Name, chain.IBCPath); err != nil {
			panic(err)
		}

		for _, asset := range chain.Assets {
			if err := k.RegisterAssetToCosmosChain(ctx, asset, chain.Name); err != nil {
				panic(err)
			}
		}
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

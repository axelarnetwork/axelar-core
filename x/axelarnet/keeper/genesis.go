package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the reward module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, nexus types.Nexus, genState *types.GenesisState) {
	k.setParams(ctx, nexus, genState.Params)
	k.SetFeeCollector(ctx, genState.CollectorAddress)
	for _, chain := range genState.Chains {
		k.SetCosmosChain(ctx, chain)

		if err := k.RegisterIBCPath(ctx, chain.Name, chain.IBCPath); err != nil {
			panic(err)
		}

		for _, asset := range chain.Assets {
			k.RegisterAssetToCosmosChain(ctx, asset, chain.Name)
		}
	}

	for _, transfer := range genState.PendingTransfers {
		k.SetPendingIBCTransfer(ctx, transfer)
	}
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var chains []types.CosmosChain
	for _, chainName := range k.GetCosmosChains(ctx) {
		ibcPath, _ := k.GetIBCPath(ctx, chainName)
		assets := k.getAssets(ctx, chainName)
		chain, _ := k.GetCosmosChainByName(ctx, chainName)

		chains = append(chains, types.CosmosChain{
			Name:       chainName,
			Assets:     assets,
			IBCPath:    ibcPath,
			AddrPrefix: chain.AddrPrefix,
		})
	}

	var transfers []types.IBCTransfer
	for _, transfer := range k.getPendingIBCTransfers(ctx) {
		transfers = append(transfers, transfer)
	}

	collector, _ := k.GetFeeCollector(ctx)
	return types.NewGenesisState(k.getParams(ctx), collector, chains, transfers)
}

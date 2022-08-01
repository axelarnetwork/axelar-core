package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
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

	if err := k.validateIBCTransferQueueState(genState.TransferQueue, ibcTransferQueueName); err != nil {
		panic(err)
	}
	k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ImportState(genState.TransferQueue)

	for _, t := range genState.FailedTransfers {
		k.getStore(ctx).SetNew(getFailedTransferKey(t.ID), &t)
	}
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	collector, _ := k.GetFeeCollector(ctx)

	return types.NewGenesisState(
		k.getParams(ctx),
		collector,
		k.getCosmosChains(ctx),
		k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ExportState(),
		k.getFailedTransfers(ctx),
	)
}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// InitGenesis initializes the reward module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	k.setSnapshotCount(ctx, int64(len(genState.Snapshots)))
	for _, snapshot := range genState.Snapshots {
		k.setSnapshot(ctx, snapshot)
	}

	for _, proxiedValidator := range genState.ProxiedValidators {
		k.setProxiedValidator(ctx, proxiedValidator)
	}
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.getSnapshots(ctx),
		k.getProxiedValidators(ctx),
	)
}

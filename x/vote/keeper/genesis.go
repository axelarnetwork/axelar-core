package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, pollMetadata := range genState.PollMetadatas {
		k.newPollStore(ctx, pollMetadata.ID).SetMetadata(pollMetadata)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.getNonPendingPollMetadatas(ctx),
	)
}

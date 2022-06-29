package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
)

// InitGenesis initialize default parameters
// from the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, pollMetadata := range genState.PollMetadatas {
		k.setPollMetadata(ctx, pollMetadata)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		slices.Filter(
			k.getPollMetadatas(ctx),
			func(metadata exported.PollMetadata) bool { return !metadata.Is(exported.Pending) },
		),
	)
}

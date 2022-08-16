package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

// InitGenesis initializes the state from a genesis file
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	k.setParams(ctx, state.Params)
}

// ExportGenesis generates a genesis file from the state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(k.getParams(ctx), []types.KeygenSession{}, []types.SigningSession{}, []types.Key{}, []types.KeyEpoch{})
}

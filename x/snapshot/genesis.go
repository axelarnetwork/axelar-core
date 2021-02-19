package snapshot

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, g types.GenesisState) {
	k.SetParams(ctx, g.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{Params: k.GetParams(ctx)}
}

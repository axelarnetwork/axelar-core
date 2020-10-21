package broadcast

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/keeper"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k keeper.Keeper, g types.GenesisState) {
	k.SetProxyCount(ctx, g.ProxyCount)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	state := types.GenesisState{ProxyCount: k.GetProxyCount(ctx)}

	return state
}

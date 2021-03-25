package bitcoin

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k types.BTCKeeper, g types.GenesisState) {
	k.SetParams(ctx, g.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k types.BTCKeeper) types.GenesisState {
	params := k.GetParams(ctx)
	return types.GenesisState{Params: params}
}

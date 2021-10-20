package nexus

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k types.Nexus, g types.GenesisState) {
	k.SetParams(ctx, g.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k types.Nexus) types.GenesisState {
	params := k.GetParams(ctx)
	return types.GenesisState{Params: params}
}

package axelarnet

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k types.BaseKeeper, n types.Nexus, g types.GenesisState) {
	k.SetParams(ctx, n, g.Params)
}

package axelarnet

import (
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, n types.Nexus, g types.GenesisState) {
	for _, chain := range g.Params.Chains {
		n.RegisterAsset(ctx, exported.Axelarnet.Name, chain.NativeAsset)
	}
}

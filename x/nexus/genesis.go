package nexus

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, g types.GenesisState) {
	k.SetParams(ctx, g.Params)

	// add hardcoded chains (so far only bitcoin is supported)
	k.SetChain(ctx, btc.Bitcoin)
	k.RegisterAsset(ctx, btc.Bitcoin.Name, btc.Bitcoin.NativeAsset)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	return types.GenesisState{Params: params}
}

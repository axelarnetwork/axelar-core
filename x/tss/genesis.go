package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, g types.GenesisState) {
	k.SetParams(ctx, g.Params)
	k.SetGovernanceKey(ctx, g.GovernanceKey)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	governanceKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		panic("unable to fetch governance key")
	}

	return types.GenesisState{
		Params:        params,
		GovernanceKey: governanceKey,
	}
}

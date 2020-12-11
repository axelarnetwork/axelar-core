package vote

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k keeper.Keeper, state types.GenesisState) {
	k.SetVotingInterval(ctx, state.VotingInterval)
	k.SetVotingThreshold(ctx, state.VotingThreshold)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	state := types.GenesisState{
		VotingInterval:  k.GetVotingInterval(ctx),
		VotingThreshold: k.GetVotingThreshold(ctx),
	}

	return state
}

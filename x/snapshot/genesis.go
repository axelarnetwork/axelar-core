package snapshot

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(_ sdk.Context, _ keeper.Keeper, _ types.GenesisState) {
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(_ sdk.Context, _ keeper.Keeper) types.GenesisState {
	return types.GenesisState{}
}

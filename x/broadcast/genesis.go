package broadcast

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/keeper"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(_ sdk.Context, _ keeper.Keeper, _ types.GenesisState) {
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(_ sdk.Context, _ keeper.Keeper) types.GenesisState {
	return types.GenesisState{}
}

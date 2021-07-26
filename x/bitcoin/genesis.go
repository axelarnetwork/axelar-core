package bitcoin

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// InitGenesis initialize default parameters
// from the genesis state
func InitGenesis(ctx sdk.Context, k types.BTCKeeper, g types.GenesisState) {
	k.SetParams(ctx, g.Params)

	// expose some parameters from genesis of btc module
	minAmount := int64(k.GetMinOutputAmount(ctx))
	telemetry.SetGauge(float32(minAmount), "btc", "min_withdrawal_mount")
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k types.BTCKeeper) *types.GenesisState {
	params := k.GetParams(ctx)
	return &types.GenesisState{Params: params}
}

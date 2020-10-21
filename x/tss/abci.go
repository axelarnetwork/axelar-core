package tss

import (
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(_ sdk.Context, _ abci.RequestEndBlock, _ keeper.Keeper) []abci.ValidatorUpdate {
	return nil
}

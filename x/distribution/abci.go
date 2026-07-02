/*
This file is identical to the Cosmos SDK distribution ABCI file (cosmos-sdk/x/distribution/abci.go).
It is duplicated here because the BeginBlocker function accepts a keeper struct as parameter.

Since we have our own keeper overrides the token allocation, we need to copy this file and modify the keeper type
to use our custom keeper instead of the SDK's keeper.
*/

package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/distribution/keeper"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	if err := k.BeginBlocker(ctx); err != nil {
		ctx.Logger().Error("BeginBlocker failed", "error", err)
		return err
	}

	return nil
}

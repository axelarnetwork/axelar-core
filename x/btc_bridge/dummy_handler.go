package btc_bridge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

// For testing purposes only
func NewDummyHandler(_ keeper.Keeper, _ types.Voter) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		return nil, fmt.Errorf("node has no Bitcoin bridge, aborting")
	}
}

package keeper

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// for commands approval
const gasCost = storetypes.Gas(10000000)

// NewMessageRoute creates a new message route
func NewMessageRoute() nexus.MessageRoute {
	return func(ctx sdk.Context, _ nexus.RoutingContext, _ nexus.GeneralMessage) error {
		ctx.GasMeter().ConsumeGas(gasCost, "execute-message")

		return nil
	}
}

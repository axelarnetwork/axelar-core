package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	types "github.com/axelarnetwork/axelar-core/x/nexus/types"
)

type req struct {
	RouteMessages []exported.WasmMessage `json:"route_messages"`
}

// NewMessageRoute creates a new message route
func NewMessageRoute(nexus types.Nexus, account types.AccountKeeper, wasm types.WasmKeeper) exported.MessageRoute {
	return func(ctx sdk.Context, _ exported.RoutingContext, msg exported.GeneralMessage) error {
		if msg.Asset != nil {
			return fmt.Errorf("asset is not supported")
		}

		gateway := nexus.GetParams(ctx).Gateway
		if gateway.Empty() {
			return fmt.Errorf("gateway is not set")
		}

		bz, err := json.Marshal(req{RouteMessages: []exported.WasmMessage{exported.FromGeneralMessage(msg)}})
		if err != nil {
			return nil
		}

		if _, err := wasm.Execute(ctx, gateway, account.GetModuleAddress(types.ModuleName), bz, sdk.NewCoins()); err != nil {
			return err
		}

		return nil
	}
}

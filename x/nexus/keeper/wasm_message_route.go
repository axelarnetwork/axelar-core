package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	types "github.com/axelarnetwork/axelar-core/x/nexus/types"
)

type request struct {
	RouteMessagesFromNexus []exported.WasmMessage `json:"route_messages_from_nexus"`
}

// NewMessageRoute creates a new message route
func NewMessageRoute(nexus types.Nexus, ibc types.IBCKeeper, bank types.BankKeeper, account types.AccountKeeper, wasm types.WasmKeeper) exported.MessageRoute {
	return func(ctx sdk.Context, _ exported.RoutingContext, msg exported.GeneralMessage) error {
		if !nexus.IsWasmConnectionActivated(ctx) {
			return fmt.Errorf("wasm connection is not activated")
		}

		gateway := nexus.GetParams(ctx).Gateway
		if gateway.Empty() {
			return fmt.Errorf("gateway is not set")
		}

		wasmMsg := exported.FromGeneralMessage(msg)
		if err := wasmMsg.ValidateBasic(); err != nil {
			return err
		}

		coins, err := unlockCoinIfAny(ctx, nexus, ibc, bank, account, wasmMsg)
		if err != nil {
			return err
		}

		bz, err := json.Marshal(request{RouteMessagesFromNexus: []exported.WasmMessage{wasmMsg}})
		if err != nil {
			return nil
		}

		if _, err := wasm.Execute(ctx, gateway, account.GetModuleAddress(types.ModuleName), bz, coins); err != nil {
			return err
		}

		if err := nexus.SetMessageExecuted(ctx, msg.ID); err != nil {
			return err
		}

		return nil
	}
}

func unlockCoinIfAny(ctx sdk.Context, nexus types.Nexus, ibc types.IBCKeeper, bank types.BankKeeper, account types.AccountKeeper, msg exported.WasmMessage) (sdk.Coins, error) {
	if msg.Asset == nil {
		return sdk.NewCoins(), nil
	}

	lockableAsset, err := nexus.NewLockableAsset(ctx, ibc, bank, *msg.Asset)
	if err != nil {
		return sdk.NewCoins(), err
	}

	return sdk.NewCoins(*msg.Asset), lockableAsset.UnlockTo(ctx, account.GetModuleAddress(types.ModuleName))
}

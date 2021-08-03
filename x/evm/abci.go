package evm

import (
	"fmt"
	types "github.com/axelarnetwork/axelar-core/x/evm/types"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, enrich TSS signing result events
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k keeper.Keeper) []abci.ValidatorUpdate {
	//if req.Height%k.GetSigCheckInterval(ctx) != 0 {
	if req.Height%1 != 0 {
		return nil
	}

	for _, event := range ctx.EventManager().ABCIEvents() {
		if event.Type == tsstypes.EventTypeSign {
			sigID, result, ok := getEventSigIDAndAction(event)
			if !ok {
				panic(fmt.Errorf("%s event does not contain expected attributes %s and %s", tsstypes.EventTypeSign, tsstypes.AttributeKeySigID, sdk.AttributeKeyAction))
			}

			// todo: get pending sign command should store chain
			if chain, selector, ok := k.GetPendingSignCommand(ctx, sigID); ok {
				k.DeletePendingSignCommand(ctx, sigID)

				// Notify that signature result can be queried
				ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeSignedCommandID,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyCommandID, sigID),
					sdk.NewAttribute(types.AttributeKeyCommandSelector, selector),
					sdk.NewAttribute(types.AttributeKeyChain, chain),
					sdk.NewAttribute(tsstypes.AttributeKeyResult, result),
				))

				k.Logger(ctx).Info(fmt.Sprintf("%s evm gateway %s command %s is fully signed", chain, selector, sigID))
			} else {
				continue
			}
		}
	}

	return nil
}

func tssSigPollMetaStringFromCommandID(commandID string) string {
	return tsstypes.ModuleName + "_" + commandID
}

func getEventSigIDAndAction(event abci.Event) (sigID string, result string, ok bool) {
	done := func() bool {
		return len(sigID) > 0 && len(result) > 0
	}

	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case tsstypes.AttributeKeySigID:
			sigID = string(attr.Value)
			if done() {
				ok = true
				return
			}
		case sdk.AttributeKeyAction:
			result = string(attr.Value)
			if done() {
				ok = true
				return
			}
		default:
		}
	}

	return
}

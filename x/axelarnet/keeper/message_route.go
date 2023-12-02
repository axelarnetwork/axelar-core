package keeper

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// for IBC execution
const gasCost = storetypes.Gas(1000000)

// NewMessageRoute creates a new message route
func NewMessageRoute(
	keeper Keeper,
	ibcK types.IBCKeeper,
	feegrantK types.FeegrantKeeper,
	bankK types.BankKeeper,
	nexusK types.Nexus,
	accountK types.AccountKeeper,
) nexus.MessageRoute {
	return func(ctx sdk.Context, routingCtx nexus.RoutingContext, msg nexus.GeneralMessage) error {
		if routingCtx.Payload == nil {
			return fmt.Errorf("payload is required for routing messages to a cosmos chain")
		}

		bz, err := types.TranslateMessage(msg, routingCtx.Payload)
		if err != nil {
			return sdkerrors.Wrap(err, "invalid payload")
		}

		asset, err := escrowAssetToMessageSender(ctx, keeper, feegrantK, bankK, nexusK, accountK, routingCtx, msg)
		if err != nil {
			return err
		}

		ctx.GasMeter().ConsumeGas(gasCost, "execute-message")

		return ibcK.SendMessage(sdk.WrapSDKContext(ctx), msg.Recipient, asset, string(bz), msg.ID)
	}
}

// all general messages are sent from the Axelar general message sender, so receiver can use the packet sender to authenticate the message
// escrowAssetToMessageSender sends the asset to general msg sender account
func escrowAssetToMessageSender(
	ctx sdk.Context,
	keeper Keeper,
	feegrantK types.FeegrantKeeper,
	bankK types.BankKeeper,
	nexusK types.Nexus,
	accountK types.AccountKeeper,
	routingCtx nexus.RoutingContext,
	msg nexus.GeneralMessage,
) (sdk.Coin, error) {
	switch msg.Type() {
	case nexus.TypeGeneralMessage:
		// pure general message, take dust amount from sender to satisfy ibc transfer requirements
		asset := sdk.NewCoin(exported.NativeAsset, sdk.OneInt())
		sender := routingCtx.Sender

		if !routingCtx.FeeGranter.Empty() {
			req := types.RouteMessageRequest{
				Sender:     routingCtx.Sender,
				ID:         msg.ID,
				Payload:    routingCtx.Payload,
				Feegranter: routingCtx.FeeGranter,
			}
			if err := feegrantK.UseGrantedFees(ctx, routingCtx.FeeGranter, routingCtx.Sender, sdk.NewCoins(asset), []sdk.Msg{&req}); err != nil {
				return sdk.Coin{}, err
			}

			sender = routingCtx.FeeGranter
		}

		return asset, bankK.SendCoins(ctx, sender, types.AxelarGMPAccount, sdk.NewCoins(asset))
	case nexus.TypeGeneralMessageWithToken:
		// general message with token, get token from corresponding account
		asset, sender, err := prepareTransfer(ctx, keeper, nexusK, bankK, accountK, *msg.Asset)
		if err != nil {
			return sdk.Coin{}, err
		}

		return asset, bankK.SendCoins(ctx, sender, types.AxelarGMPAccount, sdk.NewCoins(asset))
	default:
		return sdk.Coin{}, fmt.Errorf("unrecognized message type")
	}
}

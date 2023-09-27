package keeper

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ wasmkeeper.Messenger = (*Messenger)(nil)

type request = []exported.ConnectionRouterMessage

type Messenger struct {
	types.Nexus
}

// NewMessenger returns a new Messenger
func NewMessenger(nexus types.Nexus) Messenger {
	return Messenger{nexus}
}

// DispatchMsg decodes the messages from the cosmowasm connection router and routes them to the nexus module if possible
func (m Messenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	req := request{}
	if err := json.Unmarshal(msg.Custom, &req); err != nil {
		return nil, nil, sdkerrors.Wrap(wasmtypes.ErrUnknownMsg, err.Error())
	}

	connectionRouter := m.GetParams(ctx).ConnectionRouter

	if len(connectionRouter) == 0 {
		return nil, nil, fmt.Errorf("connection router is not set")
	}

	if !connectionRouter.Equals(contractAddr) {
		return nil, nil, fmt.Errorf("contract address %s is not the connection router", contractAddr)
	}

	for _, msg := range req {
		routed := utils.RunCached(ctx, m, func(ctx sdk.Context) (bool, error) {
			return m.routeMsg(ctx, msg)
		})

		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.ConnectionRouterMessageReceived{Message: msg, Routed: routed}))
	}

	return nil, nil, nil
}

func (m Messenger) routeMsg(ctx sdk.Context, msg exported.ConnectionRouterMessage) (bool, error) {
	recipientChain, ok := m.GetChain(ctx, msg.RecipientChain)
	if !ok {
		return false, fmt.Errorf("recipient chain %s is not a registered chain", msg.RecipientChain)
	}

	id, _, _ := m.GenerateMessageID(ctx)
	senderChain := exported.Chain{Name: msg.SenderChain, SupportsForeignAssets: false, KeyType: tss.None, Module: wasmtypes.ModuleName}
	sender := exported.CrossChainAddress{Chain: senderChain, Address: msg.SenderAddress}
	recipient := exported.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddress}

	// set status to approved if the message is sent to a cosmos chain and set
	// to processing otherwise, because messages sent to cosmos chains require
	// translation with the original payload.
	// https://github.com/axelarnetwork/axelar-core/blob/ea48d5b974dfd94ea235311eddabe23bfa430cd9/x/axelarnet/keeper/msg_server.go#L520
	status := exported.Approved
	if !recipientChain.IsFrom(axelarnet.ModuleName) {
		status = exported.Processing
	}

	if err := m.Nexus.SetNewMessageFromWasm(ctx, exported.NewGeneralMessage(
		id,
		sender,
		recipient,
		msg.PayloadHash,
		status,
		msg.SourceTxID,
		msg.SourceTxIndex,
		nil,
	)); err != nil {
		return false, err
	}

	return true, nil
}

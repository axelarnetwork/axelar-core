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
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ wasmkeeper.Messenger = (*Messenger)(nil)

type Messenger struct {
	types.Nexus
}

// NewMessenger returns a new Messenger
func NewMessenger(nexus types.Nexus) Messenger {
	return Messenger{nexus}
}

// DispatchMsg decodes the messages from the cosmowasm gateway and routes them to the nexus module if possible
func (m Messenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	req := exported.WasmMessage{}
	if err := json.Unmarshal(msg.Custom, &req); err != nil {
		return nil, nil, sdkerrors.Wrap(wasmtypes.ErrUnknownMsg, err.Error())
	}

	gateway := m.GetParams(ctx).Gateway

	if gateway.Empty() {
		return nil, nil, fmt.Errorf("gateway is not set")
	}

	if !gateway.Equals(contractAddr) {
		return nil, nil, fmt.Errorf("contract address %s is not the gateway", contractAddr)
	}

	if err := m.routeMsg(ctx, req); err != nil {
		return nil, nil, err
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.WasmMessageRouted{Message: req}))

	return nil, nil, nil
}

func (m Messenger) routeMsg(ctx sdk.Context, msg exported.WasmMessage) error {
	destinationChain, ok := m.GetChain(ctx, msg.DestinationChain)
	if !ok {
		return fmt.Errorf("recipient chain %s is not a registered chain", msg.DestinationChain)
	}

	id, _, _ := m.GenerateMessageID(ctx)
	sourceChain := exported.Chain{Name: msg.SourceChain, SupportsForeignAssets: false, KeyType: tss.None, Module: wasmtypes.ModuleName}
	sender := exported.CrossChainAddress{Chain: sourceChain, Address: msg.SourceAddress}
	recipient := exported.CrossChainAddress{Chain: destinationChain, Address: msg.DestinationAddress}

	nexusMsg := exported.NewGeneralMessage(id, sender, recipient, msg.PayloadHash, msg.SourceTxID, msg.SourceTxIndex, nil)
	if err := m.Nexus.SetNewMessage(ctx, nexusMsg); err != nil {
		return err
	}

	// try routing the message
	_ = utils.RunCached(ctx, m, func(ctx sdk.Context) (struct{}, error) {
		return struct{}{}, m.RouteMessage(ctx, nexusMsg.ID)
	})

	return nil
}

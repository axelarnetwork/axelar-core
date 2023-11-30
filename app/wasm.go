package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/axelarnetwork/axelar-core/x/ante"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AnteHandlerMessenger struct {
	anteHandle ante.MessageAnteHandler
	encoders   wasm.MessageEncoders
	messenger  wasmkeeper.Messenger
}

func (m AnteHandlerMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	sdkMsgs, err := m.encoders.Encode(ctx, contractAddr, contractIBCPortID, msg)
	if err != nil {
		return nil, nil, err
	}

	// we can't know if this is a simulation or not at this stage, so we treat it as a regular execution
	ctx, err = m.anteHandle(ctx, sdkMsgs, false)
	if err != nil {
		return nil, nil, err
	}

	return m.messenger.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func withAnteHandlers(encoders wasmkeeper.MessageEncoders, anteHandler ante.MessageAnteHandler, messenger wasmkeeper.Messenger) wasmkeeper.Messenger {
	return AnteHandlerMessenger{
		encoders:   encoders,
		anteHandle: anteHandler,
		messenger:  messenger,
	}
}

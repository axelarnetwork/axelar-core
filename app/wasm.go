package app

import (
	"encoding/json"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/x/ante"
)

type AnteHandlerMessenger struct {
	anteHandle ante.MessageAnteHandler
	encoders   wasm.MessageEncoders
	messenger  wasmkeeper.Messenger
}

func (m AnteHandlerMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	if err := assertSingleMessageIsSet(msg); err != nil {
		return nil, nil, err
	}

	// burn and ibc send packet cannot be converted into sdk.Msg and are irrelevant for ante handler checks
	if !isBankBurnMsg(msg) && !isIBCSendPacketMsg(msg) {
		sdkMsgs, err := m.encoders.Encode(ctx, contractAddr, contractIBCPortID, msg)
		if err != nil {
			return nil, nil, err
		}

		// we can't know if this is a simulation or not at this stage, so we treat it as a regular execution
		ctx, err = m.anteHandle(ctx, sdkMsgs, false)
		if err != nil {
			return nil, nil, err
		}
	}

	return m.messenger.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

func assertSingleMessageIsSet(msg wasmvmtypes.CosmosMsg) error {
	bz, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var msgs map[string]interface{}
	err = json.Unmarshal(bz, &msgs)
	if err != nil {
		return err
	}

	msgCount := 0
	for _, typedMsgs := range msgs {
		if typedMsgs, ok := typedMsgs.(map[string]interface{}); ok {
			msgCount += len(maps.Keys(typedMsgs))
		}
	}

	if msgCount == 0 {
		return fmt.Errorf("no message set")
	} else if msgCount > 1 {
		return fmt.Errorf("only one message can be set, got %d", msgCount)
	} else {
		return nil
	}
}

func isBankBurnMsg(msg wasmvmtypes.CosmosMsg) bool {
	return msg.Bank != nil && msg.Bank.Burn != nil
}

func isIBCSendPacketMsg(msg wasmvmtypes.CosmosMsg) bool {
	return msg.IBC != nil && msg.IBC.SendPacket != nil
}

func WithAnteHandlers(encoders wasmkeeper.MessageEncoders, anteHandler ante.MessageAnteHandler, messenger wasmkeeper.Messenger) wasmkeeper.Messenger {
	return AnteHandlerMessenger{
		encoders:   encoders,
		anteHandle: anteHandler,
		messenger:  messenger,
	}
}

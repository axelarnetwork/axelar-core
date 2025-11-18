package app

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/x/ante"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

//go:generate moq -pkg mock -out ./mock/ibchooks.go . PacketI

type PacketI ibcexported.PacketI

type AnteHandlerMessenger struct {
	anteHandle ante.MessageAnteHandler
	encoders   wasm.MessageEncoders
	messenger  wasmkeeper.Messenger
}

func (m AnteHandlerMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error) {
	if err := assertSingleMessageIsSet(msg); err != nil {
		return nil, nil, nil, err
	}

	// burn and ibc send packet cannot be converted into sdk.Msg and are irrelevant for ante handler checks
	if !isBankBurnMsg(msg) && !isIBCSendPacketMsg(msg) {
		sdkMsgs, err := m.encoders.Encode(ctx, contractAddr, contractIBCPortID, msg)
		if err != nil {
			return nil, nil, nil, err
		}

		// we can't know if this is a simulation or not at this stage, so we treat it as a regular execution
		ctx, err = m.anteHandle(ctx, sdkMsgs, false)
		if err != nil {
			return nil, nil, nil, err
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
	for msgType, typedMsgs := range msgs {
		// custom and any / stargate msgs are not categorized in CosmosMsg, so the next lower structural level would be message fields and not individual messages,
		// so we can safely assume that there is only one message
		if msgType == "custom" || msgType == "any" {
			msgCount++
		} else if typedMsgs, ok := typedMsgs.(map[string]interface{}); ok {
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

func WithAnteHandlers(encoders wasmkeeper.MessageEncoders, anteHandler ante.MessageAnteHandler, messenger wasmkeeper.Messenger) wasmkeeper.Messenger {
	return AnteHandlerMessenger{
		encoders:   encoders,
		anteHandle: anteHandler,
		messenger:  messenger,
	}
}

type MsgTypeBlacklistMessenger struct {
}

func NewMsgTypeBlacklistMessenger() MsgTypeBlacklistMessenger {
	return MsgTypeBlacklistMessenger{}
}

func (m MsgTypeBlacklistMessenger) DispatchMsg(_ sdk.Context, _ sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error) {
	if isIBCSendPacketMsg(msg) || isAnyMsg(msg) {
		return nil, nil, nil, fmt.Errorf("ibc send packet and any messages are not supported")
	}

	// this means that this message handler doesn't know how to deal with these messages (i.e. they can pass through),
	// other handlers might be able to deal with them
	return nil, nil, nil, wasmtypes.ErrUnknownMsg
}

func isBankBurnMsg(msg wasmvmtypes.CosmosMsg) bool {
	return msg.Bank != nil && msg.Bank.Burn != nil
}

func isAnyMsg(msg wasmvmtypes.CosmosMsg) bool {
	return msg.Any != nil
}

func isIBCSendPacketMsg(msg wasmvmtypes.CosmosMsg) bool {
	return msg.IBC != nil && msg.IBC.SendPacket != nil
}

type WasmAppModuleBasicOverride struct {
	wasm.AppModuleBasic
}

func NewWasmAppModuleBasicOverride(wasmModule wasm.AppModuleBasic) WasmAppModuleBasicOverride {
	// Both the server and the cosmwasm client use this parameter to validate MsgStoreCode.
	// Because the AppModuleBasic provides server and client commands, it's sufficient to do the override here to set it for both.
	if MaxWasmSize != "" {
		// Override the default max wasm code size
		wasmtypes.MaxWasmSize = funcs.Must(strconv.Atoi(MaxWasmSize))
	}

	return WasmAppModuleBasicOverride{
		AppModuleBasic: wasmModule,
	}
}

// DefaultGenesis returns an override for the wasm module's DefaultGenesis,
// because as soon as the module is initialized the restriction to contract upload and instantiation must hold
func (m WasmAppModuleBasicOverride) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&wasm.GenesisState{
		Params: wasmtypes.Params{
			CodeUploadAccess:             wasmtypes.AllowNobody,
			InstantiateDefaultPermission: wasmtypes.AccessTypeNobody,
		},
	})
}

// QueryRequest is the custom queries wasm contracts can make for the core modules
type QueryRequest struct {
	Nexus *nexus.WasmQueryRequest
}

// NewQueryPlugins returns a new instance of the custom query plugins
func NewQueryPlugins(nexus nexustypes.Nexus) *wasmkeeper.QueryPlugins {
	nexusWasmQuerier := nexusKeeper.NewWasmQuerier(nexus)

	return &wasmkeeper.QueryPlugins{
		Custom: func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
			req := QueryRequest{}
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, wasmvmtypes.InvalidRequest{Err: "invalid Custom query request", Request: request}
			}

			if req.Nexus != nil {
				return nexusWasmQuerier.Query(ctx, *req.Nexus)
			}

			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Custom query request"}
		},
	}
}

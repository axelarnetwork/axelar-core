package types

import (
	"encoding/json"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType      = funcs.Must(abi.NewType("string", "string", nil))
	stringArrayType = funcs.Must(abi.NewType("string[]", "string[]", nil))
	bytesType       = funcs.Must(abi.NewType("bytes", "bytes", nil))
	bytes32Type     = funcs.Must(abi.NewType("bytes32", "bytes32", nil))

	// payloadWithVersion is a payload with message version number
	// - bytes32(0) To Native
	// - bytes32(1) To Cosmwasm Contract
	payloadWithVersion = abi.Arguments{{Type: bytes32Type}, {Type: bytesType}}

	// abi encoded bytes, with the following format:
	// wasm method name, argument name list, argument type list, argument value list
	payloadArguments = abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}
)

type contractCall struct {
	SourceChain string `json:"source_chain"`
	// The sender address on the source chain
	Sender string `json:"sender"`
	// Contract is the address of the wasm contract
	Contract string `json:"contract"`
	// Msg is a json struct {"methodName": {"arg1": "val1", "arg2": "val2"}}
	Msg map[string]interface{} `json:"msg"`
}

// wasm is the json that gets passed to the IBC memo field
type wasm struct {
	Wasm contractCall `json:"wasm"`
}

type message struct {
	SourceChain string `json:"source_chain"`
	Sender      string `json:"sender"`
	Payload     []byte `json:"payload"`
	Type        int    `json:"type"`
}

// TranslateMessage constructs the message gets passed to a cosmos chain from abi encoded payload
func TranslateMessage(msg nexus.GeneralMessage, payload []byte) ([]byte, error) {
	version, payload, err := unpackPayload(payload)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid payload with version")
	}

	return constructMessage(msg, version, payload)
}

// unpackPayload returns the version and actual payload
func unpackPayload(payload []byte) ([32]byte, []byte, error) {
	params, err := evm.StrictDecode(payloadWithVersion, payload)
	if err != nil {
		return [32]byte{}, nil, sdkerrors.Wrap(err, "failed to unpack payload")
	}

	return params[0].([32]byte), params[1].([]byte), nil
}

// constructMessage constructs message based on the payload version
func constructMessage(gm nexus.GeneralMessage, version [32]byte, payload []byte) ([]byte, error) {
	var bz []byte
	var err error

	switch hexutil.Encode(version[:]) {
	case NativeV1:
		bz, err = ConstructNativeMessage(gm, payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to construct native payload")
		}
	case CosmwasmV1:
		bz, err = ConstructWasmMessage(gm, payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to construct wasm payload")
		}
	default:
		return nil, fmt.Errorf("unknown payload version")
	}

	return bz, nil
}

// ConstructWasmMessage creates a json serialized wasm message from Axelar defined abi encoded payload
// The abi encoded payload must contain the following information in order
// - method name (string)
// - argument names ([]string)
// - argument types ([]string)
// - argument values (bytes)
func ConstructWasmMessage(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
	args, err := evm.StrictDecode(payloadArguments, payload)
	if err != nil {
		return nil, err
	}

	methodName := args[0].(string)
	argNames := args[1].([]string)
	argTypes := args[2].([]string)

	if len(argNames) != len(argTypes) {
		return nil, fmt.Errorf("payload argument name and type length mismatch")
	}

	abiArguments, err := buildArguments(argTypes)
	if err != nil {
		return nil, err
	}

	// unpack to actual argument values
	argValues, err := evm.StrictDecode(abiArguments, args[3].([]byte))
	if err != nil {
		return nil, err
	}

	// convert to execute msg payload
	executeMsg := make(map[string]interface{})
	for i := 0; i < len(argNames); i++ {
		executeMsg[argNames[i]] = argValues[i]
	}

	msg := wasm{
		Wasm: contractCall{
			Contract:    gm.GetDestinationAddress(),
			SourceChain: gm.GetSourceChain().String(),
			Sender:      gm.GetSourceAddress(),
			Msg: map[string]interface{}{
				methodName: executeMsg,
			},
		},
	}

	return json.Marshal(msg)
}

// ConstructNativeMessage creates a json serialized cross chain message
func ConstructNativeMessage(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
	return json.Marshal(message{
		SourceChain: gm.GetSourceChain().String(),
		Sender:      gm.GetSourceAddress(),
		Payload:     payload,
		Type:        int(gm.Type()),
	})
}

// build abi arguments based on argument types to decode the actual wasm contract arguments
func buildArguments(argTypes []string) (abi.Arguments, error) {
	var arguments abi.Arguments
	for _, typeStr := range argTypes {
		argType, err := abi.NewType(typeStr, typeStr, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid argument type %s", typeStr)
		}

		arguments = append(arguments, abi.Argument{Type: argType})
	}

	return arguments, nil
}

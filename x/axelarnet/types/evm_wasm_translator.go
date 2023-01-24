package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType      = funcs.Must(abi.NewType("string", "string", nil))
	stringArrayType = funcs.Must(abi.NewType("string[]", "string[]", nil))
	bytesType       = funcs.Must(abi.NewType("bytes", "bytes", nil))

	// abi encoded bytes, with the following format:
	// wasm method name, argument name list, argument type list, argument value list
	payloadArguments = abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}
)

type contractCall struct {
	SourceChain string `json:"source_chain"`
	Sender      string `json:"sender"`
	// Contract is the address of the wasm contract
	Contract string `json:"contract"`
	// Msg is a json struct {"methodName": {"arg1": "val1", "arg2": "val2"}}
	Msg map[string]interface{} `json:"msg"`
}

// wasm is the json that gets passed to the IBC memo field
type wasm struct {
	Wasm contractCall `json:"wasm"`
}

// ConstructWasmMessage creates a json serialized wasm message from Axelar defined abi encoded payload
// The abi encoded payload must contain the following information in order
// - method name (string)
// - argument names ([]string)
// - argument types ([]string)
// - argument values (bytes)
func ConstructWasmMessage(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
	args, err := payloadArguments.Unpack(payload)
	if err != nil {
		return nil, err
	}

	methodName := args[0].(string)
	argNames := args[1].([]string)
	argTypes := args[2].([]string)

	abiArguments, err := buildArguments(argTypes)
	if err != nil {
		return nil, err
	}

	// unpack to actual argument values
	argValues, err := abiArguments.Unpack(args[3].([]byte))
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
			Contract:    gm.Receiver,
			SourceChain: gm.SourceChain.String(),
			Sender:      gm.Sender,
			Msg: map[string]interface{}{
				methodName: executeMsg,
			},
		},
	}

	return json.Marshal(msg)
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

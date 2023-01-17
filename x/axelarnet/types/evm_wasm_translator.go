package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType      = funcs.Must(abi.NewType("string", "string", nil))
	stringArrayType = funcs.Must(abi.NewType("string[]", "string[]", nil))
	bytesType       = funcs.Must(abi.NewType("bytes", "bytes", nil))

	// abi encoded bytes, with the following format:
	// wasm method name, argument type list, encoded bytes contain actual argument names and values
	payloadArguments = abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: bytesType}}
)

type contract struct {
	// Contract is the address of the wasm contract
	Contract string `json:"contract"`
	// Msg is a json struct {"methodName": {"arg1": "val1", "arg2": "val2"}}
	Msg map[string]interface{} `json:"msg"`
}

// wasm is the json get passed to the IBC memo field
type wasm struct {
	Wasm contract `json:"wasm"`
}

// ConstructWasmMessage creates a json serialized wasm message from Axelar defined abi encoded payload
func ConstructWasmMessage(contractAddr string, payload []byte) ([]byte, error) {
	args, err := payloadArguments.Unpack(payload)
	if err != nil {
		return nil, err
	}

	abiArguments, err := buildArguments(args[1].([]string))
	if err != nil {
		return nil, err
	}

	// unpack argument bytes to list of arg name and arg value
	arguments, err := abiArguments.Unpack(args[2].([]byte))
	if err != nil {
		return nil, err
	}

	// convert to execute msg payload
	executeMsg := make(map[string]interface{})
	for idx := 0; idx < len(arguments); idx += 2 {
		argName := arguments[idx].(string)
		argValue := arguments[idx+1]

		executeMsg[argName] = argValue
	}

	methodName := args[0].(string)

	msg := wasm{
		Wasm: contract{
			Contract: contractAddr,
			Msg: map[string]interface{}{
				methodName: executeMsg,
			},
		},
	}

	return json.Marshal(msg)
}

// build abi arguments based on argument types to decode the payload
func buildArguments(argTypes []string) (abi.Arguments, error) {
	var arguments abi.Arguments
	for _, typeStr := range argTypes {
		argType, err := abi.NewType(typeStr, typeStr, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid argument type %s", typeStr)
		}

		// arguments hold argument name and type pairs
		// the first argument is always string, and the second argument is the actual type
		arguments = append(arguments, abi.Argument{Type: stringType}, abi.Argument{Type: argType})
	}

	return arguments, nil
}

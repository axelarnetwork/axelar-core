package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType  = funcs.Must(abi.NewType("string", "string", nil))
	uint256Type = funcs.Must(abi.NewType("uint256", "uint256", nil))
	bytesType   = funcs.Must(abi.NewType("bytes", "bytes", nil))

	// abi encoded bytes, with the following format:
	// wasm method name, argument number, comma seperated argument types, encoded bytes contain actual arguments name and value
	payloadArguments = abi.Arguments{{Type: stringType}, {Type: uint256Type}, {Type: stringType}, {Type: bytesType}}
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

// Interpreter translates an abi encoded payload to json serialized wasm message
type Interpreter struct {
	methodName string
	argNum     uint64
	argTypes   []string
	argBytes   []byte
}

// NewInterpreter returns an evm interpreter
func NewInterpreter(payload []byte) (Interpreter, error) {
	args, err := payloadArguments.Unpack(payload)
	if err != nil {
		return Interpreter{}, err
	}

	argNum := sdk.NewUintFromBigInt(args[1].(*big.Int)).Uint64()
	typeStrLst := strings.Split(args[2].(string), ",")

	if uint64(len(typeStrLst)) != argNum {
		return Interpreter{}, fmt.Errorf("argument number %d does not match argument type length %d", args[1], len(typeStrLst))
	}

	return Interpreter{
		methodName: args[0].(string),
		argNum:     argNum,
		argTypes:   typeStrLst,
		argBytes:   args[3].([]byte),
	}, err
}

// ToWasmMsg converts the abi encoded payload to json serialized wasm message
func (i Interpreter) ToWasmMsg(contractAddr string) ([]byte, error) {
	// a list of arg name and value e.g. [arg_name1, arg_val1, arg_name2, arg_val2]
	arguments, err := i.parseArguments()
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

	msg := wasm{
		Wasm: contract{
			Contract: contractAddr,
			Msg: map[string]interface{}{
				i.methodName: executeMsg,
			},
		},
	}

	return json.Marshal(msg)
}

// parseArguments returns a list of Go format arguments to construct wasm message
func (i Interpreter) parseArguments() ([]interface{}, error) {
	var payloadArgs abi.Arguments
	for _, typeStr := range i.argTypes {
		argType, err := abi.NewType(typeStr, typeStr, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid argument type %s", typeStr)
		}

		// payloadArgs hold argument name and type pairs.
		payloadArgs = append(payloadArgs, abi.Argument{Type: stringType}, abi.Argument{Type: argType})
	}

	// list of arg name and value e.g. [arg_name1, arg_val1, arg_name2, arg_val2]
	return payloadArgs.Unpack(i.argBytes)
}

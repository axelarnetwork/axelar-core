package types

import (
	"encoding/json"
	"fmt"
	"strings"

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

	// abi encoded bytes, with the following format:
	// wasm method name, argument name list, argument type list, argument value list
	payloadArguments = abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}
)

const (
	// versionSize is the size of the version in the payload
	// - bytes4(0) To Native
	// - bytes4(1) To CosmWasm Contract
	// - bytes4(2) To CosmWasm Contract with json encoded payload
	versionSize = 4

	sourceChain   = "source_chain"
	sourceAddress = "source_address"
)

type version [versionSize]byte

type contractCall struct {
	SourceChain string `json:"source_chain"`
	// The sender address on the source chain
	SourceAddress string `json:"source_address"`
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
	SourceChain   string `json:"source_chain"`
	SourceAddress string `json:"source_address"`
	Payload       []byte `json:"payload"`
	Type          int64  `json:"type"`
}

// TranslateMessage constructs the message gets passed to a cosmos chain from a versioned payload
func TranslateMessage(msg nexus.GeneralMessage, versionedPayload []byte) ([]byte, error) {
	version, payload, err := unpackVersionedPayload(versionedPayload)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid versioned payload")
	}

	var bz []byte
	switch hexutil.Encode(version[:]) {
	case NativeV1:
		bz, err = ConstructNativeMessage(msg, payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to construct native payload")
		}
	case CosmWasmV1:
		bz, err = ConstructWasmMessageV1(msg, payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to construct wasm payload")
		}
	case CosmWasmV2:
		bz, err = ConstructWasmMessageV2(msg, payload)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to construct wasm payload")
		}
	default:
		return nil, fmt.Errorf("unknown payload version")
	}

	return bz, nil
}

// unpackVersionedPayload returns the version and actual payload
func unpackVersionedPayload(versionedPayload []byte) (version, []byte, error) {
	if len(versionedPayload) <= versionSize {
		return version{}, nil, fmt.Errorf("invalid versionedPayload length")
	}

	// first 4 bytes is the version
	var v version
	copy(v[:], versionedPayload[:versionSize])

	return v, versionedPayload[versionSize:], nil
}

// ConstructWasmMessageV1 creates a json serialized wasm message from Axelar defined abi encoded payload
// The abi encoded payload must contain the following information in order
// - method name (string)
// - argument names ([]string)
// - argument types ([]string)
// - argument values (bytes)
func ConstructWasmMessageV1(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
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

	err = checkSourceInfo(gm.Sender, executeMsg)
	if err != nil {
		return nil, err
	}

	msg := wasm{
		Wasm: contractCall{
			Contract:      gm.GetDestinationAddress(),
			SourceChain:   gm.GetSourceChain().String(),
			SourceAddress: gm.GetSourceAddress(),
			Msg: map[string]interface{}{
				methodName: executeMsg,
			},
		},
	}

	return json.Marshal(msg)
}

// ConstructWasmMessageV2 creates a json serialized wasm message from json encoded payload
// The payload must contain only a single key, the method name, and an arg name val map as value
func ConstructWasmMessageV2(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
	executeMsg := make(map[string]interface{})
	err := json.Unmarshal(payload, &executeMsg)
	if err != nil {
		return nil, err
	}

	// json payload must have single key, the method name
	if len(executeMsg) != 1 {
		return nil, fmt.Errorf("invalid payload")
	}

	// iterating over the map, as the key (method name) is dynamic
	for _, msg := range executeMsg {
		// value must be a map
		args, ok := msg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}

		err = checkSourceInfo(gm.Sender, args)
		if err != nil {
			return nil, err
		}
	}

	// When JSON unmarshalling the user payload to a map[string]interface{} type,
	// numbers will get converted to floats. When this is marshalled again, the floats aren't converted back,
	// leading to loss of precision, and potential non-determinism.
	// So we leave the payload blank before the marshalling the following,
	// and then inject the original payload into the json string instead.
	wasmMsg := wasm{
		Wasm: contractCall{
			Contract:      gm.GetDestinationAddress(),
			SourceChain:   gm.GetSourceChain().String(),
			SourceAddress: gm.GetSourceAddress(),
			Msg:           nil,
		},
	}

	msg, err := json.Marshal(wasmMsg)
	if err != nil {
		return nil, err
	}

	originalField := `"msg":null`
	replacementField := fmt.Sprintf("\"msg\":%s", string(payload))
	msg = []byte(strings.Replace(string(msg), originalField, replacementField, 1))

	return msg, err
}

// ConstructNativeMessage creates a json serialized cross chain message
func ConstructNativeMessage(gm nexus.GeneralMessage, payload []byte) ([]byte, error) {
	return json.Marshal(message{
		SourceChain:   gm.GetSourceChain().String(),
		SourceAddress: gm.GetSourceAddress(),
		Payload:       payload,
		Type:          int64(gm.Type()),
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

func checkSourceInfo(sender nexus.CrossChainAddress, msg map[string]interface{}) error {
	chain, ok := msg[sourceChain]
	if ok {
		chain, ok := chain.(string)
		if !ok {
			return fmt.Errorf("source chain must have type string")
		}

		if !sender.Chain.Name.Equals(nexus.ChainName(chain)) {
			return fmt.Errorf("source chain does not match expected")
		}
	}

	addr, ok := msg[sourceAddress]
	if ok {
		// Convert interface to string to support the scenario where addrI uses abi.Address type
		// Note: Avoid using common.HexToAddress without checking if it's a valid address first since it doesn't handle invalid inputs well.
		addr := fmt.Sprint(addr)
		if !strings.EqualFold(sender.Address, addr) {
			return fmt.Errorf("source address does not match expected")
		}
	}

	return nil
}

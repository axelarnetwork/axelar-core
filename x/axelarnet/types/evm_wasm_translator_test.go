package types_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType = funcs.Must(abi.NewType("string", "string", nil))

	bytes32Type = funcs.Must(abi.NewType("bytes32", "bytes32", nil))
	bytesType   = funcs.Must(abi.NewType("bytes", "bytes", nil))

	uint8Type   = funcs.Must(abi.NewType("uint8", "uint8", nil))
	uint128Type = funcs.Must(abi.NewType("uint128", "uint128", nil))
	uint256Type = funcs.Must(abi.NewType("uint256", "uint256", nil))

	stringArrayType  = funcs.Must(abi.NewType("string[]", "string[]", nil))
	uint256ArrayType = funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))
)

func TestNewInterpreter(t *testing.T) {
	// repeated arg name (string) and arg value (depends on actual type)
	methodArguments := abi.Arguments{
		{Type: stringType}, {Type: stringType},
		{Type: stringType}, {Type: uint256Type},
		{Type: stringType}, {Type: stringType},
		{Type: stringType}, {Type: stringType},
		{Type: stringType}, {Type: uint8Type},
		{Type: stringType}, {Type: stringType},
		{Type: stringType}, {Type: bytesType},
	}

	slippage := uint8(2)

	argumentBz, err := methodArguments.Pack(
		"input_denom", "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
		"input_amount", math.MaxBig256,
		"output_denom", "ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5",
		"destination", "chain-a",
		"max_price_impact_percentage", slippage,
		"receiver", rand.AccAddr().String(),
		"payload", rand.Bytes(int(rand.I64Between(1, 10000))),
	)
	assert.NoError(t, err)

	schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: bytesType}}

	payload, err := schema.Pack(
		"swap_and_forward",
		[]string{"string", "uint128", "string", "string", "uint8", "string", "bytes"},
		argumentBz,
	)
	assert.NoError(t, err)

	msg, err := types.ConstructWasmMessage(rand.AccAddr().String(), payload)
	assert.NoError(t, err)

	fmt.Println(string(msg))
}

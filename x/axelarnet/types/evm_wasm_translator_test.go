package types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	stringType      = funcs.Must(abi.NewType("string", "string", nil))
	bytesType       = funcs.Must(abi.NewType("bytes", "bytes", nil))
	uint8Type       = funcs.Must(abi.NewType("uint8", "uint8", nil))
	uint256Type     = funcs.Must(abi.NewType("uint256", "uint256", nil))
	stringArrayType = funcs.Must(abi.NewType("string[]", "string[]", nil))
)

func TestNewInterpreter(t *testing.T) {
	methodArguments := abi.Arguments{
		{Type: stringType},
		{Type: uint256Type},
		{Type: stringType},
		{Type: stringType},
		{Type: uint8Type},
		{Type: stringType},
		{Type: bytesType},
	}

	slippage := uint8(2)

	argumentTypes := []string{"string", "uint128", "string", "string", "uint8", "string", "bytes"}
	argumentNames := []string{"input_denom", "input_amount", "output_denom", "destination", "max_price_impact_percentage", "receiver", "payload"}
	argumentValues, err := methodArguments.Pack(
		"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
		math.MaxBig256,
		"ibc/EA1D43981D5C9A1C4AAEA9C23BB1D4FA126BA9BC7020A25E0AE4AA841EA25DC5",
		"chain-a",
		slippage,
		rand.AccAddr().String(),
		rand.Bytes(int(rand.I64Between(1, 10000))),
	)
	assert.NoError(t, err)

	schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

	payload, err := schema.Pack(
		"swap_and_forward",
		argumentNames,
		argumentTypes,
		argumentValues,
	)
	assert.NoError(t, err)

	_, err = types.ConstructWasmMessage(rand.AccAddr().String(), payload)
	assert.NoError(t, err)
}

package types_test

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var (
	boolType              = funcs.Must(abi.NewType("bool", "bool", nil))
	stringType            = funcs.Must(abi.NewType("string", "string", nil))
	bytesType             = funcs.Must(abi.NewType("bytes", "bytes", nil))
	bytes32Type           = funcs.Must(abi.NewType("bytes32", "bytes32", nil))
	uint8Type             = funcs.Must(abi.NewType("uint8", "uint8", nil))
	uint64Type            = funcs.Must(abi.NewType("uint64", "uint64", nil))
	uint64ArrayType       = funcs.Must(abi.NewType("uint64[]", "uint64[]", nil))
	uint64ArrayNestedType = funcs.Must(abi.NewType("uint64[][]", "uint64[][]", nil))
	stringArrayType       = funcs.Must(abi.NewType("string[]", "string[]", nil))
	stringArrayNestedType = funcs.Must(abi.NewType("string[][]", "string[][]", nil))
)

func TestTranslator(t *testing.T) {
	methodName := "swap_and_forward"
	str := "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"
	maxUint256Str := utils.MaxUint.String()
	maxUint8 := uint8(math.MaxUint8)
	maxUint64 := uint64(math.MaxUint64)
	boolTrue := true

	stringArray := slices.Expand2(rand.AccAddr().String, 10)
	uint64Array := slices.Expand2(func() uint64 { return math.MaxUint64 }, 10)
	uint64NestedArray := slices.Expand2(func() []uint64 { return uint64Array }, 5)
	stringNestedArray := slices.Expand2(func() []string { return stringArray }, 5)
	bz := rand.Bytes(int(rand.I64Between(1, 1000)))
	hexBzStr := hex.EncodeToString(bz)

	argumentNames := []string{"str", "max_uint256_str", "max_uint8", "max_uint64", "bool_true", "string_array", "uint64_array", "uint64_array_nested", "string_array_nested", "bytes", "hex_string"}
	argumentTypes := []string{"string", "string", "uint8", "uint64", "bool", "string[]", "uint64[]", "uint64[][]", "string[][]", "bytes", "string"}

	methodArguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: uint8Type},
		{Type: uint64Type},
		{Type: boolType},
		{Type: stringArrayType},
		{Type: uint64ArrayType},
		{Type: uint64ArrayNestedType},
		{Type: stringArrayNestedType},
		{Type: bytesType},
		{Type: stringType},
	}

	argumentValues, err := methodArguments.Pack(
		str,
		maxUint256Str,
		maxUint8,
		maxUint64,
		boolTrue,
		stringArray,
		uint64Array,
		uint64NestedArray,
		stringNestedArray,
		bz,
		hexBzStr,
	)
	assert.NoError(t, err)

	schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

	payload, err := schema.Pack(
		methodName,
		argumentNames,
		argumentTypes,
		argumentValues,
	)
	assert.NoError(t, err)

	gm := nexus.GeneralMessage{

		Sender: nexus.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: evmtestutils.RandomAddress().Hex(),
		},
		Recipient: nexus.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.AccAddr().String(),
		},
	}

	var v [32]byte
	copy(v[:], funcs.Must(hexutil.Decode(types.CosmwasmV1)))

	msg, err := types.ConstructWasmMessageV1(gm, payload)
	assert.NoError(t, err)

	jsonObject := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewBuffer(msg))
	decoder.UseNumber()
	err = decoder.Decode(&jsonObject)
	assert.NoError(t, err)

	wasmRaw, ok := jsonObject["wasm"]
	assert.True(t, ok)

	wasm, ok := wasmRaw.(map[string]interface{})
	assert.True(t, ok)

	actualSourceChain, ok := wasm["source_chain"].(string)
	assert.True(t, ok)
	assert.Equal(t, gm.GetSourceChain().String(), actualSourceChain)

	actualSender, ok := wasm["source_address"].(string)
	assert.True(t, ok)
	assert.Equal(t, gm.GetSourceAddress(), actualSender)

	actualContractAddr, ok := wasm["contract"].(string)
	assert.True(t, ok)
	assert.Equal(t, gm.GetDestinationAddress(), actualContractAddr)

	// make sure execute message can be serialized
	_, err = json.Marshal(wasmRaw)
	assert.NoError(t, err)

	wasmMsg, ok := wasm["msg"].(map[string]interface{})
	assert.True(t, ok)

	executeMsg, ok := wasmMsg[methodName].(map[string]interface{})
	assert.True(t, ok)

	actualStr, ok := executeMsg["str"].(string)
	assert.True(t, ok)
	assert.Equal(t, str, actualStr)

	actualMaxUint256Str, ok := executeMsg["max_uint256_str"].(string)
	assert.True(t, ok)
	assert.Equal(t, maxUint256Str, actualMaxUint256Str)

	jsonNumber, ok := executeMsg["max_uint8"].(json.Number)
	assert.True(t, ok)
	assert.Equal(t, maxUint8, uint8(funcs.Must(strconv.ParseUint(jsonNumber.String(), 10, 8))))

	jsonNumber, ok = executeMsg["max_uint64"].(json.Number)
	assert.True(t, ok)
	assert.Equal(t, maxUint64, funcs.Must(strconv.ParseUint(jsonNumber.String(), 10, 64)))

	actualBool, ok := executeMsg["bool_true"].(bool)
	assert.True(t, ok)
	assert.Equal(t, boolTrue, actualBool)

	arrayInterface, ok := executeMsg["string_array"].([]interface{})
	actualStringArray := slices.Map(arrayInterface, func(t interface{}) string { return t.(string) })
	assert.Equal(t, stringArray, actualStringArray)

	arrayInterface, ok = executeMsg["uint64_array"].([]interface{})
	uint64StrArray := slices.Map(arrayInterface, func(t interface{}) string { return t.(json.Number).String() })
	assert.Equal(t, uint64Array, slices.Map(uint64StrArray, func(t string) uint64 { return funcs.Must(strconv.ParseUint(t, 10, 64)) }))

	arrayInterface, ok = executeMsg["uint64_array_nested"].([]interface{})
	actualUint64NestedArray := slices.Map(arrayInterface, func(inner interface{}) []uint64 {
		return slices.Map(inner.([]interface{}), func(t interface{}) uint64 {
			return funcs.Must(strconv.ParseUint(t.(json.Number).String(), 10, 64))
		})
	})
	assert.Equal(t, uint64NestedArray, actualUint64NestedArray)

	arrayInterface, ok = executeMsg["string_array_nested"].([]interface{})
	actualStringNestedArray := slices.Map(arrayInterface, func(inner interface{}) []string {
		return slices.Map(inner.([]interface{}), func(t interface{}) string {
			return t.(string)
		})
	})
	assert.Equal(t, stringNestedArray, actualStringNestedArray)

	base64BzString, ok := executeMsg["bytes"].(string)
	assert.True(t, ok)
	assert.Equal(t, b64.StdEncoding.EncodeToString(bz), base64BzString)

	actualHexString, ok := executeMsg["hex_string"].(string)
	assert.True(t, ok)
	assert.Equal(t, hexBzStr, actualHexString)
}

func TestConstructWasmMessageV2(t *testing.T) {
	args := abi.Arguments{{Type: bytes32Type}, {Type: bytesType}}
	var ver [32]byte
	copy(ver[:], funcs.Must(hexutil.Decode(types.CosmwasmV2)))

	t.Run("should return error if invalid message payload", func(t *testing.T) {
		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": [3, 12, 143]},
				"contract_name2": {"source_chain": "ethereum", "source_address": [3, 12, 143]}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := funcs.Must(args.Pack(ver, wasmMsg))
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "invalid payload")
	})

	t.Run("should return error if invalid message payload", func(t *testing.T) {
		wasmMsg := []byte(`{"contract_name": [1,2,3,4,5]}`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := funcs.Must(args.Pack(ver, wasmMsg))
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "invalid arguments")
	})

	t.Run("should return error if incorrect payload contains incorrect source chain", func(t *testing.T) {
		wasmMsg := []byte(`{"contract_name": {"source_chain": 1.1, "source_address": [3, 12, 143]}}`)

		msg := nexustestutils.RandomMessage()
		payload := funcs.Must(args.Pack(ver, wasmMsg))
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source chain does not match expected")
	})

	t.Run("should return error if incorrect payload contains incorrect source address", func(t *testing.T) {
		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": [3, 12, 143]}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := funcs.Must(args.Pack(ver, wasmMsg))
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should return error if incorrect payload contains incorrect source address", func(t *testing.T) {

		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": [3, 12, 143]}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := funcs.Must(args.Pack(ver, wasmMsg))
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should construct wasm message", func(t *testing.T) {
		bz := []byte(`
			{
				"contract_name": {
					"source_chain": "Ethereum",
					"source_address": "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955",
					"nested": {
						"denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", 
						"amount": "1000000000000000000000"
					}
				}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		msg.Sender.Address = "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"
		payload := funcs.Must(args.Pack(ver, bz))

		translatedBz, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		jsonObject := make(map[string]interface{})
		err = json.Unmarshal(translatedBz, &jsonObject)
		assert.NoError(t, err)

		wasmRaw, ok := jsonObject["wasm"]
		assert.True(t, ok)

		wasm, ok := wasmRaw.(map[string]interface{})
		assert.True(t, ok)

		// make sure execute message can be serialized
		_, err = json.Marshal(wasmRaw)
		assert.NoError(t, err)

		wasmMsg, ok := wasm["msg"].(map[string]interface{})
		assert.True(t, ok)

		executeMsg, ok := wasmMsg["contract_name"].(map[string]interface{})
		assert.True(t, ok)

		actualSourceChain, ok := executeMsg["source_chain"].(string)
		assert.True(t, ok)
		assert.Equal(t, "Ethereum", actualSourceChain)

		actualSourceAddress, ok := executeMsg["source_address"].(string)
		assert.True(t, ok)
		assert.Equal(t, "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955", actualSourceAddress)

		actualNested, ok := executeMsg["nested"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", actualNested["denom"])
		assert.Equal(t, "1000000000000000000000", actualNested["amount"])
	})
}

package types_test

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	fmt "fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var (
	boolType              = funcs.Must(abi.NewType("bool", "bool", nil))
	addressType           = funcs.Must(abi.NewType("address", "address", nil))
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

func TestTranslateMessage(t *testing.T) {
	t.Run("should return error if version encoding is invalid", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		_, err := types.TranslateMessage(msg, []byte{0x01})
		assert.ErrorContains(t, err, "invalid versioned payload")
	})

	t.Run("should return error if invalid version", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := axelartestutils.PackPayloadWithVersion("0x99999999", rand.Bytes(64))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "unknown payload version")
	})
}

func TestConstructWasmMessageV1Large(t *testing.T) {
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

	payload := funcs.Must(constructABIPayload(
		methodName,
		argumentNames,
		[]abi.Type{
			stringType,
			stringType,
			uint8Type,
			uint64Type,
			boolType,
			stringArrayType,
			uint64ArrayType,
			uint64ArrayNestedType,
			stringArrayNestedType,
			bytesType,
			stringType},
		[]interface{}{
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
		},
	))

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

	msg, err := types.TranslateMessage(gm, payload)
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

func TestConstructWasmMessageV1(t *testing.T) {
	version := types.CosmWasmV1

	t.Run("should return error if invalid abi encoding", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := axelartestutils.PackPayloadWithVersion(version, rand.Bytes(10))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "abi:")
	})

	t.Run("should return error if abi encoding has trailing data", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		methodArguments := abi.Arguments([]abi.Argument{{Type: stringType}})
		schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}, {Type: uint64Type}}

		bz, err := schema.Pack(
			"method",
			[]string{"x"},
			[]string{"string"},
			funcs.Must(methodArguments.Pack("hello, world!")),
			uint64(0),
		)
		assert.NoError(t, err)

		payload := axelartestutils.PackPayloadWithVersion(version, bz)
		_, err = types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "wrong data")
	})

	t.Run("should return error if mismatching arg names", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		methodArguments := abi.Arguments([]abi.Argument{{Type: stringType}})
		argValues, err := methodArguments.Pack("hello world")
		assert.NoError(t, err)

		schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

		bz, err := schema.Pack(
			"method",
			[]string{"x", "y"},
			[]string{"string"},
			argValues,
		)
		assert.NoError(t, err)

		payload := axelartestutils.PackPayloadWithVersion(version, bz)
		_, err = types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "payload argument name and type length mismatch")
	})

	t.Run("should return error if invalid arg types", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		methodArguments := abi.Arguments([]abi.Argument{{Type: stringType}})
		argValues, err := methodArguments.Pack("hello world")
		assert.NoError(t, err)

		schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

		bz, err := schema.Pack(
			"method",
			[]string{"x"},
			[]string{"unknown"},
			argValues,
		)
		assert.NoError(t, err)

		payload := axelartestutils.PackPayloadWithVersion(version, bz)
		_, err = types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "invalid argument type")
	})

	t.Run("should return error if mismatching arg length", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		methodArguments := abi.Arguments([]abi.Argument{
			{Type: stringType}, {Type: stringType},
		})
		argValues, err := methodArguments.Pack("hello", "world")
		assert.NoError(t, err)

		schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

		bz, err := schema.Pack(
			"method",
			[]string{"x"},
			[]string{"string"},
			argValues,
		)
		assert.NoError(t, err)

		payload := axelartestutils.PackPayloadWithVersion(version, bz)
		_, err = types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "wrong data")
	})

	t.Run("should return error if invalid source chain type", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := funcs.Must(constructABIPayload(
			"method",
			[]string{"source_chain", "source_address"},
			[]abi.Type{boolType, stringType},
			[]interface{}{true, msg.GetSourceAddress()},
		))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source chain must have type string")
	})

	t.Run("should return error if invalid source chain value", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := funcs.Must(constructABIPayload(
			"method",
			[]string{"source_chain", "source_address"},
			[]abi.Type{stringType, stringType},
			[]interface{}{rand.Str(10), msg.GetSourceAddress()},
		))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source chain does not match expected")
	})

	t.Run("should return error if invalid source address type", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := funcs.Must(constructABIPayload(
			"method",
			[]string{"source_chain", "source_address"},
			[]abi.Type{stringType, uint64Type},
			[]interface{}{msg.GetSourceChain().String(), uint64(1)},
		))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should return error if invalid source address value", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		payload := funcs.Must(constructABIPayload(
			"method",
			[]string{"source_chain", "source_address"},
			[]abi.Type{stringType, stringType},
			[]interface{}{msg.GetSourceChain().String(), "invalid"},
		))

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should succeed with source address being address type", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		method := rand.StrBetween(0, 10)
		msg.Sender.Address = "0x" + rand.HexStr(40)
		args := struct {
			SourceChain   string `json:"source_chain"`
			SourceAddress string `json:"source_address"`
		}{
			SourceChain:   msg.GetSourceChain().String(),
			SourceAddress: msg.GetSourceAddress(),
		}

		payload := funcs.Must(constructABIPayload(
			method,
			[]string{"source_chain", "source_address"},
			[]abi.Type{stringType, addressType},
			[]interface{}{args.SourceChain, common.HexToAddress(args.SourceAddress)},
		))

		translatedMsg, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		checkWasmMsg(t, translatedMsg, msg, method, args)
	})

	t.Run("should succeed if valid source chain and address", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		method := rand.Str(10)
		args := struct {
			SourceChain   string `json:"source_chain"`
			SourceAddress string `json:"source_address"`
		}{
			SourceChain:   msg.GetSourceChain().String(),
			SourceAddress: msg.GetSourceAddress(),
		}

		payload := funcs.Must(constructABIPayload(
			method,
			[]string{"source_chain", "source_address"},
			[]abi.Type{stringType, stringType},
			[]interface{}{args.SourceChain, args.SourceAddress},
		))

		translatedMsg, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		checkWasmMsg(t, translatedMsg, msg, method, args)
	})

	t.Run("should succeed if valid args and source chain", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		method := rand.Str(10)
		args := struct {
			X           bool
			Y           uint64
			SourceChain string `json:"source_chain"`
		}{
			X:           rand.Bools(0.5).Next(),
			Y:           uint64(rand.I64Between(0, 1000)),
			SourceChain: msg.GetSourceChain().String(),
		}

		payload := funcs.Must(constructABIPayload(
			method,
			[]string{"x", "y", "source_chain"},
			[]abi.Type{boolType, uint64Type, stringType},
			[]interface{}{args.X, args.Y, args.SourceChain},
		))

		translatedMsg, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		checkWasmMsg(t, translatedMsg, msg, method, args)
	})
}

func TestConstructWasmMessageV2(t *testing.T) {
	version := types.CosmWasmV2

	t.Run("should return error if payload is not a json object", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		wasmMsg := []byte(`"a json string"`)
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "cannot unmarshal string into Go value of type map[string]interface {}")
	})

	t.Run("should return error if payload is not a valid json object", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		wasmMsg := []byte(`{"key": "invalid json with open bracket"`)
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)

		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "unexpected end of JSON input")
	})

	t.Run("should return error if wasm call has multiple methods", func(t *testing.T) {
		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": [3, 12, 143]},
				"contract_name2": {"source_chain": "ethereum", "source_address": [3, 12, 143]}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "invalid payload")
	})

	t.Run("should return error if wasm call has no args", func(t *testing.T) {
		wasmMsg := []byte(`{"contract_name": [1,2,3,4,5]}`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "invalid arguments")
	})

	t.Run("should return error if incorrect payload contains incorrect source chain type", func(t *testing.T) {
		wasmMsg := []byte(`{"contract_name": {"source_chain": 1.1, "source_address": [3, 12, 143]}}`)

		msg := nexustestutils.RandomMessage()
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source chain must have type string")
	})

	t.Run("should return error if incorrect payload contains incorrect source chain value", func(t *testing.T) {
		wasmMsg := []byte(`{"contract_name": {"source_chain": "unknown", "source_address": [3, 12, 143]}}`)

		msg := nexustestutils.RandomMessage()
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source chain does not match expected")
	})

	t.Run("should return error if incorrect payload contains incorrect source address type", func(t *testing.T) {
		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": [3, 12, 143]}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should return error if incorrect payload contains incorrect source address value", func(t *testing.T) {
		wasmMsg := []byte(`
			{
				"contract_name": {"source_chain": "ethereum", "source_address": "axelar123"}
			}
		`)

		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "ethereum"
		payload := axelartestutils.PackPayloadWithVersion(version, wasmMsg)
		_, err := types.TranslateMessage(msg, payload)
		assert.ErrorContains(t, err, "source address does not match expected")
	})

	t.Run("should construct wasm message without modifying payload", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		method := "contract_name"
		wasmArgs := []byte(`
			{
				"float": 4.51930,
				"nested": {
					"array": [1, 2, 3, 4],
					"amount": 100000000000000000000000000000001
				},
				"array": [0, -32323, 84739338387784623428752342, -43785623.2342532],
				"nil": null
			}
		`)
		bz := []byte(fmt.Sprintf("\t\t\n\n\t{\"%s\":%s}\n\t\n\t", method, wasmArgs))

		payload := axelartestutils.PackPayloadWithVersion(version, bz)
		translatedMsg, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		// whitespace is trimmed since checkWasmMsg only compares the arg json object
		checkWasmMsg[json.RawMessage](t, translatedMsg, msg, method, []byte(strings.TrimSpace(string(wasmArgs))))
	})

	t.Run("should construct wasm message from plain json", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		msg.Sender.Chain.Name = "Ethereum"
		msg.Sender.Address = "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"
		method := rand.Str(10)
		wasmArgs := []byte(`
			{
				"source_chain": "Ethereum",
				"source_address": "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955",
				"nested": {
					"denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
					"amount": "1000000000000000000000"
				}
			}
		`)
		bz := []byte(fmt.Sprintf("{\"%s\": %s}", method, wasmArgs))

		payload := axelartestutils.PackPayloadWithVersion(version, bz)

		translatedMsg, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		// whitespace is trimmed since checkWasmMsg only compares the arg json object
		checkWasmMsg[json.RawMessage](t, translatedMsg, msg, method, []byte(strings.TrimSpace(string(wasmArgs))))

		jsonObject := make(map[string]interface{})
		err = json.Unmarshal(translatedMsg, &jsonObject)
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

		executeMsg, ok := wasmMsg[method].(map[string]interface{})
		assert.True(t, ok)

		actualSourceChain, ok := executeMsg["source_chain"].(string)
		assert.True(t, ok)
		assert.Equal(t, msg.GetSourceChain().String(), actualSourceChain)

		actualSourceAddress, ok := executeMsg["source_address"].(string)
		assert.True(t, ok)
		assert.Equal(t, msg.GetSourceAddress(), actualSourceAddress)

		actualNested, ok := executeMsg["nested"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", actualNested["denom"])
		assert.Equal(t, "1000000000000000000000", actualNested["amount"])
	})

	t.Run("should construct wasm message from constructed json", func(t *testing.T) {
		msg := nexustestutils.RandomMessage()
		msg.Sender.Address = "0x" + rand.HexStr(40)

		wasmArgs := struct {
			SourceChain   string `json:"source_chain"`
			SourceAddress string `json:"source_address"`
			Asset         sdk.Coin
			Ints          []int64
			Floats        []float64
			Nil           []byte
			Map           map[int]string
		}{
			SourceChain:   msg.GetSourceChain().String(),
			SourceAddress: strings.ToUpper(msg.GetSourceAddress()),
			Asset: sdk.NewCoin(
				rand.Denom(3, 20),
				rand.IntBetween(sdk.ZeroInt(), sdk.NewIntFromUint64(10000000)),
			),
			Ints:   []int64{0, math.MaxInt64, math.MinInt64, rand.I64Between(math.MinInt64/2, math.MaxInt64/2)},
			Floats: []float64{math.Pi, math.MaxFloat64, math.SmallestNonzeroFloat64},
			Nil:    nil,
			Map: map[int]string{
				0:           "zero",
				-10:         rand.Str(10),
				math.MaxInt: "max int",
			},
		}
		method := rand.StrBetween(0, 10)
		bz := []byte(fmt.Sprintf("{\"%s\": %s}", method, funcs.Must(json.MarshalIndent(wasmArgs, "", "    "))))
		payload := axelartestutils.PackPayloadWithVersion(version, bz)

		translatedBz, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		checkWasmMsg(t, translatedBz, msg, method, wasmArgs)
	})
}

func TestConstructNativeV1Message(t *testing.T) {
	t.Run("should translate native payload", func(t *testing.T) {
		payloadMsg := rand.Bytes(int(rand.I64Between(1, 50)))

		msg := nexustestutils.RandomMessage()
		payload := axelartestutils.PackPayloadWithVersion(types.NativeV1, payloadMsg)

		translatedBz, err := types.TranslateMessage(msg, payload)
		assert.NoError(t, err)

		var decodedMsg struct {
			SourceChain   string `json:"source_chain"`
			SourceAddress string `json:"source_address"`
			Payload       []byte `json:"payload"`
			Type          int64  `json:"type"`
		}

		err = json.Unmarshal(translatedBz, &decodedMsg)
		assert.NoError(t, err)

		assert.Equal(t, decodedMsg.SourceChain, msg.Sender.Chain.Name.String())
		assert.Equal(t, decodedMsg.SourceAddress, msg.Sender.Address)
		assert.Equal(t, decodedMsg.Type, int64(msg.Type()))
		assert.Equal(t, decodedMsg.Payload, payloadMsg)
	})
}

func constructABIPayload(method string, argNames []string, argTypes []abi.Type, args []interface{}) ([]byte, error) {
	methodArguments := abi.Arguments(slices.Map(argTypes, func(argType abi.Type) abi.Argument { return abi.Argument{Type: argType} }))

	argValues, err := methodArguments.Pack(args...)
	if err != nil {
		return nil, err
	}

	schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}

	payload, err := schema.Pack(
		method,
		argNames,
		slices.Map(argTypes, abi.Type.String),
		argValues,
	)
	if err != nil {
		return nil, err
	}

	return axelartestutils.PackPayloadWithVersion(types.CosmWasmV1, payload), nil
}

// checkWasmMsg checks that a wasm msg is correctly formatted
func checkWasmMsg[T any](t assert.TestingT, payload []byte, msg nexus.GeneralMessage, method string, args T) {
	// json unmarshalling behaviour differs when unmarshalling to map[string]interface{} vs a struct, so cover both cases
	var jsonObject map[string]interface{}
	err := json.Unmarshal(payload, &jsonObject)
	assert.NoError(t, err)

	wasm, ok := jsonObject["wasm"].(map[string]interface{})
	assert.True(t, ok)

	sourceChain, ok := wasm["source_chain"]
	assert.True(t, ok)
	assert.Equal(t, sourceChain, msg.GetSourceChain().String())

	sourceAddress, ok := wasm["source_address"]
	assert.True(t, ok)
	assert.Equal(t, sourceAddress, msg.GetSourceAddress())

	contract, ok := wasm["contract"]
	assert.True(t, ok)
	assert.Equal(t, contract, msg.GetDestinationAddress())

	wasmMsg, ok := wasm["msg"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, len(wasmMsg), 1)

	_, ok = wasmMsg[method].(map[string]interface{})
	assert.True(t, ok)

	// now structurally decode the JSON to retrieve the typed wasm arguments
	var typedMsg struct {
		Wasm struct {
			SourceChain   string       `json:"source_chain"`
			SourceAddress string       `json:"source_address"`
			Contract      string       `json:"contract"`
			Msg           map[string]T `json:"msg"`
		} `json:"wasm"`
	}

	err = json.Unmarshal(payload, &typedMsg)
	assert.NoError(t, err)

	wasmArgs, ok := typedMsg.Wasm.Msg[method]
	assert.True(t, ok)

	assert.Equal(t, args, wasmArgs)
}

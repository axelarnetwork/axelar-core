package tests

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
)

func TestDeployToken(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()

	details := types.TokenDetails{
		TokenName: rand.Str(10),
		Symbol:    rand.Str(3),
		Decimals:  uint8(rand.I64Between(3, 10)),
		Capacity:  sdk.NewIntFromBigInt(big.NewInt(rand.I64Between(100, 100000))),
	}
	address := types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))

	capBz := make([]byte, 8)
	binary.BigEndian.PutUint64(capBz, details.Capacity.Uint64())
	capHex := hex.EncodeToString(capBz)

	expectedParams := fmt.Sprintf("00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000%s%s000000000000000000000000%s000000000000000000000000000000000000000000000000000000000000000a%s000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003%s0000000000000000000000000000000000000000000000000000000000",
		hex.EncodeToString([]byte{byte(details.Decimals)}),
		strings.Repeat("0", 64-len(capHex))+capHex,
		hex.EncodeToString(address.Bytes()),
		hex.EncodeToString([]byte(details.TokenName)),
		hex.EncodeToString([]byte(details.Symbol)),
	)
	actual, err := types.CreateDeployTokenCommand(chainID, keyID, details, address)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedName, decodedSymbol, decodedDecs, decodedCap, err := types.DecodeDeployTokenParams(actual.Params)
	assert.NoError(t, err)
	assert.Equal(t, details.TokenName, decodedName)
	assert.Equal(t, details.Symbol, decodedSymbol)
	assert.Equal(t, details.Decimals, decodedDecs)
	assert.Equal(t, details.Capacity.BigInt(), decodedCap)
}

func TestCreateMintTokenCommand(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	commandID := types.NewCommandID(rand.Bytes(32), chainID)
	symbol := rand.Str(3)
	address := common.BytesToAddress(rand.Bytes(common.AddressLength))
	amount := big.NewInt(rand.I64Between(100, 100000))

	amountBz := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBz, amount.Uint64())
	amountHex := hex.EncodeToString(amountBz)

	expectedParams := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000060000000000000000000000000%s%s0000000000000000000000000000000000000000000000000000000000000003%s0000000000000000000000000000000000000000000000000000000000",
		hex.EncodeToString(address.Bytes()),
		strings.Repeat("0", 64-len(amountHex))+amountHex,
		hex.EncodeToString([]byte(symbol)),
	)
	actual, err := types.CreateMintTokenCommand(keyID, commandID, symbol, address, amount)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedSymbol, decodedAddr, decodedAmount, err := types.DecodeMintTokenParams(actual.Params)
	assert.NoError(t, err)
	assert.Equal(t, symbol, decodedSymbol)
	assert.Equal(t, address, decodedAddr)
	assert.Equal(t, amount, decodedAmount)

}

func TestCreateBurnTokenCommand(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	symbol := rand.Str(3)
	salt := common.BytesToHash(rand.Bytes(common.HashLength))
	height := rand.I64Between(100, 10000)

	expectedParams := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000040%s0000000000000000000000000000000000000000000000000000000000000003%s0000000000000000000000000000000000000000000000000000000000",
		hex.EncodeToString(salt.Bytes()),
		hex.EncodeToString([]byte(symbol)),
	)
	actual, err := types.CreateBurnTokenCommand(
		chainID,
		keyID,
		height,
		types.BurnerInfo{Symbol: symbol, Salt: types.Hash(salt)},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedSymbol, decodedSalt, err := types.DecodeBurnTokenParams(actual.Params)
	assert.NoError(t, err)
	assert.Equal(t, symbol, decodedSymbol)
	assert.Equal(t, salt, decodedSalt)
}

func TestCreateSinglesigTransferCommand_Ownership(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	newOwnerAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))

	expectedParams := fmt.Sprintf("000000000000000000000000%s", hex.EncodeToString(newOwnerAddr.Bytes()))
	actual, err := types.CreateSinglesigTransferCommand(
		types.Ownership,
		chainID,
		keyID,
		newOwnerAddr,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedAddr, err := types.DecodeTransferSinglesigParams(actual.Params)
	assert.NoError(t, err)
	assert.Equal(t, newOwnerAddr, decodedAddr)
}

func TestCreateSinglesigTransferCommand_Operatorship(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	newOperatorAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))

	expectedParams := fmt.Sprintf("000000000000000000000000%s", hex.EncodeToString(newOperatorAddr.Bytes()))
	actual, err := types.CreateSinglesigTransferCommand(
		types.Operatorship,
		chainID,
		keyID,
		newOperatorAddr,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedAddr, err := types.DecodeTransferSinglesigParams(actual.Params)
	assert.NoError(t, err)
	assert.Equal(t, newOperatorAddr, decodedAddr)
}

func TestCreateMultisigTransferCommand_Ownership(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()

	addresses := []common.Address{
		common.HexToAddress("0xd59ca627Af68D29C547B91066297a7c469a7bF72"),
		common.HexToAddress("0xc2FCc7Bcf743153C58Efd44E6E723E9819E9A10A"),
		common.HexToAddress("0x2ad611e02E4F7063F515C8f190E5728719937205"),
	}
	threshold := uint8(2)

	expectedParams := "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003000000000000000000000000d59ca627af68d29c547b91066297a7c469a7bf72000000000000000000000000c2fcc7bcf743153c58efd44e6e723e9819e9a10a0000000000000000000000002ad611e02e4f7063f515c8f190e5728719937205"
	actual, err := types.CreateMultisigTransferCommand(
		types.Ownership,
		chainID,
		keyID,
		threshold,
		addresses...,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedAddrs, decodedThreshold, err := types.DecodeTransferMultisigParams(actual.Params)
	assert.NoError(t, err)
	assert.ElementsMatch(t, addresses, decodedAddrs)
	assert.Equal(t, threshold, decodedThreshold)
}
func TestCreateMultisigTransferCommand_Operatorship(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()

	addresses := []common.Address{
		common.HexToAddress("0xd59ca627Af68D29C547B91066297a7c469a7bF72"),
		common.HexToAddress("0xc2FCc7Bcf743153C58Efd44E6E723E9819E9A10A"),
		common.HexToAddress("0x2ad611e02E4F7063F515C8f190E5728719937205"),
	}
	threshold := uint8(2)

	expectedParams := "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003000000000000000000000000d59ca627af68d29c547b91066297a7c469a7bf72000000000000000000000000c2fcc7bcf743153c58efd44e6e723e9819e9a10a0000000000000000000000002ad611e02e4f7063f515c8f190e5728719937205"
	actual, err := types.CreateMultisigTransferCommand(
		types.Ownership,
		chainID,
		keyID,
		threshold,
		addresses...,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedAddrs, decodedThreshold, err := types.DecodeTransferMultisigParams(actual.Params)
	assert.NoError(t, err)
	assert.ElementsMatch(t, addresses, decodedAddrs)
	assert.Equal(t, threshold, decodedThreshold)
}

func TestGetSignHash(t *testing.T) {
	data := common.FromHex("0000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f")

	expected := "0xe7bce8f57491e71212d930096bacf9288c711e5f27200946edd570e3a93546bf"
	actual := types.GetSignHash(data)

	assert.Equal(t, expected, actual.Hex())
}

func TestCreateExecuteDataSinglesig(t *testing.T) {
	commandData := common.FromHex("0000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f")
	commandSig := types.Signature{}
	copy(commandSig[:], common.FromHex("42b936b3c37fb7deed86f52154798d0c9abfe5ba838b2488f4a7e5193a9bb60b5d8c521f5c8c64f9442fc745ecd3bc496b04dc03a81b4e89c72342ab5903284d1c"))

	expected := "09c5eabe000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002e00000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000026000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f000000000000000000000000000000000000000000000000000000000000004142b936b3c37fb7deed86f52154798d0c9abfe5ba838b2488f4a7e5193a9bb60b5d8c521f5c8c64f9442fc745ecd3bc496b04dc03a81b4e89c72342ab5903284d1c00000000000000000000000000000000000000000000000000000000000000"
	actual, err := types.CreateExecuteDataSinglesig(commandData, commandSig)

	assert.NoError(t, err)
	assert.Equal(t, expected, common.Bytes2Hex(actual))
}

func TestCreateExecuteDataMultisig(t *testing.T) {
	commandData := common.FromHex("0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000140000000000000000000000000000000000000000000000000000000000000000186c71b9698cc55f8238266b026414ed9880bcd3dafd254cfc1079f1d4c2098800000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000006000000000000000000000000017ec8597ff92c3f44523bdc65bf0f1be632917ff000000000000000000000000000000000000000000000000000000000152a1c000000000000000000000000000000000000000000000000000000000000000034141540000000000000000000000000000000000000000000000000000000000")
	commandSigs := make([]types.Signature, 2)
	copy(commandSigs[0][:], common.FromHex("226f548e306ba150c2895f192c71de4e455655508bb0762d6808756ac5cae9dd41145781fa6f7bcd52c3a71d492b3bf15d8792c431568e1b379b8d52a479b0971c"))
	copy(commandSigs[1][:], common.FromHex("44e9e6a66df68d798802914c41f57c0ef488e0ca5f244afa60e3438a5078356803213e2de2f4d41a4002fb3115722b5804ff8cd0a5101d7b37ba97fadd223fc51b"))

	expected := "09c5eabe00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000002a000000000000000000000000000000000000000000000000000000000000002400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000140000000000000000000000000000000000000000000000000000000000000000186c71b9698cc55f8238266b026414ed9880bcd3dafd254cfc1079f1d4c2098800000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000006000000000000000000000000017ec8597ff92c3f44523bdc65bf0f1be632917ff000000000000000000000000000000000000000000000000000000000152a1c0000000000000000000000000000000000000000000000000000000000000000341415400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000041226f548e306ba150c2895f192c71de4e455655508bb0762d6808756ac5cae9dd41145781fa6f7bcd52c3a71d492b3bf15d8792c431568e1b379b8d52a479b0971c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004144e9e6a66df68d798802914c41f57c0ef488e0ca5f244afa60e3438a5078356803213e2de2f4d41a4002fb3115722b5804ff8cd0a5101d7b37ba97fadd223fc51b00000000000000000000000000000000000000000000000000000000000000"
	actual, err := types.CreateExecuteDataMultisig(commandData, commandSigs...)

	assert.NoError(t, err)
	assert.Equal(t, expected, common.Bytes2Hex(actual))
}

func TestGetTokenAddress(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)

	chain := "Ethereum"
	asset := "axelar"
	tokenName := "axelar token"
	tokenSymbol := "at"
	decimals := uint8(18)
	capacity := sdk.NewIntFromUint64(uint64(10000))

	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
	expected := types.Address(common.HexToAddress("0xcf097C27D8C22a2351a05dA7106aA4471a198c2C"))

	keeper := k.ForChain(chain)
	keeper.SetParams(ctx, types.DefaultParams()[0])
	keeper.SetPendingGateway(ctx, axelarGateway)
	keeper.ConfirmPendingGateway(ctx)
	tokenDetails := types.NewTokenDetails(tokenName, tokenSymbol, decimals, capacity)
	token, err := keeper.CreateERC20Token(ctx, asset, tokenDetails, sdk.NewInt(1000000), types.ZeroAddress)
	assert.NoError(t, err)
	assert.Equal(t, expected, token.GetAddress())
}

func TestGetBurnerAddressAndSalt(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	ctx = ctx.WithHeaderHash(common.Hex2Bytes("CA36CA3751A5B6E8B8ED4072BFA5E6E5BAC8B6E06E02DE029E1BD86AB141F2F1"))
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000))
	ctx.GasMeter().ConsumeGas(1000, "test")
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)

	t.Run("should work for internal erc20 tokens", testutils.Func(func(t *testing.T) {
		axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
		recipient := "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L"
		tokenAddr := types.Address(common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D"))
		expectedBurnerAddr := common.HexToAddress("0x54507775880357AD534957B0520009b15CAB7c6F")
		expectedSalt := common.Hex2Bytes("b365d534cb5d28d511a8baf1125240c97b09cb46710645b30ed64f302c4ae7ff")

		chainKeeper := k.ForChain(exported.Ethereum.Name)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		actualburnerAddr, actualSalt, err := chainKeeper.GetBurnerAddressAndSalt(ctx, tokenAddr, recipient, axelarGateway, false)

		assert.NoError(t, err)
		assert.Equal(t, expectedBurnerAddr, actualburnerAddr)
		assert.Equal(t, common.Bytes2Hex(expectedSalt), common.Bytes2Hex(actualSalt[:]))
	}))

	t.Run("should work for external erc20 tokens", testutils.Func(func(t *testing.T) {
		axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
		recipient := "axelar1aguuy756cpaqnfd5t5qn68u7ck7w2sp64023hk"
		tokenAddr := types.Address(common.HexToAddress("0xFDFEF9D10d929cB3905C71400ce6be1990EA0F34"))
		expectedBurnerAddr := common.HexToAddress("0xCAd52c6D47F7c759732A27556e71288B610Dbcfb")
		expectedSalt := common.Hex2Bytes("2321c4ff5401853a7a9960fd93a0281cde689966a62d049bdc5c5b16733954f1")

		chainKeeper := k.ForChain(exported.Ethereum.Name)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		actualburnerAddr, actualSalt, err := chainKeeper.GetBurnerAddressAndSalt(ctx, tokenAddr, recipient, axelarGateway, true)

		assert.NoError(t, err)
		assert.Equal(t, expectedBurnerAddr, actualburnerAddr)
		assert.Equal(t, common.Bytes2Hex(expectedSalt), common.Bytes2Hex(actualSalt[:]))
	}))
}

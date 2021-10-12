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
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
)

func TestCreateMintTokenCommand_CorrectParams(t *testing.T) {
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
	actual, err := types.CreateMintTokenCommand(
		chainID,
		keyID,
		commandID,
		symbol,
		address,
		amount,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))
}

func TestCreateBurnTokenCommand_CorrectParams(t *testing.T) {
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
}

func TestCreateTransferOwnershipCommand_CorrectParams(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	newOwnerAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))

	expectedParams := fmt.Sprintf("000000000000000000000000%s", hex.EncodeToString(newOwnerAddr.Bytes()))
	actual, err := types.CreateTransferOwnershipCommand(
		chainID,
		keyID,
		newOwnerAddr,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))
}

func TestCreateTransferOperatorshipCommand_CorrectParams(t *testing.T) {
	chainID := big.NewInt(1)
	keyID := tssTestUtils.RandKeyID()
	newOperatorAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))

	expectedParams := fmt.Sprintf("000000000000000000000000%s", hex.EncodeToString(newOperatorAddr.Bytes()))
	actual, err := types.CreateTransferOperatorshipCommand(
		chainID,
		keyID,
		newOperatorAddr,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))
}

func TestGetSignHash_CorrectSignHash(t *testing.T) {
	data := common.FromHex("0000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f")

	expected := "0xe7bce8f57491e71212d930096bacf9288c711e5f27200946edd570e3a93546bf"
	actual := types.GetSignHash(data)

	assert.Equal(t, expected, actual.Hex())
}

func TestCreateExecuteData_CorrectExecuteData(t *testing.T) {
	commandData := common.FromHex("0000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f")
	commandSig := types.Signature{}
	copy(commandSig[:], common.FromHex("42b936b3c37fb7deed86f52154798d0c9abfe5ba838b2488f4a7e5193a9bb60b5d8c521f5c8c64f9442fc745ecd3bc496b04dc03a81b4e89c72342ab5903284d1c"))

	expected := "09c5eabe000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002e00000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000026000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000001ec78d9c22c08bb9f0ecd5d95571ae83e3f22219c5a9278c3270691d50abfd91b000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000096d696e74546f6b656e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000014141540000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000063fc2ad3d021a4d7e64323529a55a9442c444da00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000270f000000000000000000000000000000000000000000000000000000000000004142b936b3c37fb7deed86f52154798d0c9abfe5ba838b2488f4a7e5193a9bb60b5d8c521f5c8c64f9442fc745ecd3bc496b04dc03a81b4e89c72342ab5903284d1c00000000000000000000000000000000000000000000000000000000000000"
	actual, err := types.CreateExecuteData(commandData, commandSig)

	assert.NoError(t, err)
	assert.Equal(t, expected, common.Bytes2Hex(actual))
}

func TestGetTokenAddress_CorrectData(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("testKey"), paramsK)

	chain := "Ethereum"
	asset := "axelar"
	tokenName := "axelar token"
	tokenSymbol := "at"
	decimals := uint8(18)
	capacity := sdk.NewIntFromUint64(uint64(10000))

	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
	expected := types.Address(common.HexToAddress("0xf0a19fEAF7B6121817f999D3d5405CB0B419Cfa8"))

	k.SetParams(ctx, types.DefaultParams()...)
	keeper := k.ForChain(chain)
	keeper.SetGatewayAddress(ctx, axelarGateway)
	tokenDetails := types.NewTokenDetails(tokenName, tokenSymbol, decimals, capacity)
	token, err := keeper.CreateERC20Token(ctx, asset, tokenDetails)
	assert.NoError(t, err)
	assert.Equal(t, expected, token.GetAddress())
}

func TestGetBurnerAddressAndSalt_CorrectData(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("testKey"), paramsK)

	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
	recipient := "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L"
	tokenAddr := types.Address(common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D"))
	expectedBurnerAddr := common.HexToAddress("0xC857f4173BdC159B6254504ABd88d144eba6Aa1B")
	expectedSalt := common.Hex2Bytes("35f28b34202f4e3de20c1710696e3f294ebe4df686b17be00fedf991190f9654")

	k.SetParams(ctx, types.DefaultParams()...)

	actualburnerAddr, actualSalt, err := k.ForChain(exported.Ethereum.Name).GetBurnerAddressAndSalt(ctx, tokenAddr, recipient, axelarGateway)

	assert.NoError(t, err)
	assert.Equal(t, expectedBurnerAddr, actualburnerAddr)
	assert.Equal(t, expectedSalt, actualSalt[:])
}

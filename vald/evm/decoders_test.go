package evm_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

func TestDecodeEventTokenSent(t *testing.T) {
	log := &geth.Log{
		Topics: []common.Hash{
			common.HexToHash("0x651d93f66c4329630e8d0f62488eff599e3be484da587335e8dc0fcf46062726"),
			common.HexToHash("0x00000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb8"),
		},
		Data: common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000989680000000000000000000000000000000000000000000000000000000000000000a657468657265756d2d3200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a30783538656134313033656439353564434262646338613066456261626133393542366534346431354600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f657468657265756d2d312d7561786c0000000000000000000000000000000000"),
	}

	expected := types.EventTokenSent{
		Sender:             types.Address(common.HexToAddress("0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8")),
		DestinationChain:   "ethereum-2",
		DestinationAddress: "0x58ea4103ed955dCBbdc8a0fEbaba395B6e44d15F",
		Symbol:             "ethereum-1-uaxl",
		Amount:             sdk.NewUint(10000000),
	}
	actual, err := evm.DecodeEventTokenSent(log)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestDecodeEventContractCall(t *testing.T) {
	log := &geth.Log{
		Topics: []common.Hash{
			common.HexToHash("0x30ae6cc78c27e651745bf2ad08a11de83910ac1e347a52f7ac898c0fbef94dae"),
			common.HexToHash("0x000000000000000000000000d48e199950589a4336e4dc43bd2c72ba0c0baa86"),
			common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e"),
		},
		Data: common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000a657468657265756d2d3200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a3078623938343566393234376138354565353932323733613739363035663334453836303764376537350000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000066275666665720000000000000000000000000000000000000000000000000000"),
	}

	expected := types.EventContractCall{
		Sender:           types.Address(common.HexToAddress("0xD48E199950589A4336E4dc43bd2C72Ba0C0baA86")),
		DestinationChain: "ethereum-2",
		ContractAddress:  "0xb9845f9247a85Ee592273a79605f34E8607d7e75",
		PayloadHash:      types.Hash(common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e")),
	}
	actual, err := evm.DecodeEventContractCall(log)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestDecodeEventContractCallWithToken(t *testing.T) {
	log := &geth.Log{
		Topics: []common.Hash{
			common.HexToHash("0x7e50569d26be643bda7757722291ec66b1be66d8283474ae3fab5a98f878a7a2"),
			common.HexToHash("0x00000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb8"),
			common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e"),
		},
		Data: common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000140000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000009896800000000000000000000000000000000000000000000000000000000000000008657468657265756d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a307837366130363034333339313731326245333941333433643166343331363538353466434636446533000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006627566666572000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000047561786c00000000000000000000000000000000000000000000000000000000"),
	}

	expected := types.EventContractCallWithToken{
		Sender:           types.Address(common.HexToAddress("0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8")),
		DestinationChain: "ethereum",
		ContractAddress:  "0x76a06043391712bE39A343d1f43165854fCF6De3",
		PayloadHash:      types.Hash(common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e")),
		Symbol:           "uaxl",
		Amount:           sdk.NewUint(10000000),
	}
	actual, err := evm.DecodeEventContractCallWithToken(log)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestDecodeTokenDeployEvent_CorrectData(t *testing.T) {
	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")

	tokenDeploySig := evm.ERC20TokenDeploymentSig
	expectedAddr := common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D")
	expectedSymbol := "XPTO"
	data := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000e7481ecb61f9c84b91c03414f3d5d48e5436045d00000000000000000000000000000000000000000000000000000000000000045850544f00000000000000000000000000000000000000000000000000000000")

	l := &geth.Log{Address: axelarGateway, Data: data, Topics: []common.Hash{tokenDeploySig}}

	tokenDeployed, err := evm.DecodeERC20TokenDeploymentEvent(l)
	assert.NoError(t, err)
	assert.Equal(t, expectedSymbol, tokenDeployed.Symbol)
	assert.Equal(t, types.Address(expectedAddr), tokenDeployed.TokenAddress)
}

func TestDecodeErc20TransferEvent_NotErc20Transfer(t *testing.T) {
	l := geth.Log{
		Topics: []common.Hash{
			common.BytesToHash(rand.Bytes(common.HashLength)),
			common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(rand.Bytes(common.AddressLength)).Bytes(), common.HashLength)),
			common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(rand.Bytes(common.AddressLength)).Bytes(), common.HashLength)),
		},
		Data: common.LeftPadBytes(big.NewInt(2).Bytes(), common.HashLength),
	}

	_, err := evm.DecodeERC20TransferEvent(&l)
	assert.Error(t, err)
}

func TestDecodeErc20TransferEvent_InvalidErc20Transfer(t *testing.T) {
	erc20TransferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	l := geth.Log{
		Topics: []common.Hash{
			erc20TransferEventSig,
			common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(rand.Bytes(common.AddressLength)).Bytes(), common.HashLength)),
		},
		Data: common.LeftPadBytes(big.NewInt(2).Bytes(), common.HashLength),
	}

	_, err := evm.DecodeERC20TransferEvent(&l)

	assert.Error(t, err)
}

func TestDecodeErc20TransferEvent_CorrectData(t *testing.T) {
	erc20TransferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	expectedFrom := common.BytesToAddress(rand.Bytes(common.AddressLength))
	expectedTo := common.BytesToAddress(rand.Bytes(common.AddressLength))
	expectedAmount := sdk.NewUint(uint64(rand.I64Between(1, 10000)))

	l := geth.Log{
		Topics: []common.Hash{
			erc20TransferEventSig,
			common.BytesToHash(common.LeftPadBytes(expectedFrom.Bytes(), common.HashLength)),
			common.BytesToHash(common.LeftPadBytes(expectedTo.Bytes(), common.HashLength)),
		},
		Data: common.LeftPadBytes(expectedAmount.BigInt().Bytes(), common.HashLength),
	}

	transfer, err := evm.DecodeERC20TransferEvent(&l)

	assert.NoError(t, err)
	assert.Equal(t, types.Address(expectedTo), transfer.To)
	assert.Equal(t, expectedAmount, transfer.Amount)
}

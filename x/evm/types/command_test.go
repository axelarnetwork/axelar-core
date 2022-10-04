package types

import (
	"encoding/binary"
	"encoding/hex"
	fmt "fmt"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigMock "github.com/axelarnetwork/axelar-core/x/multisig/exported/mock"
	multisigTestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func TestNewApproveContractCallWithMintCommand(t *testing.T) {
	chainID := sdk.NewInt(1)
	keyID := multisigTestutils.KeyID()
	sourceChain := nexus.ChainName("polygon")
	txID := Hash(common.HexToHash("0x5bb45dc24ddd6b90fa37f26eecfcf203328427c3226db29d1c01051b965ca93b"))
	index := uint64(99)
	sourceAddress := "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"
	contractAddress := common.HexToAddress("0x956dA338C1518a7FB213042b70c60c021aeBd554")
	payloadHash := common.HexToHash("0x7c6498469c4e2d466b6fc9af3c910587f6c0bdade714a16ab279a08a759a5c14")
	symbol := "testA"
	amount := sdk.NewUint(20000)
	event := EventContractCallWithToken{
		Sender:          Address(common.HexToAddress(sourceAddress)),
		ContractAddress: contractAddress.Hex(),
		PayloadHash:     Hash(payloadHash),
		Symbol:          rand.NormalizedStrBetween(1, 5),
	}

	expectedParams := "00000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000140000000000000000000000000956da338c1518a7fb213042b70c60c021aebd5547c6498469c4e2d466b6fc9af3c910587f6c0bdade714a16ab279a08a759a5c1400000000000000000000000000000000000000000000000000000000000001a00000000000000000000000000000000000000000000000000000000000004e205bb45dc24ddd6b90fa37f26eecfcf203328427c3226db29d1c01051b965ca93b00000000000000000000000000000000000000000000000000000000000000630000000000000000000000000000000000000000000000000000000000000007706f6c79676f6e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a3078363842393330343566653744383739346137634146333237653766383535434436436430334242380000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000057465737441000000000000000000000000000000000000000000000000000000"
	actual := NewApproveContractCallWithMintCommand(
		chainID,
		keyID,
		sourceChain,
		txID,
		index,
		event,
		amount,
		symbol,
	)

	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	actualSourceChain, actualSourceAddress, actualContractAddress, actualPayloadHash, actualSymbol, actualAmount, actualSourceTxID, actualSourceEventIndex := decodeApproveContractCallWithMintParams(actual.Params)
	assert.Equal(t, sourceChain.String(), actualSourceChain)
	assert.Equal(t, sourceAddress, actualSourceAddress)
	assert.Equal(t, contractAddress, actualContractAddress)
	assert.Equal(t, payloadHash, actualPayloadHash)
	assert.Equal(t, symbol, actualSymbol)
	assert.Equal(t, amount.BigInt(), actualAmount)
	assert.Equal(t, txID, Hash(actualSourceTxID))
	assert.Equal(t, index, actualSourceEventIndex.Uint64())
}

func TestNewMintTokenCommand(t *testing.T) {
	chainID := sdk.NewInt(1)
	keyID := multisigTestutils.KeyID()
	commandID := NewCommandID(rand.Bytes(32), chainID)
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
	actual := NewMintTokenCommand(keyID, commandID, symbol, address, amount)

	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedSymbol, decodedAddr, decodedAmount := decodeMintTokenParams(actual.Params)
	assert.Equal(t, symbol, decodedSymbol)
	assert.Equal(t, address, decodedAddr)
	assert.Equal(t, amount, decodedAmount)
}

func TestNewBurnTokenCommand(t *testing.T) {
	chainID := sdk.NewInt(1)
	keyID := multisigTestutils.KeyID()
	symbol := rand.Str(3)
	salt := common.BytesToHash(rand.Bytes(common.HashLength))
	height := rand.I64Between(100, 10000)

	expectedParams := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000040%s0000000000000000000000000000000000000000000000000000000000000003%s0000000000000000000000000000000000000000000000000000000000",
		hex.EncodeToString(salt.Bytes()),
		hex.EncodeToString([]byte(symbol)),
	)
	actual := NewBurnTokenCommand(
		chainID,
		keyID,
		height,
		BurnerInfo{Symbol: symbol, Salt: Hash(salt)},
		false,
	)

	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))

	decodedSymbol, decodedSalt := decodeBurnTokenParams(actual.Params)
	assert.Equal(t, symbol, decodedSymbol)
	assert.Equal(t, salt, decodedSalt)
}

func TestNewMultisigTransferCommand(t *testing.T) {
	pubKeys := slices.Map([]string{
		"046e0fc68835979b6f0248e284035fdc5084d15bf974908d06cafcba8c6da0ef4ab98be05e6a08529b0c869ab0cc2497dbe11b4293255a528ce53396305d7a09cf",
		"02f7f54741653c9f1ad9b84645a507e43e75f7dc6fe81d2629aeb36bd161f065ae",
		"03b13092611105a5d31403a7b6519c8149867932559c79fea8931a3948d413e625",
		"02c027059d874f594a6a36b9a4baac92a6fa50846d68434f6ac78c294a2b8decf7",
		"03b448a1acb25e9085bcd1f8d869043245f3d0a26b7d6112cdae44d9a2267aae50",
		"027a4089cf8ea231a8d09a01e420c327dfd62b8848621a1b21694e82869876d6fc",
		"02e659958a5e3c5ac33765342ab28e0ce0ed8a9f8833e837feb1c3ce29639f0b23",
		"034b2d8119648d8678220594750779618ee704228858f7238a8d0965cf70df1001",
	}, func(pk string) multisig.PublicKey { return funcs.Must(hex.DecodeString(pk)) })
	weights := slices.Map([]uint64{1, 2, 3, 4, 5, 6, 7, 8}, sdk.NewUint)
	participants := slices.Expand(func(_ int) sdk.ValAddress { return rand.ValAddr() }, len(pubKeys))

	key := &multisigMock.KeyMock{
		GetParticipantsFunc: func() []sdk.ValAddress { return participants },
		GetPubKeyFunc: func(v sdk.ValAddress) (multisig.PublicKey, bool) {
			for i, p := range participants {
				if v.Equals(p) {
					return pubKeys[i], true
				}
			}

			return nil, false
		},
		GetWeightFunc: func(v sdk.ValAddress) sdk.Uint {
			for i, p := range participants {
				if v.Equals(p) {
					return weights[i]
				}
			}

			return sdk.ZeroUint()
		},
		GetMinPassingWeightFunc: func() sdk.Uint { return sdk.NewUint(30) },
	}

	chainID := sdk.NewInt(1)
	keyID := multisigTestutils.KeyID()
	actual := NewMultisigTransferCommand(chainID, keyID, key)
	expectedParams := "00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000800000000000000000000000019cc2044857d23129a29f763d0338da837ce35f60000000000000000000000002ab6fa7de5e9e9423125a4246e4de1b9c755607400000000000000000000000037cc4b7e8f9f505ca8126db8a9d070566ed5dae70000000000000000000000003e56f0d4497ac44993d9ea272d4707f8be6b42a6000000000000000000000000462b96f617d5d92f63f9949c6f4626623ea73fa400000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb80000000000000000000000009e77c30badbbc412a0c20c6ce43b671c6f103434000000000000000000000000c1c0c8d2131cc866834c6382096eadfef1af2f52000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005"

	assert.Equal(t, expectedParams, hex.EncodeToString(actual.Params))
	assert.Equal(t, keyID, actual.KeyID)

	decodedAddresses, decodedWeights, decodedThreshold := decodeTransferMultisigParams(actual.Params)
	assert.Len(t, decodedAddresses, len(participants))
	assert.Len(t, decodedWeights, len(participants))
	assert.EqualValues(t, 30, decodedThreshold.Uint64())
}

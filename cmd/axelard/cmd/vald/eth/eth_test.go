package eth

import (
	"context"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	mock2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/types/mock"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/eth/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestDecodeTokenDeployEvent_CorrectData(t *testing.T) {
	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
	tokenDeploySig := ERC20TokenDeploymentSig
	expectedAddr := common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D")
	expectedSymbol := "XPTO"
	data := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000e7481ecb61f9c84b91c03414f3d5d48e5436045d00000000000000000000000000000000000000000000000000000000000000045850544f00000000000000000000000000000000000000000000000000000000")

	l := &geth.Log{Address: axelarGateway, Data: data, Topics: []common.Hash{tokenDeploySig}}

	symbol, tokenAddr, err := decodeERC20TokenDeploymentEvent(l)
	assert.NoError(t, err)
	assert.Equal(t, expectedSymbol, symbol)
	assert.Equal(t, expectedAddr, tokenAddr)
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

	_, _, err := decodeERC20TransferEvent(&l)

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

	_, _, err := decodeERC20TransferEvent(&l)

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

	actualTo, actualAmount, err := decodeERC20TransferEvent(&l)

	assert.NoError(t, err)
	assert.Equal(t, expectedTo, actualTo)
	assert.Equal(t, expectedAmount, actualAmount)
}

func TestMgr_ProccessDepositConfirmation(t *testing.T) {
	var (
		mgr         *Mgr
		attributes  []sdk.Attribute
		rpc         *mock.ClientMock
		broadcaster *mock2.BroadcasterMock
	)
	setup := func() {
		cdc := testutils.MakeEncodingConfig().Amino
		poll := exported.NewPollMetaWithNonce(ethTypes.ModuleName, rand.StrBetween(5, 20), rand.PosI64(), rand.I64Between(1, 1000))

		burnAddrBytes := rand.Bytes(common.AddressLength)
		tokenAddrBytes := rand.Bytes(common.AddressLength)
		blockNumber := rand.PInt64Gen().Where(func(i int64) bool { return i != 0 }).Next() // restrict to int64 so the block number in the receipt doesn't overflow
		confHeight := rand.I64Between(0, blockNumber-1)
		amount := rand.PosI64() // restrict to int64 so the amount in the receipt doesn't overflow
		attributes = []sdk.Attribute{
			sdk.NewAttribute(ethTypes.AttributeKeyTxID, common.Bytes2Hex(rand.Bytes(common.HashLength))),
			sdk.NewAttribute(ethTypes.AttributeKeyAmount, strconv.FormatUint(uint64(amount), 10)),
			sdk.NewAttribute(ethTypes.AttributeKeyBurnAddress, common.Bytes2Hex(burnAddrBytes)),
			sdk.NewAttribute(ethTypes.AttributeKeyTokenAddress, common.Bytes2Hex(tokenAddrBytes)),
			sdk.NewAttribute(ethTypes.AttributeKeyConfHeight, strconv.FormatUint(uint64(confHeight), 10)),
			sdk.NewAttribute(ethTypes.AttributeKeyPoll, string(cdc.MustMarshalJSON(poll))),
		}

		rpc = &mock.ClientMock{
			BlockNumberFunc: func(context.Context) (uint64, error) {
				return uint64(blockNumber), nil
			},
			TransactionReceiptFunc: func(context.Context, common.Hash) (*geth.Receipt, error) {
				receipt := &geth.Receipt{
					BlockNumber: big.NewInt(rand.I64Between(0, blockNumber-confHeight)),
					Logs: []*geth.Log{
						/* ERC20 transfer to burner address of a random token */
						{
							Address: common.BytesToAddress(rand.Bytes(common.AddressLength)),
							Topics: []common.Hash{
								ERC20TransferSig,
								common.BytesToHash(common.LeftPadBytes(rand.Bytes(common.AddressLength), common.HashLength)),
								common.BytesToHash(common.LeftPadBytes(burnAddrBytes, common.HashLength)),
							},
							Data: common.LeftPadBytes(big.NewInt(rand.PosI64()).Bytes(), common.HashLength),
						},
						/* not a ERC20 transfer */
						{
							Address: common.BytesToAddress(tokenAddrBytes),
							Topics: []common.Hash{
								common.BytesToHash(rand.Bytes(common.HashLength)),
								common.BytesToHash(common.LeftPadBytes(rand.Bytes(common.AddressLength), common.HashLength)),
								common.BytesToHash(common.LeftPadBytes(burnAddrBytes, common.HashLength)),
							},
							Data: common.LeftPadBytes(big.NewInt(rand.PosI64()).Bytes(), common.HashLength),
						},
						/* an invalid ERC20 transfer */
						{
							Address: common.BytesToAddress(tokenAddrBytes),
							Topics: []common.Hash{
								ERC20TransferSig,
								common.BytesToHash(common.LeftPadBytes(rand.Bytes(common.AddressLength), common.HashLength)),
							},
							Data: common.LeftPadBytes(big.NewInt(rand.PosI64()).Bytes(), common.HashLength),
						},
						/* an ERC20 transfer of our concern */
						{
							Address: common.BytesToAddress(tokenAddrBytes),
							Topics: []common.Hash{
								ERC20TransferSig,
								common.BytesToHash(common.LeftPadBytes(rand.Bytes(common.AddressLength), common.HashLength)),
								common.BytesToHash(common.LeftPadBytes(burnAddrBytes, common.HashLength)),
							},
							Data: common.LeftPadBytes(big.NewInt(amount).Bytes(), common.HashLength),
						},
					},
				}
				return receipt, nil
			},
		}
		broadcaster = &mock2.BroadcasterMock{}
		mgr = NewMgr(rpc, broadcaster, rand.Bytes(sdk.AddrLen), log.TestingLogger(), cdc)
	}
	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := mgr.ProcessDepositConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.True(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmDepositRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("missing attributes", testutils.Func(func(t *testing.T) {
		setup()
		for i := 0; i < len(attributes); i++ {
			// remove one attribute at a time
			wrongAttributes := make([]sdk.Attribute, len(attributes))
			copy(wrongAttributes, attributes)
			wrongAttributes = append(wrongAttributes[:i], wrongAttributes[(i+1):]...)

			err := mgr.ProcessDepositConfirmation(wrongAttributes)
			assert.Error(t, err)
			assert.Len(t, broadcaster.BroadcastCalls(), 0)
		}
	}).Repeat(repeats))

	t.Run("no tx receipt", testutils.Func(func(t *testing.T) {
		setup()
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return nil, fmt.Errorf("error") }

		err := mgr.ProcessDepositConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmDepositRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("no block number", testutils.Func(func(t *testing.T) {
		setup()
		rpc.BlockNumberFunc = func(context.Context) (uint64, error) {
			return 0, fmt.Errorf("error")
		}

		err := mgr.ProcessDepositConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmDepositRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("amount mismatch", testutils.Func(func(t *testing.T) {
		setup()
		for i, attribute := range attributes {
			if attribute.Key == ethTypes.AttributeKeyAmount {
				// have to use index, otherwise this would only change the copy of the attribute, not the one in the slice
				attributes[i].Value = strconv.FormatUint(mathRand.Uint64(), 10)
				break
			}
		}

		err := mgr.ProcessDepositConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmDepositRequest).Confirmed)
	}).Repeat(repeats))
}

func TestMgr_ProccessTokenConfirmation(t *testing.T) {
	var (
		mgr              *Mgr
		attributes       []sdk.Attribute
		rpc              *mock.ClientMock
		broadcaster      *mock2.BroadcasterMock
		gatewayAddrBytes []byte
	)
	setup := func() {
		cdc := testutils.MakeEncodingConfig().Amino
		poll := exported.NewPollMetaWithNonce(ethTypes.ModuleName, rand.StrBetween(5, 20), rand.PosI64(), rand.I64Between(1, 1000))

		gatewayAddrBytes = rand.Bytes(common.AddressLength)
		tokenAddrBytes := rand.Bytes(common.AddressLength)
		blockNumber := rand.PInt64Gen().Where(func(i int64) bool { return i != 0 }).Next() // restrict to int64 so the block number in the receipt doesn't overflow
		confHeight := rand.I64Between(0, blockNumber-1)

		symbol := rand.StrBetween(5, 20)
		attributes = []sdk.Attribute{
			sdk.NewAttribute(ethTypes.AttributeKeyTxID, common.Bytes2Hex(rand.Bytes(common.HashLength))),
			sdk.NewAttribute(ethTypes.AttributeKeyGatewayAddress, common.Bytes2Hex(gatewayAddrBytes)),
			sdk.NewAttribute(ethTypes.AttributeKeyTokenAddress, common.Bytes2Hex(tokenAddrBytes)),
			sdk.NewAttribute(ethTypes.AttributeKeySymbol, symbol),
			sdk.NewAttribute(ethTypes.AttributeKeyConfHeight, strconv.FormatUint(uint64(confHeight), 10)),
			sdk.NewAttribute(ethTypes.AttributeKeyPoll, string(cdc.MustMarshalJSON(poll))),
		}

		rpc = &mock.ClientMock{
			BlockNumberFunc: func(context.Context) (uint64, error) {
				return uint64(blockNumber), nil
			},
			TransactionReceiptFunc: func(context.Context, common.Hash) (*geth.Receipt, error) {
				receipt := &geth.Receipt{
					BlockNumber: big.NewInt(rand.I64Between(0, blockNumber-confHeight)),
					Logs: createTokenLogs(
						symbol,
						common.BytesToAddress(gatewayAddrBytes),
						common.BytesToAddress(tokenAddrBytes),
						ERC20TokenDeploymentSig,
						true,
					),
				}
				return receipt, nil
			},
		}
		broadcaster = &mock2.BroadcasterMock{}
		mgr = NewMgr(rpc, broadcaster, rand.Bytes(sdk.AddrLen), log.TestingLogger(), cdc)
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := mgr.ProcessTokenConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.True(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmTokenRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("missing attributes", testutils.Func(func(t *testing.T) {
		setup()
		for i := 0; i < len(attributes); i++ {
			// remove one attribute at a time
			wrongAttributes := make([]sdk.Attribute, len(attributes))
			copy(wrongAttributes, attributes)
			wrongAttributes = append(wrongAttributes[:i], wrongAttributes[(i+1):]...)

			err := mgr.ProcessTokenConfirmation(wrongAttributes)
			assert.Error(t, err)
			assert.Len(t, broadcaster.BroadcastCalls(), 0)
		}
	}).Repeat(repeats))

	t.Run("no tx receipt", testutils.Func(func(t *testing.T) {
		setup()
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return nil, fmt.Errorf("error") }

		err := mgr.ProcessTokenConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmTokenRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("no block number", testutils.Func(func(t *testing.T) {
		setup()
		rpc.BlockNumberFunc = func(context.Context) (uint64, error) {
			return 0, fmt.Errorf("error")
		}

		err := mgr.ProcessTokenConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmTokenRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("no deploy event", testutils.Func(func(t *testing.T) {
		setup()
		receipt, _ := rpc.TransactionReceipt(context.Background(), common.Hash{})
		var correctLogIdx int
		for i, l := range receipt.Logs {
			if l.Address == common.BytesToAddress(gatewayAddrBytes) {
				correctLogIdx = i
				break
			}
		}
		// remove the deploy event
		receipt.Logs = append(receipt.Logs[:correctLogIdx], receipt.Logs[correctLogIdx+1:]...)
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return receipt, nil }

		err := mgr.ProcessTokenConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmTokenRequest).Confirmed)
	}).Repeat(repeats))

	t.Run("wrong deploy event", testutils.Func(func(t *testing.T) {
		setup()
		receipt, _ := rpc.TransactionReceipt(context.Background(), common.Hash{})
		for _, l := range receipt.Logs {
			if l.Address == common.BytesToAddress(gatewayAddrBytes) {
				l.Data = rand.Bytes(int(rand.I64Between(0, 1000)))
				break
			}
		}
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return receipt, nil }

		err := mgr.ProcessTokenConfirmation(attributes)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.False(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*ethTypes.VoteConfirmTokenRequest).Confirmed)
	}).Repeat(repeats))
}

func createTokenLogs(denom string, gateway, tokenAddr common.Address, deploySig common.Hash, hasCorrectLog bool) []*geth.Log {
	numLogs := rand.I64Between(1, 100)
	correctPos := rand.I64Between(0, numLogs)
	var logs []*geth.Log

	for i := int64(0); i < numLogs; i++ {
		stringType, err := abi.NewType("string", "string", nil)
		if err != nil {
			panic(err)
		}
		addressType, err := abi.NewType("address", "address", nil)
		if err != nil {
			panic(err)
		}
		args := abi.Arguments{{Type: stringType}, {Type: addressType}}

		switch {
		case hasCorrectLog && i == correctPos:
			data, err := args.Pack(denom, tokenAddr)
			if err != nil {
				panic(err)
			}
			logs = append(logs, &geth.Log{Address: gateway, Data: data, Topics: []common.Hash{deploySig}})
		default:
			randDenom := rand.StrBetween(5, 20)
			randAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))
			randData, err := args.Pack(randDenom, randAddr)
			if err != nil {
				panic(err)
			}
			logs = append(logs, &geth.Log{
				Address: common.BytesToAddress(rand.Bytes(common.AddressLength)),
				Data:    randData,
				Topics:  []common.Hash{common.BytesToHash(rand.Bytes(common.HashLength))},
			})
		}

	}

	return logs
}

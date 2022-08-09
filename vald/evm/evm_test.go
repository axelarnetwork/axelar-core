package evm

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/crypto/sha3"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	mock2 "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmRpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

// testHasher is the helper tool for transaction/receipt list hashing.
// The original hasher is trie, in order to get rid of import cycle,
// use the testing hasher instead.
type testHasher struct {
	hasher hash.Hash
}

func newHasher() *testHasher {
	return &testHasher{hasher: sha3.NewLegacyKeccak256()}
}

func (h *testHasher) Reset() {
	h.hasher.Reset()
}

func (h *testHasher) Update(key, val []byte) {
	h.hasher.Write(key)
	h.hasher.Write(val)
}

func (h *testHasher) Hash() common.Hash {
	return common.BytesToHash(h.hasher.Sum(nil))
}

func TestDecodeEventTokenSent(t *testing.T) {
	log := &geth.Log{
		Topics: []common.Hash{
			common.HexToHash("0x651d93f66c4329630e8d0f62488eff599e3be484da587335e8dc0fcf46062726"),
			common.HexToHash("0x00000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb8"),
		},
		Data: common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000989680000000000000000000000000000000000000000000000000000000000000000a657468657265756d2d3200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a30783538656134313033656439353564434262646338613066456261626133393542366534346431354600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f657468657265756d2d312d7561786c0000000000000000000000000000000000"),
	}

	expected := evmTypes.EventTokenSent{
		Sender:             evmTypes.Address(common.HexToAddress("0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8")),
		DestinationChain:   "ethereum-2",
		DestinationAddress: "0x58ea4103ed955dCBbdc8a0fEbaba395B6e44d15F",
		Symbol:             "ethereum-1-uaxl",
		Amount:             sdk.NewUint(10000000),
	}
	actual, err := decodeEventTokenSent(log)

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

	expected := evmTypes.EventContractCall{
		Sender:           evmTypes.Address(common.HexToAddress("0xD48E199950589A4336E4dc43bd2C72Ba0C0baA86")),
		DestinationChain: "ethereum-2",
		ContractAddress:  "0xb9845f9247a85Ee592273a79605f34E8607d7e75",
		PayloadHash:      evmTypes.Hash(common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e")),
	}
	actual, err := decodeEventContractCall(log)

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

	expected := evmTypes.EventContractCallWithToken{
		Sender:           evmTypes.Address(common.HexToAddress("0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8")),
		DestinationChain: "ethereum",
		ContractAddress:  "0x76a06043391712bE39A343d1f43165854fCF6De3",
		PayloadHash:      evmTypes.Hash(common.HexToHash("0x9fcef596d62dca8e51b6ba3414901947c0e6821d4483b2f3327ce87c2d4e662e")),
		Symbol:           "uaxl",
		Amount:           sdk.NewUint(10000000),
	}
	actual, err := decodeEventContractCallWithToken(log)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestDecodeTokenDeployEvent_CorrectData(t *testing.T) {
	axelarGateway := common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA")
	tokenDeploySig := ERC20TokenDeploymentSig
	expectedAddr := common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D")
	expectedSymbol := "XPTO"
	data := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000e7481ecb61f9c84b91c03414f3d5d48e5436045d00000000000000000000000000000000000000000000000000000000000000045850544f00000000000000000000000000000000000000000000000000000000")

	l := &geth.Log{Address: axelarGateway, Data: data, Topics: []common.Hash{tokenDeploySig}}

	tokenDeployed, err := decodeERC20TokenDeploymentEvent(l)
	assert.NoError(t, err)
	assert.Equal(t, expectedSymbol, tokenDeployed.Symbol)
	assert.Equal(t, evmTypes.Address(expectedAddr), tokenDeployed.TokenAddress)
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

	_, err := decodeERC20TransferEvent(&l)
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

	_, err := decodeERC20TransferEvent(&l)

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

	transfer, err := decodeERC20TransferEvent(&l)

	assert.NoError(t, err)
	assert.Equal(t, evmTypes.Address(expectedTo), transfer.To)
	assert.Equal(t, expectedAmount, transfer.Amount)
}

func TestMgr_validate(t *testing.T) {
	t.Run("should work for moonbeam", testutils.Func(func(t *testing.T) {
		mgr := Mgr{logger: log.TestingLogger()}

		latestFinalizedBlockHash := common.BytesToHash(rand.Bytes(common.HashLength))
		latestFinalizedBlockNumber := rand.I64Between(1000, 10000)
		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(rand.PosI64()), uint64(rand.PosI64()), big.NewInt(rand.PosI64()), rand.Bytes(int(rand.I64Between(100, 1000))))
		receipt := &geth.Receipt{
			BlockNumber: big.NewInt(latestFinalizedBlockNumber - rand.I64Between(1, 100)),
			TxHash:      tx.Hash(),
			Status:      1,
		}

		rpc := &mock.MoonbeamClientMock{
			TransactionByHashFunc: func(_ context.Context, hash common.Hash) (*geth.Transaction, bool, error) {
				if bytes.Equal(hash.Bytes(), tx.Hash().Bytes()) {
					return tx, false, nil
				}

				return nil, false, fmt.Errorf("not found")
			},
			TransactionReceiptFunc: func(_ context.Context, txHash common.Hash) (*geth.Receipt, error) {
				if bytes.Equal(txHash.Bytes(), tx.Hash().Bytes()) {
					return receipt, nil
				}

				return nil, fmt.Errorf("not found")
			},
			ChainGetFinalizedHeadFunc: func(_ context.Context) (common.Hash, error) { return latestFinalizedBlockHash, nil },
			ChainGetHeaderFunc: func(ctx context.Context, hash common.Hash) (*evmRpc.MoonbeamHeader, error) {
				if bytes.Equal(hash.Bytes(), latestFinalizedBlockHash.Bytes()) {
					blockNumber := hexutil.Big(*big.NewInt(latestFinalizedBlockNumber))

					return &evmRpc.MoonbeamHeader{Number: &blockNumber}, nil
				}

				return nil, fmt.Errorf("not found")
			},
			BlockByNumberFunc: func(ctx context.Context, number *big.Int) (*geth.Block, error) {
				if number.Cmp(receipt.BlockNumber) == 0 {
					return geth.NewBlock(&geth.Header{}, []*geth.Transaction{tx}, []*geth.Header{}, []*geth.Receipt{receipt}, newHasher()), nil
				}

				return nil, fmt.Errorf("not found")
			},
		}

		isFinalized := mgr.validate(rpc, tx.Hash(), 0, func(_ *geth.Transaction, _ *geth.Receipt) bool { return true })
		assert.True(t, isFinalized)
	}))
}

func TestMgr_ProccessDepositConfirmation(t *testing.T) {
	var (
		mgr            *Mgr
		event          *evmTypes.ConfirmDepositStarted
		rpc            *mock.ClientMock
		broadcaster    *mock2.BroadcasterMock
		encodingConfig params.EncodingConfig
		valAddr        sdk.ValAddress
	)
	setup := func() {
		encodingConfig = app.MakeEncodingConfig()
		cdc := encodingConfig.Amino
		pollID := vote.PollID(rand.I64Between(10, 100))

		burnAddrBytes := rand.Bytes(common.AddressLength)
		tokenAddrBytes := rand.Bytes(common.AddressLength)
		blockNumber := rand.PInt64Gen().Where(func(i int64) bool { return i != 0 }).Next() // restrict to int64 so the block number in the receipt doesn't overflow
		confHeight := rand.I64Between(0, blockNumber-1)
		amount := rand.PosI64() // restrict to int64 so the amount in the receipt doesn't overflow
		valAddr = rand.ValAddr()
		event = &types.ConfirmDepositStarted{
			TxID:               types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Chain:              "Ethereum",
			DepositAddress:     types.Address(common.BytesToAddress(burnAddrBytes)),
			TokenAddress:       types.Address(common.BytesToAddress(tokenAddrBytes)),
			ConfirmationHeight: uint64(confHeight),
			PollParticipants: vote.PollParticipants{
				PollID:       pollID,
				Participants: []sdk.ValAddress{valAddr},
			},
		}

		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		receipt := &geth.Receipt{
			TxHash:      tx.Hash(),
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
				/* an ERC20 transfer with 0 amount */
				{
					Address: common.BytesToAddress(tokenAddrBytes),
					Topics: []common.Hash{
						ERC20TransferSig,
						common.BytesToHash(common.LeftPadBytes(rand.Bytes(common.AddressLength), common.HashLength)),
						common.BytesToHash(common.LeftPadBytes(burnAddrBytes, common.HashLength)),
					},
					Data: common.LeftPadBytes(big.NewInt(0).Bytes(), common.HashLength),
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
			Status: 1,
		}
		rpc = &mock.ClientMock{
			BlockByNumberFunc: func(ctx context.Context, number *big.Int) (*geth.Block, error) {
				return geth.NewBlock(&geth.Header{}, []*geth.Transaction{tx}, []*geth.Header{}, []*geth.Receipt{receipt}, newHasher()), nil
			},
			BlockNumberFunc: func(context.Context) (uint64, error) {
				return uint64(blockNumber), nil
			},
			TransactionByHashFunc: func(ctx context.Context, hash common.Hash) (*geth.Transaction, bool, error) {
				return &geth.Transaction{}, false, nil
			},
			TransactionReceiptFunc: func(context.Context, common.Hash) (*geth.Receipt, error) {
				return receipt, nil
			},
		}
		broadcaster = &mock2.BroadcasterMock{
			BroadcastFunc: func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmRpc.Client)
		evmMap["ethereum"] = rpc
		mgr = NewMgr(evmMap, client.Context{}, broadcaster, log.TestingLogger(), cdc, valAddr)
	}
	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := mgr.ProcessDepositConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 1)
	}).Repeat(repeats))

	t.Run("no tx receipt", testutils.Func(func(t *testing.T) {
		setup()
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return nil, fmt.Errorf("error") }

		err := mgr.ProcessDepositConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).Repeat(repeats))

	t.Run("no block number", testutils.Func(func(t *testing.T) {
		setup()
		rpc.BlockNumberFunc = func(context.Context) (uint64, error) {
			return 0, fmt.Errorf("error")
		}

		err := mgr.ProcessDepositConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).Repeat(repeats))
}

func TestMgr_ProccessTokenConfirmation(t *testing.T) {
	var (
		mgr              *Mgr
		event            *evmTypes.ConfirmTokenStarted
		rpc              *mock.ClientMock
		broadcaster      *mock2.BroadcasterMock
		gatewayAddrBytes []byte
		encodingConfig   params.EncodingConfig
		valAddr          sdk.ValAddress
	)
	setup := func() {
		encodingConfig = app.MakeEncodingConfig()
		cdc := encodingConfig.Amino
		pollID := vote.PollID(rand.I64Between(10, 100))

		gatewayAddrBytes = rand.Bytes(common.AddressLength)
		tokenAddrBytes := rand.Bytes(common.AddressLength)
		blockNumber := rand.PInt64Gen().Where(func(i int64) bool { return i != 0 }).Next() // restrict to int64 so the block number in the receipt doesn't overflow
		confHeight := rand.I64Between(0, blockNumber-1)

		symbol := rand.Denom(5, 20)
		valAddr = rand.ValAddr()
		event = &types.ConfirmTokenStarted{
			TxID:               types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Chain:              "Ethereum",
			GatewayAddress:     types.Address(common.BytesToAddress(gatewayAddrBytes)),
			TokenAddress:       types.Address(common.BytesToAddress(tokenAddrBytes)),
			TokenDetails:       types.TokenDetails{Symbol: symbol},
			ConfirmationHeight: uint64(confHeight),
			PollParticipants: vote.PollParticipants{
				PollID:       pollID,
				Participants: []sdk.ValAddress{valAddr},
			},
		}

		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		receipt := &geth.Receipt{
			TxHash:      tx.Hash(),
			BlockNumber: big.NewInt(rand.I64Between(0, blockNumber-confHeight)),
			Logs: createTokenLogs(
				symbol,
				common.BytesToAddress(gatewayAddrBytes),
				common.BytesToAddress(tokenAddrBytes),
				ERC20TokenDeploymentSig,
				true,
			),
			Status: 1,
		}
		rpc = &mock.ClientMock{
			BlockByNumberFunc: func(ctx context.Context, number *big.Int) (*geth.Block, error) {
				return geth.NewBlock(&geth.Header{}, []*geth.Transaction{tx}, []*geth.Header{}, []*geth.Receipt{receipt}, newHasher()), nil
			},
			BlockNumberFunc: func(context.Context) (uint64, error) {
				return uint64(blockNumber), nil
			},
			TransactionByHashFunc: func(ctx context.Context, hash common.Hash) (*geth.Transaction, bool, error) {
				return &geth.Transaction{}, false, nil
			},
			TransactionReceiptFunc: func(context.Context, common.Hash) (*geth.Receipt, error) {
				return receipt, nil
			},
		}
		broadcaster = &mock2.BroadcasterMock{
			BroadcastFunc: func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmRpc.Client)
		evmMap["ethereum"] = rpc
		mgr = NewMgr(evmMap, client.Context{}, broadcaster, log.TestingLogger(), cdc, valAddr)
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 1)
	}).Repeat(repeats))

	t.Run("no tx receipt", testutils.Func(func(t *testing.T) {
		setup()
		rpc.TransactionReceiptFunc = func(context.Context, common.Hash) (*geth.Receipt, error) { return nil, fmt.Errorf("error") }

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).Repeat(repeats))

	t.Run("no block number", testutils.Func(func(t *testing.T) {
		setup()
		rpc.BlockNumberFunc = func(context.Context) (uint64, error) {
			return 0, fmt.Errorf("error")
		}

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
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

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)

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

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*voteTypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).Repeat(repeats))
}

func TestMgr_ProcessTransferKeyConfirmation(t *testing.T) {
	var (
		mgr            *Mgr
		event          *types.ConfirmKeyTransferStarted
		rpc            *mock.ClientMock
		broadcaster    *mock2.BroadcasterMock
		txID           types.Hash
		gatewayAddress types.Address
		pollID         vote.PollID
		txReceipt      *geth.Receipt
		valAddr        sdk.ValAddress
	)

	givenEvmMgr := Given("EVM mgr", func() {
		rpc = &mock.ClientMock{}
		broadcaster = &mock2.BroadcasterMock{
			BroadcastFunc: func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmRpc.Client)
		evmMap["ethereum"] = rpc
		valAddr = rand.ValAddr()
		mgr = NewMgr(evmMap, client.Context{}, broadcaster, log.TestingLogger(), app.MakeEncodingConfig().Amino, valAddr)
	})

	givenTxReceiptAndBlockAreFound := Given("tx receipt and block can be found", func() {
		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		blockNumber := uint64(rand.I64Between(1, 1000))

		txID = types.Hash(tx.Hash())
		txReceipt = &geth.Receipt{
			TxHash:      common.Hash(txID),
			BlockNumber: big.NewInt(rand.I64Between(0, int64(blockNumber-evmTypes.DefaultParams()[0].ConfirmationHeight+2))),
			Logs:        []*geth.Log{},
			Status:      1,
		}
		block := *geth.NewBlock(&geth.Header{Number: big.NewInt(int64(blockNumber))}, []*geth.Transaction{tx}, []*geth.Header{}, []*geth.Receipt{txReceipt}, newHasher())

		rpc.TransactionByHashFunc = func(_ context.Context, hash common.Hash) (*geth.Transaction, bool, error) {
			return &geth.Transaction{}, false, nil
		}
		rpc.TransactionReceiptFunc = func(ctx context.Context, txHash common.Hash) (*geth.Receipt, error) {
			if txHash == common.Hash(txID) {
				return txReceipt, nil
			}

			return nil, fmt.Errorf("not found")
		}
		rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) { return blockNumber, nil }
		rpc.BlockByNumberFunc = func(ctx context.Context, number *big.Int) (*geth.Block, error) {
			if number.Cmp(txReceipt.BlockNumber) == 0 {
				return &block, nil
			}

			return nil, fmt.Errorf("not found")
		}
	})

	givenEventConfirmKeyTransfer := Given("event confirm key transfer", func() {
		gatewayAddress = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
		pollID = vote.PollID(rand.PosI64())
		event = types.NewConfirmKeyTransferStarted(
			exported.Ethereum.Name,
			txID,
			gatewayAddress,
			evmTypes.DefaultParams()[0].ConfirmationHeight,
			vote.PollParticipants{
				PollID:       pollID,
				Participants: []sdk.ValAddress{valAddr},
			},
		)
	})

	assertAndGetVoteEvents := func(t *testing.T, isEmpty bool) *evmTypes.VoteEvents {
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)

		voteEvents := broadcaster.BroadcastCalls()[0].Msgs[0].(*voteTypes.VoteRequest).Vote.GetCachedValue().(*evmTypes.VoteEvents)
		if isEmpty {
			assert.Empty(t, voteEvents.Events)
		} else {
			assert.Len(t, voteEvents.Events, 1)
		}

		return voteEvents
	}

	thenShouldVoteNoEvent := Then("should vote no event", func(t *testing.T) {
		err := mgr.ProcessTransferKeyConfirmation(event)
		assert.NoError(t, err)

		assertAndGetVoteEvents(t, true)
	})

	givenEvmMgr.
		Given2(givenTxReceiptAndBlockAreFound).
		Given2(givenEventConfirmKeyTransfer).
		Branch(
			When("is not operatorship transferred event", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{common.BytesToHash(rand.Bytes(common.HashLength))},
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is not emitted from the gateway", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.BytesToAddress(rand.Bytes(common.AddressLength)),
					Topics:  []common.Hash{MultisigTransferOperatorshipSig},
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is invalid operatorship transferred event", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{MultisigTransferOperatorshipSig},
					Data:    rand.Bytes(int(rand.I64Between(0, 1000))),
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is valid operatorship transferred event", func() {
				newOperatorsData := common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000800000000000000000000000019cc2044857d23129a29f763d0338da837ce35f60000000000000000000000002ab6fa7de5e9e9423125a4246e4de1b9c755607400000000000000000000000037cc4b7e8f9f505ca8126db8a9d070566ed5dae70000000000000000000000003e56f0d4497ac44993d9ea272d4707f8be6b42a6000000000000000000000000462b96f617d5d92f63f9949c6f4626623ea73fa400000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb80000000000000000000000009e77c30badbbc412a0c20c6ce43b671c6f103434000000000000000000000000c1c0c8d2131cc866834c6382096eadfef1af2f52000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005")
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{})
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{MultisigTransferOperatorshipSig},
					Data:    funcs.Must(abi.Arguments{{Type: funcs.Must(abi.NewType("bytes", "bytes", nil))}}.Pack(newOperatorsData)),
				})
			}).
				Then("should vote for the correct event", func(t *testing.T) {
					err := mgr.ProcessTransferKeyConfirmation(event)
					assert.NoError(t, err)

					actual := assertAndGetVoteEvents(t, false)
					assert.Equal(t, exported.Ethereum.Name, actual.Chain)
					assert.Equal(t, exported.Ethereum.Name, actual.Events[0].Chain)
					assert.Equal(t, txID, actual.Events[0].TxID)
					assert.EqualValues(t, 1, actual.Events[0].Index)
					assert.IsType(t, &types.Event_MultisigOperatorshipTransferred{}, actual.Events[0].Event)

					actualEvent := actual.Events[0].Event.(*types.Event_MultisigOperatorshipTransferred)
					assert.Len(t, actualEvent.MultisigOperatorshipTransferred.NewOperators, 8)
					assert.Len(t, actualEvent.MultisigOperatorshipTransferred.NewWeights, 8)
					assert.EqualValues(t, 30, actualEvent.MultisigOperatorshipTransferred.NewThreshold.BigInt().Int64())
				}),
		).
		Run(t, 5)
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

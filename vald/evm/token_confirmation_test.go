package evm_test

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	broadcastmock "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmrpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
)

func TestMgr_ProccessTokenConfirmation(t *testing.T) {
	var (
		mgr              *evm.Mgr
		event            *types.ConfirmTokenStarted
		rpc              *mock.ClientMock
		broadcaster      *broadcastmock.BroadcasterMock
		gatewayAddrBytes []byte
		valAddr          sdk.ValAddress
		receipt          *geth.Receipt
	)

	setup := func() {
		pollID := vote.PollID(rand.I64Between(10, 100))

		gatewayAddrBytes = rand.Bytes(common.AddressLength)
		tokenAddrBytes := rand.Bytes(common.AddressLength)
		blockNumber := rand.PInt64Gen().Where(func(i int64) bool { return i != 0 }).Next() // restrict to int64 so the block number in the receipt doesn't overflow
		confHeight := rand.I64Between(0, blockNumber-1)

		symbol := rand.Denom(5, 20)
		valAddr = rand.ValAddr()

		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		receipt = &geth.Receipt{
			TxHash:      tx.Hash(),
			BlockNumber: big.NewInt(rand.I64Between(0, blockNumber-confHeight)),
			Logs: createTokenLogs(
				symbol,
				common.BytesToAddress(gatewayAddrBytes),
				common.BytesToAddress(tokenAddrBytes),
				evm.ERC20TokenDeploymentSig,
				true,
			),
			Status: 1,
		}
		event = &types.ConfirmTokenStarted{
			TxID:               types.Hash(receipt.TxHash),
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

		rpc = &mock.ClientMock{
			HeaderByNumberFunc: func(ctx context.Context, number *big.Int) (*evmrpc.Header, error) {
				return &evmrpc.Header{Transactions: []common.Hash{receipt.TxHash}}, nil
			},
			TransactionReceiptsFunc: func(ctx context.Context, txHashes []common.Hash) ([]evmrpc.TxReceiptResult, error) {
				return slices.Map(txHashes, func(txHash common.Hash) evmrpc.TxReceiptResult {
					if bytes.Equal(txHash.Bytes(), receipt.TxHash.Bytes()) {
						return evmrpc.TxReceiptResult(results.FromOk(*receipt))
					}

					return evmrpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
				}), nil
			},
			LatestFinalizedBlockNumberFunc: func(ctx context.Context, confirmations uint64) (*big.Int, error) {
				return receipt.BlockNumber, nil
			},
		}
		broadcaster = &broadcastmock.BroadcasterMock{
			BroadcastFunc: func(context.Context, ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmrpc.Client)
		evmMap["ethereum"] = rpc
		mgr = evm.NewMgr(evmMap, broadcaster, valAddr, rand.AccAddr(), &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(_ nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(_ nexus.ChainName, _ *big.Int) {},
		})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 1)
	}).
		Repeat(repeats))

	t.Run("no tx receipt", testutils.Func(func(t *testing.T) {
		setup()
		rpc.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmrpc.TxReceiptResult, error) {
			return slices.Map(txHashes, func(txHash common.Hash) evmrpc.TxReceiptResult {
				return evmrpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
			}), nil
		}

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).
		Repeat(repeats))

	t.Run("no deploy event", testutils.Func(func(t *testing.T) {
		setup()

		var correctLogIdx int
		for i, l := range receipt.Logs {
			if l.Address == common.BytesToAddress(gatewayAddrBytes) {
				correctLogIdx = i
				break
			}
		}
		// remove the deploy event
		receipt.Logs = append(receipt.Logs[:correctLogIdx], receipt.Logs[correctLogIdx+1:]...)

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)

	}).
		Repeat(repeats))

	t.Run("wrong deploy event", testutils.Func(func(t *testing.T) {
		setup()

		for _, l := range receipt.Logs {
			if l.Address == common.BytesToAddress(gatewayAddrBytes) {
				l.Data = rand.Bytes(int(rand.I64Between(0, 1000)))
				break
			}
		}

		err := mgr.ProcessTokenConfirmation(event)

		assert.NoError(t, err)
		assert.Len(t, broadcaster.BroadcastCalls(), 1)

		msg := broadcaster.BroadcastCalls()[0].Msgs[0]
		actualVoteEvents := msg.(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Equal(t, nexus.ChainName("Ethereum"), actualVoteEvents.Chain)
		assert.Len(t, actualVoteEvents.Events, 0)
	}).
		Repeat(repeats))
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

func TestMgr_ProcessTokenConfirmationNoTopicsNotPanics(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	receipt := geth.Receipt{
		Logs:        []*geth.Log{{Topics: make([]common.Hash, 0)}},
		BlockNumber: big.NewInt(1),
		Status:      geth.ReceiptStatusSuccessful,
	}
	rpcClient := &mock.ClientMock{TransactionReceiptsFunc: func(_ context.Context, _ []common.Hash) ([]evmrpc.TxReceiptResult, error) {
		return []evmrpc.TxReceiptResult{evmrpc.TxReceiptResult(results.FromOk(receipt))}, nil
	}}
	cache := &evmmock.LatestFinalizedBlockCacheMock{GetFunc: func(chain nexus.ChainName) *big.Int {
		return big.NewInt(100)
	}}

	broadcaster := &broadcastmock.BroadcasterMock{BroadcastFunc: func(_ context.Context, _ ...sdk.Msg) (*sdk.TxResponse, error) {
		return nil, nil
	}}

	valAddr := rand.ValAddr()
	mgr := evm.NewMgr(map[string]evmrpc.Client{chain.String(): rpcClient}, broadcaster, valAddr, rand.AccAddr(), cache)

	assert.NotPanics(t, func() {
		mgr.ProcessTokenConfirmation(&types.ConfirmTokenStarted{TxID: types.Hash{1},
			PollParticipants: vote.PollParticipants{PollID: 10, Participants: []sdk.ValAddress{valAddr}},
			Chain:            chain,
		})
	})
}

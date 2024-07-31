package evm_test

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	broadcastmock "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmrpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_ProcessTransferKeyConfirmation(t *testing.T) {
	var (
		mgr            *evm.Mgr
		event          *types.ConfirmKeyTransferStarted
		rpc            *mock.ClientMock
		broadcaster    *broadcastmock.BroadcasterMock
		txID           types.Hash
		gatewayAddress types.Address
		pollID         vote.PollID
		txReceipt      *geth.Receipt
		valAddr        sdk.ValAddress
	)

	givenEvmMgr := Given("EVM mgr", func() {
		rpc = &mock.ClientMock{}
		broadcaster = &broadcastmock.BroadcasterMock{
			BroadcastFunc: func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmrpc.Client)
		evmMap["ethereum"] = rpc
		valAddr = rand.ValAddr()
		mgr = evm.NewMgr(evmMap, broadcaster, valAddr, rand.AccAddr(), &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(_ nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(_ nexus.ChainName, _ *big.Int) {},
		})
	})

	givenTxReceiptAndBlockAreFound := Given("tx receipt and block can be found", func() {
		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		blockNumber := uint64(rand.I64Between(1, 1000))

		txID = types.Hash(tx.Hash())
		txReceipt = &geth.Receipt{
			TxHash:      common.Hash(txID),
			BlockNumber: big.NewInt(rand.I64Between(0, int64(blockNumber-types.DefaultParams()[0].ConfirmationHeight+2))),
			Logs:        []*geth.Log{},
			Status:      1,
		}

		rpc.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmrpc.TxReceiptResult, error) {
			return slices.Map(txHashes, func(hash common.Hash) evmrpc.TxReceiptResult {
				if bytes.Equal(hash.Bytes(), txID.Bytes()) {
					return evmrpc.TxReceiptResult(results.FromOk(*txReceipt))
				}

				return evmrpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
			}), nil
		}
		rpc.HeaderByNumberFunc = func(ctx context.Context, number *big.Int) (*evmrpc.Header, error) {
			if number.Cmp(txReceipt.BlockNumber) == 0 {
				number := hexutil.Big(*big.NewInt(int64(blockNumber)))
				return &evmrpc.Header{Number: &number, Transactions: []common.Hash{txReceipt.TxHash}}, nil
			}

			return nil, fmt.Errorf("not found")
		}
		rpc.LatestFinalizedBlockNumberFunc = func(ctx context.Context, confirmations uint64) (*big.Int, error) {
			return txReceipt.BlockNumber, nil
		}
	})

	givenEventConfirmKeyTransfer := Given("event confirm key transfer", func() {
		gatewayAddress = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
		pollID = vote.PollID(rand.PosI64())
		event = types.NewConfirmKeyTransferStarted(
			exported.Ethereum.Name,
			txID,
			gatewayAddress,
			types.DefaultParams()[0].ConfirmationHeight,
			vote.PollParticipants{
				PollID:       pollID,
				Participants: []sdk.ValAddress{valAddr},
			},
		)
	})

	assertAndGetVoteEvents := func(t *testing.T, isEmpty bool) *types.VoteEvents {
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)

		voteEvents := broadcaster.BroadcastCalls()[0].Msgs[0].(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
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
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is invalid operatorship transferred event", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
					Data:    rand.Bytes(int(rand.I64Between(0, 1000))),
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is valid operatorship transferred event", func() {
				newOperatorsData := common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000800000000000000000000000019cc2044857d23129a29f763d0338da837ce35f60000000000000000000000002ab6fa7de5e9e9423125a4246e4de1b9c755607400000000000000000000000037cc4b7e8f9f505ca8126db8a9d070566ed5dae70000000000000000000000003e56f0d4497ac44993d9ea272d4707f8be6b42a6000000000000000000000000462b96f617d5d92f63f9949c6f4626623ea73fa400000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb80000000000000000000000009e77c30badbbc412a0c20c6ce43b671c6f103434000000000000000000000000c1c0c8d2131cc866834c6382096eadfef1af2f52000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005")
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{})
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
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

func TestMgr_ProcessTransferKeyConfirmationNoTopicsNotPanics(t *testing.T) {
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
		mgr.ProcessTransferKeyConfirmation(&types.ConfirmKeyTransferStarted{TxID: types.Hash{1},
			PollParticipants: vote.PollParticipants{PollID: 10, Participants: []sdk.ValAddress{valAddr}},
			Chain:            chain,
		})
	})
}

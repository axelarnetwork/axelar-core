package evm_test

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmRpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_GetTxReceiptIfFinalized(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(rand.PosI64()), uint64(rand.PosI64()), big.NewInt(rand.PosI64()), rand.Bytes(int(rand.I64Between(100, 1000))))

	var (
		mgr                        *evm.Mgr
		rpcClient                  *mock.ClientMock
		cache                      *evmmock.LatestFinalizedBlockCacheMock
		confHeight                 uint64
		latestFinalizedBlockNumber uint64
	)

	givenMgr := Given("evm mgr", func() {
		rpcClient = &mock.ClientMock{}
		cache = &evmmock.LatestFinalizedBlockCacheMock{}
		confHeight = uint64(rand.I64Between(1, 50))
		latestFinalizedBlockNumber = uint64(rand.I64Between(1000, 10000))

		mgr = evm.NewMgr(map[string]evmRpc.Client{chain.String(): rpcClient}, nil, rand.ValAddr(), rand.AccAddr(), cache)
	})

	givenMgr.
		When("the rpc client determines that the tx failed", func() {
			receipt := &geth.Receipt{
				BlockNumber: big.NewInt(int64(latestFinalizedBlockNumber) - rand.I64Between(1, 100)),
				TxHash:      tx.Hash(),
				Status:      geth.ReceiptStatusFailed,
			}

			cache.GetFunc = func(_ nexus.ChainName) *big.Int {
				return receipt.BlockNumber
			}
			rpcClient.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
				return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
					if bytes.Equal(hash.Bytes(), tx.Hash().Bytes()) {
						return evmRpc.TxReceiptResult(results.FromOk(*receipt))
					}

					return evmRpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
				}), nil
			}
		}).
		Then("tx is considered failed", func(t *testing.T) {
			txReceipt, err := mgr.GetTxReceiptIfFinalized(chain, tx.Hash(), confHeight)

			assert.NoError(t, err)
			assert.Equal(t, txReceipt.Err(), evm.ErrTxFailed)
		}).
		Run(t)

	givenMgr.
		When("the latest finalized block cache does not have the result", func() {
			cache.GetFunc = func(_ nexus.ChainName) *big.Int {
				return big.NewInt(0)
			}
			cache.SetFunc = func(_ nexus.ChainName, blockNumber *big.Int) {}
		}).
		When("the rpc client determines that the tx is finalized", func() {
			receipt := &geth.Receipt{
				BlockNumber: big.NewInt(int64(latestFinalizedBlockNumber) - rand.I64Between(1, 100)),
				TxHash:      tx.Hash(),
				Status:      geth.ReceiptStatusSuccessful,
			}

			rpcClient.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
				return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
					if bytes.Equal(hash.Bytes(), tx.Hash().Bytes()) {
						return evmRpc.TxReceiptResult(results.FromOk(*receipt))
					}

					return evmRpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
				}), nil
			}
			rpcClient.HeaderByNumberFunc = func(ctx context.Context, number *big.Int) (*evmRpc.Header, error) {
				if number.Cmp(receipt.BlockNumber) == 0 {
					return &evmRpc.Header{Transactions: []common.Hash{receipt.TxHash}}, nil
				}

				return nil, fmt.Errorf("not found")
			}
			rpcClient.LatestFinalizedBlockNumberFunc = func(ctx context.Context, confirmations uint64) (*big.Int, error) {
				return big.NewInt(int64(latestFinalizedBlockNumber)), nil
			}
		}).
		Then("tx is considered finalized", func(t *testing.T) {
			txReceipt, err := mgr.GetTxReceiptIfFinalized(chain, tx.Hash(), confHeight)

			assert.NoError(t, err)
			assert.NoError(t, txReceipt.Err())
			assert.NotNil(t, txReceipt.Ok())
		}).
		Run(t, 5)

	givenMgr.
		When("the latest finalized block cache has the result", func() {
			cache.GetFunc = func(_ nexus.ChainName) *big.Int {
				return big.NewInt(int64(latestFinalizedBlockNumber))
			}
		}).
		When("the rpc client can find the tx receipt", func() {
			receipt := &geth.Receipt{
				BlockNumber: big.NewInt(int64(latestFinalizedBlockNumber) - rand.I64Between(1, 100)),
				TxHash:      tx.Hash(),
				Status:      geth.ReceiptStatusSuccessful,
			}

			rpcClient.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
				return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
					if bytes.Equal(hash.Bytes(), tx.Hash().Bytes()) {
						return evmRpc.TxReceiptResult(results.FromOk(*receipt))
					}

					return evmRpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
				}), nil
			}
			rpcClient.HeaderByNumberFunc = func(ctx context.Context, number *big.Int) (*evmRpc.Header, error) {
				if number.Cmp(receipt.BlockNumber) == 0 {
					return &evmRpc.Header{Transactions: []common.Hash{receipt.TxHash}}, nil
				}

				return nil, fmt.Errorf("not found")
			}
		}).
		Then("tx is considered finalized", func(t *testing.T) {
			txReceipt, err := mgr.GetTxReceiptIfFinalized(chain, tx.Hash(), confHeight)

			assert.NoError(t, err)
			assert.NoError(t, txReceipt.Err())
			assert.NotNil(t, txReceipt.Ok())
		}).
		Run(t, 5)
}

func TestMgr_GetTxReceiptsIfFinalized(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	txHashes := slices.Expand2(func() common.Hash { return common.BytesToHash(rand.Bytes(common.HashLength)) }, 100)

	var (
		mgr                        *evm.Mgr
		confHeight                 uint64
		latestFinalizedBlockNumber int64
		evmClient                  *mock.ClientMock
		cache                      *evmmock.LatestFinalizedBlockCacheMock
	)

	givenMgr := Given("evm mgr", func() {
		evmClient = &mock.ClientMock{
			LatestFinalizedBlockNumberFunc: func(context.Context, uint64) (*big.Int, error) {
				return big.NewInt(latestFinalizedBlockNumber), nil
			},
		}
		cache = &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(chain nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(nexus.ChainName, *big.Int) {},
		}
		mgr = evm.NewMgr(map[string]evmRpc.Client{chain.String(): evmClient}, nil, rand.ValAddr(), rand.AccAddr(), cache)
	})

	confHeight = uint64(rand.I64Between(1, 50))

	givenMgr.
		Branch(
			When("transactions failed", func() {
				latestFinalizedBlockNumber = rand.I64Between(1000, 10000)

				evmClient.TransactionReceiptsFunc = func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
					return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
						return evmRpc.TxReceiptResult(results.FromOk(geth.Receipt{
							BlockNumber: big.NewInt(latestFinalizedBlockNumber - rand.I64Between(1, 100)),
							TxHash:      hash,
							Status:      geth.ReceiptStatusFailed,
						}))
					}), nil
				}
			}).
				Then("should not retrieve receipts", func(t *testing.T) {
					receipts, err := mgr.GetTxReceiptsIfFinalized(chain, txHashes, confHeight)

					assert.NoError(t, err)
					slices.ForEach(receipts, func(result results.Result[geth.Receipt]) { assert.Equal(t, result.Err(), evm.ErrTxFailed) })
				}),

			When("transactions are finalized", func() {
				latestFinalizedBlockNumber = rand.I64Between(1000, 10000)

				evmClient.TransactionReceiptsFunc = func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
					return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
						return evmRpc.TxReceiptResult(results.FromOk(geth.Receipt{
							BlockNumber: big.NewInt(latestFinalizedBlockNumber - rand.I64Between(1, 100)),
							TxHash:      hash,
							Status:      geth.ReceiptStatusSuccessful,
						}))
					}), nil
				}
			}).
				Then("should return receipt results", func(t *testing.T) {
					receipts, err := mgr.GetTxReceiptsIfFinalized(chain, txHashes, confHeight)

					assert.NoError(t, err)
					assert.True(t, slices.All(receipts, func(result results.Result[geth.Receipt]) bool { return result.Err() == nil }))
				}),

			When("some transactions are not finalized", func() {
				evmClient.TransactionReceiptsFunc = func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
					i := 0
					return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
						var blockNumber *big.Int
						// half of the transactions are finalized
						if i < len(txHashes)/2 {
							blockNumber = big.NewInt(latestFinalizedBlockNumber - rand.I64Between(1, 100))
						} else {
							blockNumber = big.NewInt(latestFinalizedBlockNumber + rand.I64Between(1, 100))
						}
						i++

						return evmRpc.TxReceiptResult(results.FromOk(geth.Receipt{
							BlockNumber: blockNumber,
							TxHash:      hash,
							Status:      geth.ReceiptStatusSuccessful,
						}))
					}), nil
				}
			}).
				Then("should return error results for not found", func(t *testing.T) {
					receipts, err := mgr.GetTxReceiptsIfFinalized(chain, txHashes, confHeight)

					assert.NoError(t, err)
					finalized := receipts[:len(txHashes)/2]
					notFinalized := receipts[len(txHashes)/2:]

					assert.True(t, slices.All(finalized, func(result results.Result[geth.Receipt]) bool { return result.Err() == nil }))
					assert.True(t, slices.All(notFinalized, func(result results.Result[geth.Receipt]) bool { return result.Err() == evm.ErrNotFinalized }))
				}),
		).
		Run(t, 5)
}

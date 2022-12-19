package rpc_test

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

var (
	bytesType   = funcs.Must(abi.NewType("bytes", "bytes", nil))
	bytes32Type = funcs.Must(abi.NewType("bytes32", "bytes32", nil))
	uint256Type = funcs.Must(abi.NewType("uint256", "uint256", nil))

	rollupEventSig       = crypto.Keccak256Hash([]byte("StateBatchAppended(uint256,bytes32,uint256,uint256,bytes)"))
	rollupEventArguments = abi.Arguments{{Type: bytes32Type}, {Type: uint256Type}, {Type: uint256Type}, {Type: bytesType}}
)

func TestIsFinalized(t *testing.T) {
	var (
		l1EthClient *mock.EthereumJSONRPCClientMock

		rollupEvents                 []types.Log
		totalRollups                 sdk.Uint
		contractStateCommitmentChain common.Address
		l2TxBlockNumber              *big.Int
		client                       rpc.Client
	)

	givenOptimismClient := Given("optimism client", func() {
		l1EthClient = &mock.EthereumJSONRPCClientMock{}
		l2EthClient := &mock.EthereumJSONRPCClientMock{}
		l1RpcClient := &mock.JSONRPCClientMock{}
		l2RpcClient := &mock.JSONRPCClientMock{}

		l1EthClient.BlockNumberFunc = func(context.Context) (uint64, error) { return 0, nil }
		l1RpcClient.CallContextFunc = func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			if method == "eth_getBlockByNumber" {
				*result.(**types.Header) = &types.Header{Number: big.NewInt(0)}
			}

			return nil
		}
		l1Client := funcs.Must(rpc.NewEthereum2Client(funcs.Must(rpc.NewEthereumClient(l1EthClient, l1RpcClient))))

		l2EthClient.BlockNumberFunc = func(ctx context.Context) (uint64, error) { return 0, nil }
		l2RpcClient.CallContextFunc = func(ctx context.Context, result interface{}, method string, args ...interface{}) error {
			return nil
		}
		contractStateCommitmentChain = common.BytesToAddress(rand.Bytes(common.HashLength))
		client = funcs.Must(rpc.NewOptimismClient(funcs.Must(rpc.NewEthereumClient(l2EthClient, l2RpcClient)), l1Client, contractStateCommitmentChain))
	})

	whenRollupsArePosted := When("some rollups are posted on L1", func() {
		rollupEventCount := rand.I64Between(1, 1000)
		rollupEvents, totalRollups = randomRollupEvents(contractStateCommitmentChain, rollupEventCount)

		l1EthClient.FilterLogsFunc = func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
			if len(q.Addresses) != 1 || q.Addresses[0] != contractStateCommitmentChain {
				return []types.Log{}, nil
			}

			if len(q.Topics) != 2 && len(q.Topics[0]) != 1 && len(q.Topics[1]) != 1 {
				return []types.Log{}, nil
			}

			if q.Topics[0][0] != rollupEventSig {
				return []types.Log{}, nil
			}

			index := int(new(big.Int).SetBytes(q.Topics[1][0].Bytes()).Int64())
			if index >= len(rollupEvents) {
				return []types.Log{}, nil
			}

			return []types.Log{rollupEvents[index]}, nil
		}
		l1EthClient.CallContractFunc = func(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
			return common.BytesToHash(big.NewInt(rollupEventCount).Bytes()).Bytes(), nil
		}
	})

	whenL2TxIsIncluded := When("the L2 transaction is included in one of those rollups posted on L1", func() {
		l2TxBlockNumber = big.NewInt(rand.I64Between(1, totalRollups.AddUint64(1).BigInt().Int64()))
	})
	whenL2TxIsNotIncluded := When("the L2 transaction is not included in one of those rollups posted on L1", func() {
		switch rand.Bools(0.5).Next() {
		case true:
			l2TxBlockNumber = totalRollups.AddUint64(uint64(rand.I64Between(1, 10))).BigInt()
		default:
			l1EthClient.FilterLogsFunc = func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
				return nil, nil
			}
		}
	})

	whenL1RollupTxIsFinalized := When("the l1 rollup tx is finalized", func() {
		l1EthClient.TransactionReceiptFunc = func(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
			for _, event := range rollupEvents {
				if event.TxHash != txHash {
					continue
				}

				args := funcs.Must(evmtypes.StrictDecode(rollupEventArguments, event.Data))
				batchSize := sdk.NewUintFromBigInt(args[1].(*big.Int))
				prevTotalElements := sdk.NewUintFromBigInt(args[2].(*big.Int))
				l2TxIndex := sdk.NewUintFromBigInt(l2TxBlockNumber).SubUint64(1)

				if l2TxIndex.LT(prevTotalElements) || l2TxIndex.GTE(prevTotalElements.Add(batchSize)) {
					return nil, ethereum.NotFound
				}

				return &types.Receipt{BlockNumber: big.NewInt(0)}, nil
			}

			return nil, ethereum.NotFound
		}
	})
	whenL1RollupTxIsNotFinalized := When("the l1 rollup tx is not finalized", func() {
		l1EthClient.TransactionReceiptFunc = func(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
			return &types.Receipt{BlockNumber: big.NewInt(1)}, nil
		}
	})

	givenOptimismClient.
		When2(whenRollupsArePosted).
		When2(whenL2TxIsIncluded).
		When2(whenL1RollupTxIsFinalized).
		Then("should return true", func(t *testing.T) {
			actual, err := client.IsFinalized(context.Background(), 0, &types.Receipt{BlockNumber: l2TxBlockNumber})

			assert.NoError(t, err)
			assert.True(t, actual)
		}).
		Run(t, 20)

	givenOptimismClient.
		When2(whenRollupsArePosted).
		When2(whenL2TxIsNotIncluded).
		When2(whenL1RollupTxIsFinalized).
		Then("should return false", func(t *testing.T) {
			actual, err := client.IsFinalized(context.Background(), 0, &types.Receipt{BlockNumber: l2TxBlockNumber})

			assert.NoError(t, err)
			assert.False(t, actual)
		}).
		Run(t, 5)

	givenOptimismClient.
		When2(whenRollupsArePosted).
		When2(whenL2TxIsIncluded).
		When2(whenL1RollupTxIsNotFinalized).
		Then("should return false", func(t *testing.T) {
			actual, err := client.IsFinalized(context.Background(), 0, &types.Receipt{BlockNumber: l2TxBlockNumber})

			assert.NoError(t, err)
			assert.False(t, actual)
		}).
		Run(t, 5)
}

func randomRollupEvents(contractStateCommitmentChain common.Address, count int64) ([]types.Log, sdk.Uint) {
	results := make([]types.Log, count)
	prevTotalElements := sdk.NewUint(0)

	for i := int64(0); i < count; i++ {
		batchSize := sdk.NewUint(uint64(rand.I64Between(10, 1000)))
		results[i] = types.Log{
			Address: contractStateCommitmentChain,
			Topics: []common.Hash{
				common.HexToHash("0x16be4c5129a4e03cf3350262e181dc02ddfb4a6008d925368c0899fcd97ca9c5"),
				common.HexToHash(strconv.FormatInt(i, 16)),
			},
			Data: funcs.Must(rollupEventArguments.Pack(
				common.HexToHash(""),
				batchSize.BigInt(),
				prevTotalElements.BigInt(),
				[]byte{},
			)),
			TxHash: common.BytesToHash(rand.Bytes(common.HashLength)),
		}

		prevTotalElements = prevTotalElements.Add(batchSize)
	}

	return results, prevTotalElements
}

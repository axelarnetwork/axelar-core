package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	bytesType   = funcs.Must(abi.NewType("bytes", "bytes", nil))
	bytes32Type = funcs.Must(abi.NewType("bytes32", "bytes32", nil))
	uint256Type = funcs.Must(abi.NewType("uint256", "uint256", nil))

	stateBatchAppendedEventSig       = crypto.Keccak256Hash([]byte("StateBatchAppended(uint256,bytes32,uint256,uint256,bytes)"))
	stateBatchAppendedEventArguments = abi.Arguments{{Type: bytes32Type}, {Type: uint256Type}, {Type: uint256Type}, {Type: bytesType}}
	stateCommitmentChainABI          = funcs.Must(abi.JSON(strings.NewReader(
		`[
			{
				"inputs": [],
				"name": "getTotalBatches",
				"outputs": [
					{
						"internalType": "uint256",
						"name": "_totalBatches",
						"type": "uint256"
					}
				],
				"stateMutability": "view",
				"type": "function"
    	}
		]`,
	)))
	getTotalBatchesMethod = funcs.Must(stateCommitmentChainABI.MethodById(common.Hex2Bytes("e561dddc")))
)

type rollup struct {
	txHash     common.Hash
	minTxIndex sdk.Uint
	maxTxIndex sdk.Uint
}

func (r rollup) isAfter(txIndex sdk.Uint) bool {
	return r.minTxIndex.GT(txIndex)
}

func (r rollup) isBefore(txIndex sdk.Uint) bool {
	return r.maxTxIndex.LT(txIndex)
}

type optimismClient struct {
	*ethereumClient
	l1Client                     *ethereum2Client
	contractStateCommitmentChain common.Address
}

func newOptimismClient(ethereumClient *ethereumClient, l1Client *ethereum2Client, contractStateCommitmentChain common.Address) (*optimismClient, error) {
	client := &optimismClient{ethereumClient: ethereumClient, l1Client: l1Client, contractStateCommitmentChain: contractStateCommitmentChain}
	if _, err := client.getRollupGasPrices(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *optimismClient) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	// Every block has exactly one transaction in it. Since there's a genesis block without transaction, the
	// transaction index will always be one less than the block number.
	// https://github.com/ethereum-optimism/optimism/blob/400d81fb9932677becc8744663938570c198555a/packages/sdk/src/cross-chain-messenger.ts#L1103
	txHash, err := c.getRollupTxHashByTxIndex(ctx, sdk.NewUintFromBigInt(txReceipt.BlockNumber).SubUint64(1))
	if err == ethereum.NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	l1TxReceipt, err := c.l1Client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return false, err
	}

	return c.l1Client.IsFinalized(ctx, 0, l1TxReceipt)
}

// equivalent implementation of https://github.com/ethereum-optimism/optimism/blob/400d81fb9932677becc8744663938570c198555a/packages/sdk/src/cross-chain-messenger.ts#L1169
func (c *optimismClient) getRollupTxHashByTxIndex(ctx context.Context, txIndex sdk.Uint) (common.Hash, error) {
	totalBatches, err := c.getRollupCount(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	lowerBound := sdk.ZeroUint()
	upperBound := sdk.NewUintFromBigInt(totalBatches).SubUint64(1)

	latestRollup, err := c.getRollup(ctx, upperBound)
	if err != nil {
		return common.Hash{}, err
	}

	if latestRollup.isBefore(txIndex) {
		return common.Hash{}, ethereum.NotFound
	}

	if !latestRollup.isAfter(txIndex) {
		return latestRollup.txHash, nil
	}

	upperBound = upperBound.SubUint64(1)
	for batchIndex := lowerBound.Add(upperBound).QuoUint64(2); lowerBound.LTE(upperBound); batchIndex = lowerBound.Add(upperBound).QuoUint64(2) {
		rollup, err := c.getRollup(ctx, batchIndex)
		if err != nil {
			return common.Hash{}, err
		}

		if rollup.isBefore(txIndex) {
			lowerBound = batchIndex.AddUint64(1)
			continue
		}

		if rollup.isAfter(txIndex) {
			upperBound = batchIndex.SubUint64(1)
			continue
		}

		return rollup.txHash, nil
	}

	return common.Hash{}, ethereum.NotFound
}

func (c *optimismClient) getRollup(ctx context.Context, batchIndex sdk.Uint) (*rollup, error) {
	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{c.contractStateCommitmentChain},
		Topics: [][]common.Hash{
			{stateBatchAppendedEventSig},
			{common.BytesToHash(batchIndex.BigInt().Bytes())},
		},
	}

	logs, err := c.l1Client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, ethereum.NotFound
	}

	args := funcs.Must(evmtypes.StrictDecode(stateBatchAppendedEventArguments, logs[0].Data))

	batchSize := sdk.NewUintFromBigInt(args[1].(*big.Int))
	prevTotalElements := sdk.NewUintFromBigInt(args[2].(*big.Int))

	return &rollup{
		txHash:     logs[0].TxHash,
		minTxIndex: prevTotalElements,
		maxTxIndex: prevTotalElements.Add(batchSize).SubUint64(1),
	}, nil
}

func (c *optimismClient) getRollupCount(ctx context.Context) (*big.Int, error) {
	callMsg := ethereum.CallMsg{
		To:   &c.contractStateCommitmentChain,
		Data: getTotalBatchesMethod.ID,
	}
	bz, err := c.l1Client.CallContract(ctx, callMsg, nil)
	if len(bz) != 32 {
		return nil, fmt.Errorf("expected 32 bytes to be received, actual %d", len(bz))
	}
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(bz), nil
}

func (c *optimismClient) getRollupGasPrices(ctx context.Context) (*optimismRollupGasPrices, error) {
	var gasPrices optimismRollupGasPrices
	if err := c.rpc.CallContext(ctx, &gasPrices, "rollup_gasPrices"); err != nil {
		return nil, err
	}

	return &gasPrices, nil
}

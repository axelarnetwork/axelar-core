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
	// Every block has exactly one transaction in it. Since there's a genesis block, the
	// transaction index will always be one less than the block number.
	// https://github.com/ethereum-optimism/optimism/blob/400d81fb9932677becc8744663938570c198555a/packages/sdk/src/cross-chain-messenger.ts#L1103
	event, err := c.getStateBatchAppendedEventByTxIndex(ctx, sdk.NewUintFromBigInt(txReceipt.BlockNumber).SubUint64(1))
	if err == ethereum.NotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	l1TxReceipt, err := c.l1Client.TransactionReceipt(ctx, event.TxHash)
	if err != nil {
		return false, err
	}

	return c.l1Client.IsFinalized(ctx, 0, l1TxReceipt)
}

// equivalent implementation of https://github.com/ethereum-optimism/optimism/blob/400d81fb9932677becc8744663938570c198555a/packages/sdk/src/cross-chain-messenger.ts#L1169
func (c *optimismClient) getStateBatchAppendedEventByTxIndex(ctx context.Context, txIndex sdk.Uint) (types.Log, error) {
	totalBatches, err := c.getTotalBatches(ctx)
	if err != nil {
		return types.Log{}, err
	}

	lowerBound := sdk.ZeroUint()
	upperBound := sdk.NewUintFromBigInt(totalBatches).SubUint64(1)

	event, err := c.getStateBatchAppendedEventByBatchIndex(ctx, upperBound)
	if err != nil {
		return types.Log{}, err
	}

	if isEventLow(event, txIndex) {
		return types.Log{}, ethereum.NotFound
	}

	if !isEventHigh(event, txIndex) {
		return event, nil
	}

	upperBound = upperBound.SubUint64(1)
	for batchIndex := lowerBound.Add(upperBound).QuoUint64(2); lowerBound.LTE(upperBound); batchIndex = lowerBound.Add(upperBound).QuoUint64(2) {
		event, err := c.getStateBatchAppendedEventByBatchIndex(ctx, batchIndex)
		if err != nil {
			return types.Log{}, err
		}

		if isEventLow(event, txIndex) {
			lowerBound = batchIndex.AddUint64(1)
			continue
		}

		if isEventHigh(event, txIndex) {
			upperBound = batchIndex.SubUint64(1)
			continue
		}

		return event, nil
	}

	return types.Log{}, ethereum.NotFound
}

func isEventLow(event types.Log, txIndex sdk.Uint) bool {
	batchSize, prevTotalElements := decodeStateBatchAppendedEvent(event)

	return txIndex.GTE(prevTotalElements.Add(batchSize))
}

func isEventHigh(event types.Log, txIndex sdk.Uint) bool {
	_, prevTotalElements := decodeStateBatchAppendedEvent(event)

	return txIndex.LT(prevTotalElements)
}

func decodeStateBatchAppendedEvent(event types.Log) (sdk.Uint, sdk.Uint) {
	args := funcs.Must(evmtypes.StrictDecode(stateBatchAppendedEventArguments, event.Data))

	batchSize := sdk.NewUintFromBigInt(args[1].(*big.Int))
	prevTotalElements := sdk.NewUintFromBigInt(args[2].(*big.Int))

	return batchSize, prevTotalElements
}

func (c *optimismClient) getStateBatchAppendedEventByBatchIndex(ctx context.Context, batchIndex sdk.Uint) (types.Log, error) {
	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{c.contractStateCommitmentChain},
		Topics: [][]common.Hash{
			{stateBatchAppendedEventSig},
			{common.BytesToHash(batchIndex.BigInt().Bytes())},
		},
	}

	logs, err := c.l1Client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return types.Log{}, err
	}

	if len(logs) == 0 {
		return types.Log{}, ethereum.NotFound
	}

	return logs[0], nil
}

func (c *optimismClient) getTotalBatches(ctx context.Context) (*big.Int, error) {
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

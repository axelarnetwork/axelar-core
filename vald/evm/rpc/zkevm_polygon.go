package rpc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	zkEVMChainIDMainnet = uint64(1101)
	zkEVMChainIDTestnet = uint64(1442)

	zkEVMContractMainnet        = common.HexToAddress("0x5132a183e9f3cb7c848b0aac5ae0c4f0491b7ab2")
	zkEVMContractTestnet        = common.HexToAddress("0xa997cfd539e703921fd1e3cf25b4c241a27a4c7a")
	zkEVMExitRootManagerMainnet = common.HexToAddress("0x580bda1e7A0CFAe92Fa7F6c20A3794F169CE3CFb")
	zkEVMExitRootManagerTestnet = common.HexToAddress("0x4d9427DCA0406358445bC0a8F88C26b704004f74")

	zkEVMVerifyBatchesTrustedSig = crypto.Keccak256Hash([]byte("VerifyBatchesTrustedAggregator(uint64,bytes32,address)")).Bytes()
	zkEVMGlobalRootFuncSig       = crypto.Keccak256Hash([]byte("globalExitRootMap(bytes32)")).Bytes()[:4]
)

// ZkEvmPolygonClient is a JSON-RPC client of Moonbeam
type ZkEvmPolygonClient struct {
	*EthereumClient
	l1Client           *Ethereum2Client
	l1Contract         common.Address
	l1ExitRootsManager common.Address
}

// NewZkEvmPolygonClient is the constructor
func NewZkEvmPolygonClient(ethereumClient *EthereumClient, l1Client *Ethereum2Client) (*ZkEvmPolygonClient, error) {
	var chainID hexutil.Big
	err := ethereumClient.rpc.CallContext(context.Background(), &chainID, "eth_chainId")
	if err != nil {
		return nil, err
	}

	var client *ZkEvmPolygonClient
	switch chainID.ToInt().Uint64() {
	case zkEVMChainIDMainnet:
		client = &ZkEvmPolygonClient{
			EthereumClient:     ethereumClient,
			l1Client:           l1Client,
			l1Contract:         zkEVMContractMainnet,
			l1ExitRootsManager: zkEVMExitRootManagerMainnet,
		}
	case zkEVMChainIDTestnet:
		client = &ZkEvmPolygonClient{
			EthereumClient:     ethereumClient,
			l1Client:           l1Client,
			l1Contract:         zkEVMContractTestnet,
			l1ExitRootsManager: zkEVMExitRootManagerTestnet,
		}
	default:
		return nil, fmt.Errorf("invalid chain ID for chain zkEVM Polygon")
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	batchNumber, err := client.getBatchNumberByBlockNumber(context.Background(), header.Number)
	if err != nil {
		return nil, err
	}

	if _, err := client.getBatchByNumber(context.Background(), batchNumber); err != nil {
		return nil, err
	}

	if _, err := client.getVerifiedBatchNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
func (c *ZkEvmPolygonClient) IsFinalized(ctx context.Context, _ uint64, l2Receipt *types.Receipt) (bool, error) {

	// retrieve the batch from the block number in the l2 receipt
	batchNumber, err := c.getBatchNumberByBlockNumber(ctx, (*hexutil.Big)(l2Receipt.BlockNumber))
	if err != nil {
		return false, err
	}
	batch, err := c.getBatchByNumber(ctx, batchNumber)
	if err != nil {
		return false, err
	}

	// if no VerifyBatch tx was yet sent to the l1, it means no proof was generated yet
	if batch.VerifyBatchTxHash == nil {
		return false, nil
	}

	// ensure the batch's GlobalExitRoot value is known by the l1's manager
	timestamp, err := c.getL1GlobalExitRootTimestamp(ctx, batch.GlobalExitRoot)
	if err != nil {
		return false, err
	}
	if timestamp.Uint64() == 0 {
		return false, fmt.Errorf("global exit root %s not found at manager", batch.GlobalExitRoot.Hex())
	}

	// ensure the VerifyBatch tx emitted the expected log (zkEVMVerifyBatchesTrustedAggregator)
	l1Receipt, err := c.l1Client.TransactionReceipt(ctx, *batch.VerifyBatchTxHash)
	if err != nil {
		return false, err
	}
	var log *types.Log
	for _, l := range l1Receipt.Logs {
		if len(l.Topics) > 0 && bytes.Equal(l.Topics[0].Bytes(), zkEVMVerifyBatchesTrustedSig) {
			log = l
		}
	}
	if log == nil {
		return false, fmt.Errorf("unable to find VerifyBatchesTrusted log in transaction receipt")
	}
	if len(log.Topics) != 3 {
		return false, fmt.Errorf("unexpected amount of topics at VerifyBatchesTrusted log (want 3, got %d)", len(log.Topics))
	}
	l1ContractAddress := log.Address
	if bytes.Equal(l1ContractAddress.Bytes(), c.l1Contract.Bytes()) {
		return false, fmt.Errorf("wrong contract address at log index %d (want %s, got %s)", log.Index, c.l1Contract.Hex(), l1ContractAddress.Hex())
	}

	// ensure the batch is deemed verified also on the l2
	lastVerifiedBatchNum, err := c.getVerifiedBatchNumber(ctx)
	l1LastBatchNumInSequence := log.Topics[1].Big()
	if err != nil {
		return false, err
	}
	if lastVerifiedBatchNum.ToInt().Cmp(l1LastBatchNumInSequence) < 0 {
		return false, fmt.Errorf("verified batch number on l2 must be greater than or equal to the last batch in the sequence on l1: expected at least %s, got %s", lastVerifiedBatchNum.ToInt().String(), l1LastBatchNumInSequence.String())
	}

	// ensure batch's stateRoot value matches with the event
	verifiedBatch, err := c.getBatchByNumber(ctx, (*hexutil.Big)(l1LastBatchNumInSequence))
	if err != nil {
		return false, err
	}
	l1StateRoot := common.BytesToHash(log.Data)
	if bytes.Equal(verifiedBatch.StateRoot.Bytes(), l1StateRoot.Bytes()) {
		return false, fmt.Errorf("verified stateRoot mismatch: expected %s, got %s", batch.StateRoot.Hex(), l1StateRoot.Hex())
	}

	// ensure verifyBatch tx is finalized on the l1 chain
	return c.l1Client.IsFinalized(ctx, 0, l1Receipt)
}

func (c *ZkEvmPolygonClient) getBatchNumberByBlockNumber(ctx context.Context, number *hexutil.Big) (*hexutil.Big, error) {
	var blockNumber hexutil.Big
	if err := c.rpc.CallContext(ctx, &blockNumber, "zkevm_batchNumberByBlockNumber", number.String()); err != nil {
		return nil, err
	}

	return &blockNumber, nil
}

func (c *ZkEvmPolygonClient) getBatchByNumber(ctx context.Context, number *hexutil.Big) (zkEvmPolygonHeader, error) {
	var zkHeader zkEvmPolygonHeader
	if err := c.rpc.CallContext(ctx, &zkHeader, "zkevm_getBatchByNumber", number.String(), false); err != nil {
		return zkEvmPolygonHeader{}, err
	}

	return zkHeader, nil
}

func (c *ZkEvmPolygonClient) getVerifiedBatchNumber(ctx context.Context) (*hexutil.Big, error) {
	var blockNumber hexutil.Big
	if err := c.rpc.CallContext(ctx, &blockNumber, "zkevm_verifiedBatchNumber"); err != nil {
		return nil, err
	}

	return &blockNumber, nil
}

func (c *ZkEvmPolygonClient) getL1GlobalExitRootTimestamp(ctx context.Context, globalExitRoot common.Hash) (*big.Int, error) {
	globalExitRootBytes := globalExitRoot.Bytes()
	data := append(zkEVMGlobalRootFuncSig, globalExitRootBytes...)

	callMsg := ethereum.CallMsg{
		To:   &c.l1ExitRootsManager,
		Data: data,
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

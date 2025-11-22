package evm_test

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mock2 "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmRpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/monads/results"
)

// LogData represents the structure of log data in the JSON file
type LogData struct {
	LogIndex uint     `json:"logIndex"`
	Address  string   `json:"address"`
	Topics   []string `json:"topics"`
	Data     string   `json:"data"`
}

// loadLogsFromTestdata loads the transaction logs from the testdata JSON file
func loadLogsFromTestdata(t *testing.T, txHash common.Hash, blockNumber uint64) []*geth.Log {
	t.Helper()

	data, err := os.ReadFile("testdata/polygon-eip-7702-tx-logs.json")
	require.NoError(t, err, "Failed to read logs JSON file")

	var logsData []LogData
	err = json.Unmarshal(data, &logsData)
	require.NoError(t, err, "Failed to unmarshal logs JSON")

	logs := make([]*geth.Log, len(logsData))
	for i, logData := range logsData {
		topics := make([]common.Hash, len(logData.Topics))
		for j, topic := range logData.Topics {
			topics[j] = common.HexToHash(topic)
		}

		logs[i] = &geth.Log{
			Address:     common.HexToAddress(logData.Address),
			Topics:      topics,
			Data:        common.FromHex(logData.Data),
			BlockNumber: blockNumber,
			TxHash:      txHash,
			TxIndex:     0,
			BlockHash:   common.Hash{},
			Index:       uint(logData.LogIndex),
			Removed:     false,
		}
	}

	return logs
}

func TestMgr_ProcessGatewayTxsConfirmationMissingBlockNumberNotPanics(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	receipt := geth.Receipt{Logs: []*geth.Log{{Topics: make([]common.Hash, 0)}}}
	rpcClient := &mock.ClientMock{TransactionReceiptsFunc: func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
		return []evmRpc.TxReceiptResult{evmRpc.TxReceiptResult(results.FromOk(receipt))}, nil
	}}
	cache := &evmmock.LatestFinalizedBlockCacheMock{GetFunc: func(chain nexus.ChainName) *big.Int {
		return big.NewInt(100)
	}}

	broadcaster := &mock2.BroadcasterMock{BroadcastFunc: func(_ context.Context, _ ...sdk.Msg) (*sdk.TxResponse, error) {
		return nil, nil
	}}

	valAddr := rand.ValAddr()
	mgr := evm.NewMgr(map[string]evmRpc.Client{chain.String(): rpcClient}, broadcaster, valAddr, rand.AccAddr(), cache)

	assert.NotPanics(t, func() {
		require.NoError(t, mgr.ProcessGatewayTxsConfirmation(&types.ConfirmGatewayTxsStarted{
			PollMappings: []types.PollMapping{{PollID: 10, TxID: types.Hash{1}}},
			Participants: []sdk.ValAddress{valAddr},
			Chain:        chain,
		}))
	})
}

func TestMgr_ProcessGatewayTxsConfirmationNoTopicsNotPanics(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	receipt := geth.Receipt{
		Logs:        []*geth.Log{{Topics: make([]common.Hash, 0)}},
		BlockNumber: big.NewInt(1),
		Status:      geth.ReceiptStatusSuccessful,
	}
	rpcClient := &mock.ClientMock{TransactionReceiptsFunc: func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
		return []evmRpc.TxReceiptResult{evmRpc.TxReceiptResult(results.FromOk(receipt))}, nil
	}}
	cache := &evmmock.LatestFinalizedBlockCacheMock{GetFunc: func(chain nexus.ChainName) *big.Int {
		return big.NewInt(100)
	}}

	broadcaster := &mock2.BroadcasterMock{BroadcastFunc: func(_ context.Context, _ ...sdk.Msg) (*sdk.TxResponse, error) {
		return nil, nil
	}}

	valAddr := rand.ValAddr()
	mgr := evm.NewMgr(map[string]evmRpc.Client{chain.String(): rpcClient}, broadcaster, valAddr, rand.AccAddr(), cache)

	assert.NotPanics(t, func() {
		require.NoError(t, mgr.ProcessGatewayTxsConfirmation(&types.ConfirmGatewayTxsStarted{
			PollMappings: []types.PollMapping{{PollID: 10, TxID: types.Hash{1}}},
			Participants: []sdk.ValAddress{valAddr},
			Chain:        chain,
		}))
	})
}

// TestEIP7702TransactionConfirmation tests that EIP-7702 transactions can be decoded and confirmed
func TestEIP7702TransactionConfirmation(t *testing.T) {
	// Read the EIP-7702 transaction from testdata
	txData, err := os.ReadFile("testdata/polygon-eip-7702-tx.txt")
	require.NoError(t, err, "Failed to read testdata file")

	// Remove any whitespace and the 0x prefix if present
	txHex := strings.TrimSpace(string(txData))
	txHex = strings.TrimPrefix(txHex, "0x")

	// Decode the transaction
	txBytes, err := hexutil.Decode("0x" + txHex)
	require.NoError(t, err, "Failed to decode transaction hex")

	var tx geth.Transaction
	err = tx.UnmarshalBinary(txBytes)
	require.NoError(t, err, "Failed to unmarshal transaction")

	// Verify it's an EIP-7702 transaction (type 0x04)
	assert.Equal(t, uint8(0x04), tx.Type(), "Expected EIP-7702 transaction type")

	// Verify the transaction hash matches the expected value from PolygonScan
	expectedHash := common.HexToHash("0x15ca6e45cca157db5033cd419a23063881f56241eecfac5e3f4b61b910835b62")
	assert.Equal(t, expectedHash, tx.Hash(), "Transaction hash should match")

	// Actual data from PolygonScan:
	// Block: 78233976
	// Status: Success
	// Gas Used: 554,126
	// The transaction emitted 29 logs total (indices 809-837)
	blockNumber := big.NewInt(78233976)
	gasUsed := uint64(554126)

	// Load all 29 logs from testdata
	logs := loadLogsFromTestdata(t, tx.Hash(), blockNumber.Uint64())
	require.Len(t, logs, 29, "Should have loaded all 29 logs")

	// Create a mock receipt for this transaction with actual data
	receipt := &geth.Receipt{
		Type:              tx.Type(),
		Status:            geth.ReceiptStatusSuccessful,
		TxHash:            tx.Hash(),
		BlockNumber:       blockNumber,
		GasUsed:           gasUsed,
		CumulativeGasUsed: gasUsed,
		Logs:              logs,
	}

	// Setup mock RPC client
	chain := nexus.ChainName("polygon")
	rpcClient := &mock.ClientMock{
		TransactionReceiptsFunc: func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
			require.Len(t, txHashes, 1)
			assert.Equal(t, tx.Hash(), txHashes[0])
			return []evmRpc.TxReceiptResult{evmRpc.TxReceiptResult(results.FromOk(*receipt))}, nil
		},
	}

	// Setup mock cache that returns finalized block
	// Using a block number higher than the transaction block to ensure it's finalized
	cache := &evmmock.LatestFinalizedBlockCacheMock{
		GetFunc: func(c nexus.ChainName) *big.Int {
			return big.NewInt(78375520) // Block number + ~141,544 confirmations
		},
	}

	// Setup mock broadcaster to capture the confirmation message
	var broadcastedMsg sdk.Msg
	broadcaster := &mock2.BroadcasterMock{
		BroadcastFunc: func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
			require.Len(t, msgs, 1)
			broadcastedMsg = msgs[0]
			return &sdk.TxResponse{}, nil
		},
	}

	// Create validator address and setup manager
	valAddr := rand.ValAddr()
	mgr := evm.NewMgr(
		map[string]evmRpc.Client{chain.String(): rpcClient},
		broadcaster,
		valAddr,
		rand.AccAddr(),
		cache,
	)

	// Convert transaction hash to types.Hash
	var txID types.Hash
	copy(txID[:], tx.Hash().Bytes())

	// Process the transaction through the manager using the non-deprecated API
	// Use the actual Axelar gateway address for Polygon
	gatewayAddress := types.Address(common.HexToAddress("0x6f015F16De9fC8791b234eF68D486d2bF203FBA8"))

	pollID := vote.PollID(rand.PosI64())
	err = mgr.ProcessGatewayTxsConfirmation(&types.ConfirmGatewayTxsStarted{
		PollMappings: []types.PollMapping{
			{
				TxID:   txID,
				PollID: pollID,
			},
		},
		Participants:   []sdk.ValAddress{valAddr},
		Chain:          chain,
		GatewayAddress: gatewayAddress,
	})
	require.NoError(t, err, "Failed to process gateway txs confirmation")

	// Verify that a confirmation message was broadcast
	require.NotNil(t, broadcastedMsg, "Should have broadcast a confirmation message")

	// Verify the message is a VoteRequest
	voteMsg, ok := broadcastedMsg.(*votetypes.VoteRequest)
	require.True(t, ok, "Broadcasted message should be VoteRequest, got %T", broadcastedMsg)

	// Verify the vote contains the confirmation
	assert.NotNil(t, voteMsg.Vote, "Vote should contain confirmation data")
	t.Logf("Vote broadcasted for poll ID: %v", voteMsg.PollID)

	// Decode and log what's in the vote
	voteContent := voteMsg.Vote.GetCachedValue()
	require.NotNil(t, voteContent, "Vote should have cached value")

	voteEvents, ok := voteContent.(*types.VoteEvents)
	require.True(t, ok, "Vote content should be VoteEvents, got %T", voteContent)

	t.Logf("Vote chain: %s", voteEvents.Chain)
	t.Logf("Number of events in vote: %d", len(voteEvents.Events))

	// Verify we found the expected gateway event
	require.Len(t, voteEvents.Events, 1, "Should have found 1 gateway event in the EIP-7702 transaction")

	for i, event := range voteEvents.Events {
		t.Logf("Event %d: TxID=%s, Index=%d, Type=%T", i, event.TxID.Hex(), event.Index, event.Event)

		// This is a ContractCallWithToken event - verify it
		if ccwt, ok := event.Event.(*types.Event_ContractCallWithToken); ok {
			t.Logf("  ContractCallWithToken details:")
			t.Logf("    Sender: %s", ccwt.ContractCallWithToken.Sender.Hex())
			t.Logf("    DestinationChain: %s", ccwt.ContractCallWithToken.DestinationChain)
			t.Logf("    ContractAddress: %s", ccwt.ContractCallWithToken.ContractAddress)
			t.Logf("    Symbol: %s", ccwt.ContractCallWithToken.Symbol)
			t.Logf("    Amount: %s", ccwt.ContractCallWithToken.Amount.String())
		}
	}

	// Also verify we can retrieve the receipt directly
	result, err := mgr.GetTxReceiptIfFinalized(chain, tx.Hash(), 10)
	require.NoError(t, err)
	require.Nil(t, result.Err())
	actualReceipt := result.Ok()

	// Verify receipt properties
	assert.Equal(t, tx.Hash(), actualReceipt.TxHash)
	assert.Equal(t, geth.ReceiptStatusSuccessful, actualReceipt.Status)
	assert.Equal(t, uint8(0x04), actualReceipt.Type)
	assert.Len(t, actualReceipt.Logs, 29, "Receipt should contain all 29 logs")

	// Verify first log (ERC20 Approval)
	assert.Equal(t, common.HexToAddress("0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"), actualReceipt.Logs[0].Address)
	assert.Equal(t, common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"), actualReceipt.Logs[0].Topics[0])

	// Verify last log index
	assert.Equal(t, uint(837), actualReceipt.Logs[28].Index)
}

// TestEIP7702TransactionProperties verifies the properties of the EIP-7702 transaction
func TestEIP7702TransactionProperties(t *testing.T) {
	// Read the EIP-7702 transaction from testdata
	txData, err := os.ReadFile("testdata/polygon-eip-7702-tx.txt")
	require.NoError(t, err, "Failed to read testdata file")

	// Remove any whitespace and the 0x prefix if present
	txHex := strings.TrimSpace(string(txData))
	txHex = strings.TrimPrefix(txHex, "0x")

	// Decode the transaction
	txBytes, err := hexutil.Decode("0x" + txHex)
	require.NoError(t, err, "Failed to decode transaction hex")

	var tx geth.Transaction
	err = tx.UnmarshalBinary(txBytes)
	require.NoError(t, err, "Failed to unmarshal transaction")

	// Verify transaction type
	assert.Equal(t, uint8(0x04), tx.Type(), "Expected EIP-7702 transaction type")

	// Verify the transaction hash matches the expected value from PolygonScan
	expectedHash := common.HexToHash("0x15ca6e45cca157db5033cd419a23063881f56241eecfac5e3f4b61b910835b62")
	assert.Equal(t, expectedHash, tx.Hash(), "Transaction hash should match")

	// Verify actual transaction properties from the raw transaction data
	// From PolygonScan: To address is 0x0E46f4A712a340ffF4C5b0875595723Df3E4b9FB
	expectedTo := common.HexToAddress("0x0E46f4A712a340ffF4C5b0875595723Df3E4b9FB")
	assert.Equal(t, expectedTo, *tx.To(), "To address should match")

	// From PolygonScan: Gas limit is 697,912
	assert.Equal(t, uint64(697912), tx.Gas(), "Gas limit should match")

	// From PolygonScan: Value is 0 POL
	assert.Equal(t, big.NewInt(0), tx.Value(), "Value should be 0")

	// Verify transaction has expected fields
	assert.NotNil(t, tx.To(), "Transaction should have a 'to' address")
	assert.NotNil(t, tx.Value(), "Transaction should have a value")
	assert.NotNil(t, tx.Gas(), "Transaction should have gas")
	assert.NotNil(t, tx.Data(), "Transaction should have data")

	// Log some transaction details for debugging
	t.Logf("Transaction Hash: %s", tx.Hash().Hex())
	t.Logf("Transaction Type: 0x%02x", tx.Type())
	t.Logf("To: %s", tx.To().Hex())
	t.Logf("Gas: %d", tx.Gas())
	t.Logf("Value: %s", tx.Value().String())
	t.Logf("Data length: %d bytes", len(tx.Data()))
}

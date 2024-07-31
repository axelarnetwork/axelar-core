package evm_test

import (
	"context"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	mock2 "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmRpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/monads/results"
)

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
		mgr.ProcessGatewayTxsConfirmation(&types.ConfirmGatewayTxsStarted{
			PollMappings: []types.PollMapping{{PollID: 10, TxID: types.Hash{1}}},
			Participants: []sdk.ValAddress{valAddr},
			Chain:        chain,
		})
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
		mgr.ProcessGatewayTxsConfirmation(&types.ConfirmGatewayTxsStarted{
			PollMappings: []types.PollMapping{{PollID: 10, TxID: types.Hash{1}}},
			Participants: []sdk.ValAddress{valAddr},
			Chain:        chain,
		})
	})
}

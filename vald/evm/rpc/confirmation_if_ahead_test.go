package rpc_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
)

func TestConfirmationIfAheadClient_LatestFinalizedBlockNumber(t *testing.T) {
	const confirmations = 10
	errRPC := errors.New("rpc error")

	testCases := []struct {
		name         string
		latest       uint64
		latestErr    error
		finalized    int64
		finalizedErr error
		expected     *big.Int
	}{
		{name: "finalized tag ahead of confirmations", latest: 100, finalized: 95, expected: big.NewInt(95)},
		{name: "confirmations ahead of finalized tag", latest: 100, finalized: 80, expected: big.NewInt(91)},
		{name: "finalized tag fails", latest: 100, finalizedErr: errRPC, expected: big.NewInt(91)},
		{name: "block number fails", latestErr: errRPC, finalized: 80, expected: big.NewInt(80)},
		{name: "both fail", latestErr: errRPC, finalizedErr: errRPC},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ethClient := &mock.EthereumJSONRPCClientMock{
				BlockNumberFunc: func(context.Context) (uint64, error) { return tc.latest, nil },
			}
			jsonRPC := &mock.JSONRPCClientMock{
				CallContextFunc: func(_ context.Context, result interface{}, method string, args ...interface{}) error {
					assert.Equal(t, "eth_getBlockByNumber", method)
					assert.Equal(t, "finalized", args[0])
					*result.(**types.Header) = &types.Header{Number: big.NewInt(tc.finalized)}
					return nil
				},
			}

			// the constructors probe the endpoints, so only inject failures after construction
			ethereumClient, err := rpc.NewEthereumClient(ethClient, jsonRPC)
			assert.NoError(t, err)
			ethereum2Client, err := rpc.NewEthereum2Client(ethereumClient)
			assert.NoError(t, err)
			client := rpc.NewConfirmationIfAheadClient(ethereum2Client)

			if tc.latestErr != nil {
				ethClient.BlockNumberFunc = func(context.Context) (uint64, error) { return 0, tc.latestErr }
			}
			if tc.finalizedErr != nil {
				jsonRPC.CallContextFunc = func(context.Context, interface{}, string, ...interface{}) error { return tc.finalizedErr }
			}

			actual, err := client.LatestFinalizedBlockNumber(context.Background(), confirmations)
			if tc.expected == nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

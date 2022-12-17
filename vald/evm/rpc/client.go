package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate moq -out ./mock/client.go -pkg mock . Client

// Client provides calls to EVM JSON-RPC endpoints
type Client interface {
	// TransactionReceipt returns the transaction receipt for the given transaction hash
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// HeaderByNumber returns the block header for the given block number
	HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error)
	// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
	IsFinalized(ctx context.Context, conf uint64, txReceipt *types.Receipt) (bool, error)
	// Close closes the client connection
	Close()
}

// NewClient returns an EVM JSON-RPC client
func NewClient(url string) (Client, error) {
	ethereumClient, err := newEthereumClient(url)
	if err != nil {
		return nil, err
	}

	if moonbeamClient, err := newMoonbeamClient(ethereumClient); err == nil {
		return moonbeamClient, nil
	}

	if ethereum2Client, err := newEthereum2Client(ethereumClient); err == nil {
		return ethereum2Client, nil
	}

	return ethereumClient, nil
}

// NewL2Client returns a L2 EVM JSON-RPC client
func NewL2Client(config evmTypes.EVMConfig, l1Client Client) (Client, error) {
	ethereumClient, err := newEthereumClient(config.RPCAddr)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(config.Name) {
	case "arbitrum":
		eth2Client, ok := l1Client.(*ethereum2Client)
		if !ok {
			return nil, fmt.Errorf("l1 client has to be ethereum 2.0 for arbitrum")
		}

		return newArbitrumClient(ethereumClient, eth2Client)
	case "optimism":
		eth2Client, ok := l1Client.(*ethereum2Client)
		if !ok {
			return nil, fmt.Errorf("l1 client has to be ethereum 2.0 for optimism")
		}

		if config.StateCommitmentChain == nil {
			return nil, fmt.Errorf("state commitment chain is required for optimism")
		}

		if !common.IsHexAddress(*config.StateCommitmentChain) {
			return nil, fmt.Errorf("state commitment chain is not a valid evm address")
		}

		return newOptimismClient(ethereumClient, eth2Client, common.HexToAddress(*config.StateCommitmentChain))
	default:
		return nil, fmt.Errorf("unsupported L2 chain %s", config.Name)
	}
}

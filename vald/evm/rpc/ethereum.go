package rpc

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type ethereumClient struct {
	*ethclient.Client
	rpc *rpc.Client
}

func newEthereumClient(url string) (*ethereumClient, error) {
	ethClient, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	rpc, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	client := &ethereumClient{
		Client: ethClient,
		rpc:    rpc,
	}
	// validate that the given url implements standard ethereum JSON-RPC
	if _, err := client.BlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ethereumClient) HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error) {
	var head *Header
	err := c.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}

	return head, err
}

func (c *ethereumClient) IsFinalized(ctx context.Context, conf uint64, txReceipt *types.Receipt) (bool, error) {
	blockNumber, err := c.BlockNumber(ctx)
	if err != nil {
		return false, err
	}

	latestFinalizedBlockNumber := sdk.NewIntFromUint64(blockNumber).SubRaw(int64(conf)).AddRaw(1).BigInt()

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

// copied from https://github.com/ethereum/go-ethereum/blob/69568c554880b3567bace64f8848ff1be27d084d/ethclient/ethclient.go#L565
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}

	return hexutil.EncodeBig(number)
}

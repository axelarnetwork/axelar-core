package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/utils/funcs"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	nodeInterfaceAddress = common.HexToAddress("0x00000000000000000000000000000000000000C8") // https://github.com/OffchainLabs/nitro/blob/master/contracts/src/node-interface/NodeInterface.sol
	nodeInterfaceABI     = funcs.Must(abi.JSON(strings.NewReader(
		`[
			{
				"inputs": [
					{
						"internalType": "bytes32",
						"name": "blockHash",
						"type": "bytes32"
					}
				],
				"name": "getL1Confirmations",
				"outputs": [
					{
						"internalType": "uint64",
						"name": "confirmations",
						"type": "uint64"
					}
				],
				"stateMutability": "view",
				"type": "function"
			}
		]`,
	)))
	getL1ConfirmationsMethod = funcs.Must(nodeInterfaceABI.MethodById(common.Hex2Bytes("e5ca238c")))
)

// arbitrumClient implements ArbitrumClient
type arbitrumClient struct {
	*ethereumClient
	l1Client *ethereum2Client
}

func newArbitrumClient(ethereumClient *ethereumClient, l1Client Client) (*arbitrumClient, error) {
	// TODO: verify that the given l1 client corresponds to the Arbitrum chain, but how?
	eth2Client, ok := l1Client.(*ethereum2Client)
	if !ok {
		return nil, fmt.Errorf("l1 client has to be ethereum 2.0 for arbitrum")
	}
	client := &arbitrumClient{ethereumClient: ethereumClient, l1Client: eth2Client}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	if _, err := client.getL1Confirmations(context.Background(), header.Hash); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *arbitrumClient) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	l1Confirmations, err := c.getL1Confirmations(ctx, txReceipt.BlockHash)
	if err != nil {
		return false, err
	}
	if l1Confirmations.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}

	l1LatestFinalizedBlockNumber, err := c.l1Client.latestFinalizedBlockNumber(ctx)
	if err != nil {
		return false, err
	}

	l1LatestBlockNumber, err := c.l1Client.BlockNumber(ctx)
	if err != nil {
		return false, err
	}

	return sdk.NewIntFromUint64(l1LatestBlockNumber).Sub(sdk.NewIntFromBigInt(l1LatestFinalizedBlockNumber)).AddRaw(1).BigInt().Cmp(l1Confirmations) <= 0, nil
}

func (c *arbitrumClient) getL1Confirmations(ctx context.Context, blockHash common.Hash) (*big.Int, error) {
	data := append(getL1ConfirmationsMethod.ID, funcs.Must(getL1ConfirmationsMethod.Inputs.Pack(blockHash))...)
	callMsg := ethereum.CallMsg{
		To:   &nodeInterfaceAddress,
		Data: data,
	}
	bz, err := c.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(bz), nil
}

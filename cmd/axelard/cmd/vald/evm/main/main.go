package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
)

func isTxSuccessful(txReceipt *geth.Receipt) bool {
	return txReceipt.Status == 1
}

func getLatestFinalizedBlockNumber(client rpc.Client, confHeight uint64) (*big.Int, error) {
	switch client := client.(type) {
	case rpc.MoonbeamClient:
		finalizedBlockHash, err := client.ChainGetFinalizedHead(context.Background())
		if err != nil {
			return nil, err
		}

		header, err := client.ChainGetHeader(context.Background(), finalizedBlockHash)
		if err != nil {
			return nil, err
		}

		return header.Number.ToInt(), nil
	default:
		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, err
		}

		return big.NewInt(int64(blockNumber - confHeight + 1)), nil
	}
}

func validate(rpc rpc.Client, txID common.Hash, confHeight uint64, validateTx func(tx *geth.Transaction, txReceipt *geth.Receipt) bool) bool {
	tx, _, err := rpc.TransactionByHash(context.Background(), txID)
	if err != nil {
		fmt.Print(sdkerrors.Wrap(err, "get transaction by hash call failed").Error() + "\n")
		return false
	}

	txReceipt, err := rpc.TransactionReceipt(context.Background(), txID)
	if err != nil {
		fmt.Print(sdkerrors.Wrap(err, "get transaction receipt call failed").Error() + "\n")
		return false
	}

	if !isTxSuccessful(txReceipt) {
		fmt.Printf("transaction %s failed\n", txReceipt.TxHash.String())
		return false
	}

	latestFinalizedBlockNumber, err := getLatestFinalizedBlockNumber(rpc, confHeight)
	if err != nil {
		fmt.Print(sdkerrors.Wrap(err, "get latest finalized block number failed").Error() + "\n")
		return false
	}

	if latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) < 0 {
		fmt.Printf("transaction %s is not finalized yet\n", txReceipt.TxHash.String())
		return false
	}

	txBlock, err := rpc.BlockByNumber(context.Background(), txReceipt.BlockNumber)
	if err != nil {
		fmt.Print(sdkerrors.Wrap(err, "gete block by number call failed").Error() + "\n")
		return false
	}

	fmt.Printf("txBlock.Number() %#v\n", txBlock.Number())
	fmt.Printf("txReceipt.BlockNumber %#v\n", txReceipt.BlockNumber)

	fmt.Printf("txBlock.Hash() %#v\n", txBlock.Hash().Hex())
	fmt.Printf("txReceipt.BlockHash %#v\n", txReceipt.BlockHash.Hex())

	txFound := false
	for _, tx := range txBlock.Body().Transactions {
		if bytes.Equal(tx.Hash().Bytes(), txReceipt.TxHash.Bytes()) {
			txFound = true
			break
		}
	}

	if !txFound {
		fmt.Printf("transaction %s is not found in block number %d and hash %s\n", txReceipt.TxHash.String(), txBlock.NumberU64(), txBlock.Hash().String())
		return false
	}

	return validateTx(tx, txReceipt)
}

func main() {
	// url := "https://ropsten.infura.io/v3/2be110f3450b494f8d637ed7bb6954e3"
	// txHash := "0xf5f4950206cf18c570115931df753adfe8bff958eac32c28ac222ad70a6192b2"
	url := "https://moonbeam.api.onfinality.io/public"
	txHash := "0x67a9f370afb2b6bc08e2396a302afaea18b8d6c1b060348ca6ba534b32eecbef"
	rpc, err := rpc.NewClient(url)
	if err != nil {
		panic(err)
	}

	isFinalized := validate(rpc, common.HexToHash(txHash), 15, func(tx *geth.Transaction, txReceipt *geth.Receipt) bool { return true })

	fmt.Printf("isFinalized %#v\n", isFinalized)
}

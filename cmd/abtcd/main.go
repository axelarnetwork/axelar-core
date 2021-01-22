package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"

	//"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/wallet"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cdc := app.MakeCodec()

	// Configure wallet
	config := *wallet.DefaultConfig()
	//fmt.Printf("%+v\n", config)

	wallet, err := wallet.CreateWallet(config)
	if err != nil {
		return err
	}

	// Import account
	if err := wallet.ImportMnemonicFromFile("abtcd_mnemonic2.txt", "abtcd"); err != nil {
		return err
	}

	// Instantiate REST context for building and submitting transactions
	restCtx := rest.RestContext{Codec: cdc, URL: "http://localhost:1317"}
	if err := restCtx.TxSnapshotNow(wallet); err != nil {
		return err
	}

	return nil
}

func exampleSignTx(cdc *codec.Codec, w *wallet.Wallet) error {
	stdTx, err := utils.ReadStdTxFromFile(cdc, "unsignedTx.json")
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", stdTx)

	signedTx, err := w.SignStdTx(stdTx, false)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", signedTx)

	return nil
}
package main

import (
	"fmt"
	"os"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest/modules"
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
	if err := modules.TxSnapshotNow(&wallet, &restCtx); err != nil {
		return err
	}

	return nil
}

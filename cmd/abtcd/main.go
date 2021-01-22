package main

import (
	"fmt"
	"os"
	"strings"

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

	// Import account to wallet from mnemonic file
	config := *wallet.DefaultConfig()
	w, err := wallet.CreateWallet(config)
	if err != nil {
		return err
	}
	if err := w.ImportMnemonicFromFile("abtcd_mnemonic.txt", "abtcd"); err != nil {
		return err
	}

	// Instantiate REST context for building and submitting transactions
	restCtx := rest.RestContext{Codec: cdc, URL: "http://localhost:1317"}
	app := modules.AppContext{ Wallet: w, RestCtx: restCtx }

	return mint(app)
}

func mint(app modules.AppContext) (err error) {
	// 0. Assume master keys are set up
	if err = app.TxSnapshotNow(); err != nil {
		return
	}


	// A. Link eth and btc addresses
	// 1. link eth_address
	fmt.Print("Enter ethereum address to link: ")
	//reader := bufio.NewReader(os.Stdin)
	//ethAddr, err := reader.ReadString('\n')
	ethAddr := "0x68B93045fe7D8794a7cAF327e7f855CD6Cd03BB8"
	fmt.Println(ethAddr)
	ethAddr = strings.TrimSpace(ethAddr)

	err = app.TxBitcoinLink("Ethereum", ethAddr)

	depositAddr, err := app.QueryDepositAddress("Ethereum", ethAddr)
	fmt.Printf("Deposit BTC to %s\n", depositAddr)
	fmt.Print("Hit return once btc is deposited...")
	fmt.Scanln()

	return
}

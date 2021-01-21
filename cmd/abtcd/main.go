package main

import (
	"fmt"
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

	config := *wallet.DefaultConfig()
	fmt.Printf("%+v\n", config)
	w, err := wallet.CreateWalletFromMnemoic(config, "abtcd_mnemonic.txt")
	if err != nil {
		return err
	}

	// stdTx
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

	restCtx := rest.RestContext{Codec: cdc, URL: "http://localhost:1317"}
	fmt.Printf("%+v\n", restCtx)

	if err := restCtx.TxSnapshotNow(w, ""); err != nil {
		return err
	}

	return nil
}
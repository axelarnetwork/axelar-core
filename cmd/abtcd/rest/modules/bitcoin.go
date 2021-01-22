package modules

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	bitcoinRest "github.com/axelarnetwork/axelar-core/x/bitcoin/client/rest"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
)

func (app *AppContext) TxBitcoinLink(chain string, address string) error {
	txReq := bitcoinRest.ReqLink{
		BaseReq: rest.PrepareBaseReq(&app.Wallet),
		Address: address,
	}

	// @TODO should use proper uri encoding
	// @TODO should build route from module's rest client consts
	txRoute := fmt.Sprintf("tx/bitcoin/link/%s", chain)

	return app.RestCtx.SubmitTx(&app.Wallet, txRoute, txReq)
}

func (app *AppContext) QueryDepositAddress(chain string, address string)  (string, error){
	restRoute := fmt.Sprintf("query/bitcoin/%s/%s/%s", keeper.QueryDepositAddress, chain, address)

	queryResp, err := app.RestCtx.RequestQuery(restRoute)

	var chainAddress bitcoinRest.CrossChainAddress
	err = app.RestCtx.Codec.UnmarshalJSON(queryResp.Result, &chainAddress)
	if err != nil {
		return "", err
	}

	return chainAddress.Address, err
}

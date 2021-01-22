package modules

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	bitcoinRest "github.com/axelarnetwork/axelar-core/x/bitcoin/client/rest"
)

func (app *AppContext) TxBitcoinLink(chain string, address string) error {
	txReq := bitcoinRest.ReqLink{
		BaseReq: rest.PrepareBaseReq(&app.Wallet),
		Address: address,
	}

	// @TODO should use proper uri encoding
	// @TODO should build route from module's rest client consts
	txRoute := fmt.Sprintf("tx/Bitcoin/link/%s", chain)

	return app.RestCtx.SubmitTx(&app.Wallet, txRoute, txReq)
}

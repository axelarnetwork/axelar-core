package modules

import (
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	snapshotRest "github.com/axelarnetwork/axelar-core/x/snapshot/client/rest"
)

func (app *AppContext) TxSnapshotNow() error {
	txReq := snapshotRest.ReqSnapshotNow{
		BaseReq: rest.PrepareBaseReq(&app.Wallet),
	}

	txRoute := "tx/snapshot/now"

	return app.RestCtx.SubmitTx(&app.Wallet, txRoute, txReq)
}
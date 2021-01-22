package modules

import (
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/wallet"
	snapshotRest "github.com/axelarnetwork/axelar-core/x/snapshot/client/rest"
)

func TxSnapshotNow(w *wallet.Wallet, rc *rest.RestContext) error {
	txReq := snapshotRest.ReqSnapshotNow{
		BaseReq: rest.PrepareBaseReq(w),
	}

	txRoute := "tx/snapshot/now"

	return rc.SubmitTx(w, txRoute, txReq)
}

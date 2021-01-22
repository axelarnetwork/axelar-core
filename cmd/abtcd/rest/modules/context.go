package modules

import (
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/rest"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/wallet"
)

type AppContext struct {
	Wallet wallet.Wallet
	RestCtx rest.RestContext
}
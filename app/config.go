package app

import (
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

type Config struct {
	btcTypes.BtcConfig `mapstructure:"axelar_bridge_btc"`
	types.ClientConfig `mapstructure:",squash"`
}

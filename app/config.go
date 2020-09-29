package app

import btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"

type Config struct {
	btcTypes.BtcConfig `mapstructure:"axelar_bridge_btc"`
}

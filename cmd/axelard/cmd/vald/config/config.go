package config

import (
	"time"

	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	bitcoin.BtcConfig `mapstructure:"axelar_bridge_btc"`
	tss.TssConfig     `mapstructure:",squash"`
	ClientConfig      `mapstructure:",squash"`

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		EVMConfig:    evm.DefaultConfig(),
		BtcConfig:    bitcoin.DefaultConfig(),
		TssConfig:    tss.DefaultConfig(),
		ClientConfig: ClientConfig{},
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	From              string        `mapstructure:"broadcaster-account"`
	KeyringPassphrase string        `mapstructure:"keyring-passphrase"`
	MaxRetries        int           `mapstructure:"max-retries"`
	MinTimeout        time.Duration `mapstructure:"min-timeout"`
}

// ClientConfig is the configuration for all client processes
type ClientConfig struct {
	KeyringBackend    string `mapstructure:"keyring-backend"`
	TendermintNodeURI string `mapstructure:"node"`
	ChainID           string `mapstructure:"chain-id"`
	BroadcastConfig   `mapstructure:"broadcast"`
}

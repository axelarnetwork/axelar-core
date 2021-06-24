package config

import "time"

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	From              string        `mapstructure:"broadcaster-account"`
	KeyringPassphrase string        `mapstructure:"keyring-passphrase"`
	GasAdjustment     float64       `mapstructure:"gas-adjustment"`
	Gas               uint64        `mapstructure:"gas"`
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

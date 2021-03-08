package types

type BroadcastConfig struct {
	From              string  `mapstructure:"broadcaster-account"`
	KeyringPassphrase string  `mapstructure:"keyring-passphrase"`
	GasAdjustment     float64 `mapstructure:"gas-adjustment"`
	Gas               int     `mapstructure:"gas"`
}

type ClientConfig struct {
	KeyringBackend    string `mapstructure:"keyring-backend"`
	TendermintNodeUri string `mapstructure:"node"`
	ChainID           string `mapstructure:"chain-id"`
	BroadcastConfig   `mapstructure:"broadcast"`
}

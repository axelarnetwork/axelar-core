package types

type EthConfig struct {
	EthRpcAddr    string `mapstructure:"eth_rpc_addr"`
	WithEthBridge bool   `mapstructure:"start-with-ethbridge"`
}

func DefaultConfig() EthConfig {
	return EthConfig{
		EthRpcAddr:    "http://127.0.0.1:7545",
		WithEthBridge: true,
	}
}

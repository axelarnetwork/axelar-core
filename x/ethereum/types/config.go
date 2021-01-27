package types

type EthConfig struct {
	EthRpcAddr    string `mapstructure:"rpc_addr"`
	WithEthBridge bool   `mapstructure:"start-with-bridge"`
}

func DefaultConfig() EthConfig {
	return EthConfig{
		EthRpcAddr:    "http://127.0.0.1:7545",
		WithEthBridge: true,
	}
}

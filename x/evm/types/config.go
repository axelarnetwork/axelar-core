package types

import "github.com/axelarnetwork/axelar-core/vald/evm/rpc"

// EVMConfig contains all EVM module configuration values
type EVMConfig struct {
	Name             string               `mapstructure:"name"`
	RPCAddr          string               `mapstructure:"rpc_addr"`
	WithBridge       bool                 `mapstructure:"start-with-bridge"`
	L1ChainName      *string              `mapstructure:"l1_chain_name"` // Deprecated: Do not use.
	FinalityOverride rpc.FinalityOverride `mapstructure:"finality_override"`
}

// DefaultConfig returns a configuration populated with default values
func DefaultConfig() []EVMConfig {
	return []EVMConfig{{
		Name:       "Ethereum",
		RPCAddr:    "http://127.0.0.1:7545",
		WithBridge: true,
	}}
}

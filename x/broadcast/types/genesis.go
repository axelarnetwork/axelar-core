package types

import "fmt"

type GenesisState struct {
	ProxyCount uint32
}

func DefaultGenesisState() GenesisState {
	return GenesisState{ProxyCount: 0}
}

func ValidateGenesis(g GenesisState) error {
	if g.ProxyCount != 0 {
		return fmt.Errorf("proxyCount must be 0")
	}
	return nil
}

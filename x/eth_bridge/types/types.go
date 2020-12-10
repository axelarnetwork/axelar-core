package types

import (
	"fmt"
)

type ExternalChainAddress struct {
	Chain   string
	Address string
}

func (addr ExternalChainAddress) IsInvalid() bool {
	return addr.Chain == "" || addr.Address == ""
}

func (addr ExternalChainAddress) String() string {
	return fmt.Sprintf("chain: %s, address: %s", addr.Chain, addr.Address)
}

/*
// This type provides additional functionality based on the bitcoin chain name
type Chain string

// Validate checks if the object is a valid chain
func (c Chain) Validate() error {
	switch string(c) {
	case chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name:
		return nil
	default:
		return fmt.Errorf("chain could not be parsed, choose %s or %s",
			chaincfg.MainNetParams.Name,
			chaincfg.TestNet3Params.Name,
		)
	}
}

// Params returns the configuration parameters associated with the chain
func (c Chain) Params() *chaincfg.Params {
	switch string(c) {
	case chaincfg.MainNetParams.Name:
		return &chaincfg.MainNetParams
	case chaincfg.TestNet3Params.Name:
		return &chaincfg.TestNet3Params
	default:
		return nil
	}
}
*/

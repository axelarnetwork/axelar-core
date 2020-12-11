package types

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Kovan   = "kovan"
	Rinkeby = "rinkby"
	Goerli  = "goerli"
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

// This type provides additional functionality based on the bitcoin chain name
type Chain string

// Validate checks if the object is a valid chain
func (c Chain) Validate() error {
	switch string(c) {
	case Mainnet, Ropsten, Kovan, Rinkeby, Goerli:
		return nil
	default:
		return fmt.Errorf("chain could not be parsed, choose %s, %s, %s, %s or %s",
			Mainnet,
			Ropsten,
			Kovan,
			Rinkeby,
			Goerli,
		)
	}
}

// Params returns the configuration parameters associated with the chain
func (c Chain) Params() *params.ChainConfig {
	switch string(c) {
	case Mainnet:
		return params.MainnetChainConfig
	case Ropsten:
		return params.RopstenChainConfig
	case Kovan:
		return params.RinkebyChainConfig
	case Goerli:
		return params.GoerliChainConfig
	default:
		return nil
	}
}

type TX struct {
	Hash    *common.Hash
	Amount  big.Int
	Address EthAddress
}

func (u TX) Validate() error {
	if u.Hash == nil {
		return fmt.Errorf("missing hash")
	}
	if u.Amount.Int64() <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if err := u.Address.Validate(); err != nil {
		return err
	}
	return nil
}

func (u TX) Equals(other TX) bool {
	return other.Validate() == nil &&
		bytes.Equal(u.Hash.Bytes(), other.Hash.Bytes()) &&
		bytes.Equal(u.Amount.Bytes(), other.Amount.Bytes()) &&
		u.Address == other.Address
}

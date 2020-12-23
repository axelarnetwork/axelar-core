package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Kovan   = "kovan"
	Rinkeby = "rinkeby"
	Goerli  = "goerli"

	// Ganache is a local testnet
	Ganache = "ganache"

	erc20Mint = "mint(address,uint256)"
)

var ERC20MintSel string

func init() {
	ERC20MintSel = CalcSelector(erc20Mint)
}

func CalcSelector(funcSignature string) string {

	hash := sha3.NewLegacyKeccak256()

	hash.Write([]byte(funcSignature))
	buf := hash.Sum(nil)

	return hexutil.Encode(buf[:4])
}

// This type provides additional functionality based on the ethereum network name
type Network string

// Validate checks if the object is a valid network
func (n Network) Validate() error {
	switch string(n) {
	case Mainnet, Ropsten, Kovan, Rinkeby, Goerli, Ganache:
		return nil
	default:
		return fmt.Errorf("network could not be parsed, choose %s, %s, %s, %s, %s or %s",
			Mainnet,
			Ropsten,
			Kovan,
			Rinkeby,
			Goerli,
			Ganache,
		)
	}
}

// Params returns the configuration parameters associated with the network
func (n Network) Params() *params.ChainConfig {
	switch string(n) {
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

type Tx struct {
	Hash        common.Hash
	Amount      sdk.Int
	ContractID  string
	Destination common.Address
	Network     Network
}

type TXType int

const (
	TypeSCDeploy TXType = iota
	TypeERC20mint
)

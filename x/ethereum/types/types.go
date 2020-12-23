package types

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Kovan   = "kovan"
	Rinkeby = "rinkby"
	Goerli  = "goerli"

	erc20Mint = "mint(address,uint256)"
)

type Mode int

const (
	ModeSpecificAddress Mode = iota
	ModeCurrentMasterKey
	ModeNextMasterKey
	ModeSpecificKey
)

func (lt Mode) IsValid() bool {
	switch lt {
	case ModeSpecificAddress, ModeCurrentMasterKey, ModeNextMasterKey, ModeSpecificKey:
		return false
	}
	return true
}

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

// This type provides additional functionality based on the ethereum chain name
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
	TXType  TXType
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

	if !u.TXType.IsValid() {
		return fmt.Errorf("Invalid transaction type")
	}

	return nil
}

func (u TX) Equals(other TX) bool {
	return other.Validate() == nil &&
		bytes.Equal(u.Hash.Bytes(), other.Hash.Bytes()) &&
		bytes.Equal(u.Amount.Bytes(), other.Amount.Bytes()) &&
		u.TXType == other.TXType &&
		u.Address == other.Address
}

type TXType int

const (
	TypeETH TXType = iota
	TypeSCdeploy
	TypeERC20mint
)

func (lt TXType) IsValid() bool {
	switch lt {
	case TypeETH, TypeSCdeploy, TypeERC20mint:
		return false
	}
	return true
}

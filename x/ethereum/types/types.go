package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Rinkeby = "rinkeby"
	Goerli  = "goerli"
	// Ganache is a local testnet
	Ganache = "ganache"

	erc20Mint = "mint(address,uint256)"
)

var (
	ERC20MintSel = CalcSelector(erc20Mint)
	networksByID = map[int64]Network{
		params.MainnetChainConfig.ChainID.Int64():       Mainnet,
		params.RopstenChainConfig.ChainID.Int64():       Ropsten,
		params.RinkebyChainConfig.ChainID.Int64():       Rinkeby,
		params.GoerliChainConfig.ChainID.Int64():        Goerli,
		params.AllCliqueProtocolChanges.ChainID.Int64(): Ganache,
	}
)

func CalcSelector(funcSignature string) string {
	hash := sha3.NewLegacyKeccak256()

	hash.Write([]byte(funcSignature))
	buf := hash.Sum(nil)

	return hexutil.Encode(buf[:4])
}

// This type provides additional functionality based on the ethereum network name
type Network string

func NetworkByID(id *big.Int) Network {
	return networksByID[id.Int64()]
}

// Validate checks if the object is a valid network
func (n Network) Validate() error {
	switch string(n) {
	case Mainnet, Ropsten, Rinkeby, Goerli, Ganache:
		return nil
	default:
		return fmt.Errorf("network could not be parsed, choose %s, %s, %s, %s or %s",
			Mainnet,
			Ropsten,
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
	case Rinkeby:
		return params.RinkebyChainConfig
	case Goerli:
		return params.GoerliChainConfig
	case Ganache:
		return params.AllCliqueProtocolChanges
	default:
		return nil
	}
}

type Signature [crypto.SignatureLength]byte

func ToEthSignature(sig exported.Signature, hash common.Hash, pk ecdsa.PublicKey) (Signature, error) {
	s := Signature{}
	copy(s[:32], common.LeftPadBytes(sig.R.Bytes(), 32))
	copy(s[32:], common.LeftPadBytes(sig.S.Bytes(), 32))
	// s[64] = 0 implicit

	derivedPK, err := crypto.SigToPub(hash.Bytes(), s[:])
	if err != nil {
		return Signature{}, err
	}

	if bytes.Equal(pk.Y.Bytes(), derivedPK.Y.Bytes()) {
		return s, nil
	}

	s[64] = 1

	return s, nil
}

type DeployParams struct {
	ByteCode []byte
	GasLimit uint64
}

type DeployResult struct {
	ContractAddress string
	Tx              []byte
}

type MintParams struct {
	Recipient    string
	Amount       sdk.Int
	ContractAddr string
	GasLimit     uint64
}

func CreateMintCallData(toAddr common.Address, amount *big.Int) []byte {
	paddedAddr := hexutil.Encode(common.LeftPadBytes(toAddr.Bytes(), 32))
	paddedVal := hexutil.Encode(common.LeftPadBytes(amount.Bytes(), 32))

	var data []byte

	data = append(data, common.FromHex(ERC20MintSel)...)
	data = append(data, common.FromHex(paddedAddr)...)
	data = append(data, common.FromHex(paddedVal)...)
	return data
}

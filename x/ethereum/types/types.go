package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
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

	erc20Mint            = "mint(address,uint256)"
	axelarGatewayExecute = "execute(bytes)"

	axelarGatewayCommandMint = "mintToken"
)

var (
	ERC20MintSel            = CalcSelector(erc20Mint)
	AxelarGatewayExecuteSel = CalcSelector(axelarGatewayExecute)
	networksByID            = map[int64]Network{
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

// Network provides additional functionality based on the ethereum network name
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
	Tx              *ethTypes.Transaction
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

func GetEthereumSignHash(data []byte) common.Hash {
	hash := crypto.Keccak256(data)
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(hash), hash)

	return crypto.Keccak256Hash([]byte(msg))
}

func CreateExecuteMintData(chainID *big.Int, commandID [32]byte, addresses []string, denoms []string, amounts []*big.Int) ([]byte, error) {
	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return nil, err
	}

	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	mintParams, err := createMintParams(addresses, denoms, amounts)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: uint256Type}, {Type: bytes32Type}, {Type: stringType}, {Type: bytesType}}
	result, err := arguments.Pack(
		chainID,
		commandID,
		axelarGatewayCommandMint,
		mintParams,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

/* This function would strip off anything in the strings beyond 32 bytes */
// TODO: Remove this function after https://github.com/axelarnetwork/ethereum-bridge/issues/3 is implemented
func stringArrToByte32Arr(stringArr []string) [][32]byte {
	var result [][32]byte

	for _, str := range stringArr {
		bytes := []byte(str)
		var byte32 [32]byte

		copy(byte32[:], bytes[:32])
		result = append(result, byte32)
	}

	return result
}

/* This function would strip off anything in the hex strings beyond 32 bytes */
func hexArrToByte32Arr(hexes []string) [][32]byte {
	var result [][32]byte

	for _, hex := range hexes {
		var byte32 [32]byte

		copy(byte32[:], common.LeftPadBytes(common.FromHex(hex), 32)[:32])
		result = append(result, byte32)
	}

	return result
}

func createMintParams(addresses []string, denoms []string, amounts []*big.Int) ([]byte, error) {
	length := len(addresses)

	if len(denoms) != length || len(amounts) != length {
		return nil, fmt.Errorf("addresses, denoms and amounts have different length")
	}

	bytes32ArrayType, err := abi.NewType("bytes32[]", "bytes32[]", nil)
	if err != nil {
		return nil, err
	}

	addressArrayType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return nil, err
	}

	uint256ArrayType, err := abi.NewType("uint256[]", "uint256[]", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: bytes32ArrayType}, {Type: addressArrayType}, {Type: uint256ArrayType}}
	result, err := arguments.Pack(
		stringArrToByte32Arr(denoms),
		hexArrToByte32Arr(addresses),
		amounts,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

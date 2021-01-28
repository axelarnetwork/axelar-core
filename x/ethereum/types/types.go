package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Ethereum network labels
const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Rinkeby = "rinkeby"
	Goerli  = "goerli"
	Ganache = "ganache"
)

const (
	// TODO: Check if there's a way to install the smart contract module with compiled ABI files
	axelarGatewayABI = `[
		{
			"inputs": [
				{
					"internalType": "bytes",
          "name": "input",
          "type": "bytes"
        }
			],
			"name": "execute",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`
	axelarGatewayCommandMint        = "mintToken"
	axelarGatewayCommandDeployToken = "deployToken"
	axelarGatewayFuncExecute        = "execute"
)

var (
	networksByID = map[int64]Network{
		params.MainnetChainConfig.ChainID.Int64():       Mainnet,
		params.RopstenChainConfig.ChainID.Int64():       Ropsten,
		params.RinkebyChainConfig.ChainID.Int64():       Rinkeby,
		params.GoerliChainConfig.ChainID.Int64():        Goerli,
		params.AllCliqueProtocolChanges.ChainID.Int64(): Ganache,
	}
)

// Network provides additional functionality based on the ethereum network name
type Network string

// NetworkByID looks up the Ethereum network corresponding to the given chain ID
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

// Signature encodes the parameters R,S,V in the byte format expected by Ethereum
type Signature [crypto.SignatureLength]byte

// ToEthSignature transforms a
func ToEthSignature(sig tss.Signature, hash common.Hash, pk ecdsa.PublicKey) (Signature, error) {
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

// DeployParams describe the parameter used to create a deploy contract transaction for Ethereum
type DeployParams struct {
	ByteCode []byte
	GasLimit uint64
}

// DeployResult describes the result of the deploy contract query,
// containing the raw unsigned transaction and the address to which the contract will be deployed
type DeployResult struct {
	ContractAddress string
	Tx              *ethTypes.Transaction
}

// CreateExecuteData wraps the specific command data and includes the command signature.
// Returns the data that goes into the data field of an Ethereum transaction
func CreateExecuteData(commandData []byte, commandSig Signature) ([]byte, error) {
	abiEncoder, err := abi.JSON(strings.NewReader(axelarGatewayABI))
	if err != nil {
		return nil, err
	}

	var homesteadCommandSig []byte
	homesteadCommandSig = append(homesteadCommandSig, commandSig[:]...)

	/* TODO: We have to make v 27 or 28 due to openzeppelin's implementation at https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/cryptography/ECDSA.sol
	requiring that. Consider copying and modifying it to require v to be just 0 or 1
	instead.
	*/
	if homesteadCommandSig[64] == 0 || homesteadCommandSig[64] == 1 {
		homesteadCommandSig[64] += 27
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: bytesType}, {Type: bytesType}}
	executeData, err := arguments.Pack(commandData, homesteadCommandSig)
	if err != nil {
		return nil, err
	}

	result, err := abiEncoder.Pack(axelarGatewayFuncExecute, executeData)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetEthereumSignHash returns the hash that needs to be signed so AxelarGateway accepts the given command
func GetEthereumSignHash(commandData []byte) common.Hash {
	hash := crypto.Keccak256(commandData)
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(hash), hash)

	return crypto.Keccak256Hash([]byte(msg))
}

// CreateMintCommandData returns the command data to mint tokens for the specified transfers
func CreateMintCommandData(chainID *big.Int, commandID CommandID, transfers []balance.CrossChainTransfer) ([]byte, error) {
	addresses, denoms, amounts := flatTransfers(transfers)
	mintParams, err := createMintParams(addresses, denoms, amounts)
	if err != nil {
		return nil, err
	}

	return packArguments(chainID, commandID, axelarGatewayCommandMint, mintParams)
}

// CreateDeployTokenCommandData returns the command data to deploy the specified token
func CreateDeployTokenCommandData(chainID *big.Int, commandID CommandID, tokenName string, symbol string, decimals uint8, capacity sdk.Int) ([]byte, error) {
	deployParams, err := createDeployTokenParams(tokenName, symbol, decimals, capacity.BigInt())
	if err != nil {
		return nil, err
	}

	return packArguments(chainID, commandID, axelarGatewayCommandDeployToken, deployParams)
}

// CommandID represents the unique command identifier
type CommandID [32]byte

// CalculateCommandID calculates the unique command ID that is used to protect from replay attacks in the AxelarGateway contract
// TODO: Remove this function after https://github.com/axelarnetwork/ethereum-bridge/issues/3 is implemented
func CalculateCommandID(transfers []balance.CrossChainTransfer) CommandID {
	var result CommandID
	var hash []byte

	for _, transfer := range transfers {
		idByte := make([]byte, 8)
		binary.LittleEndian.PutUint64(idByte, transfer.ID)

		hash = crypto.Keccak256(hash, idByte)
	}
	if hash == nil {
		return CommandID{}
	}

	copy(result[:], hash[:32])

	return result
}

func packArguments(chainID *big.Int, commandID CommandID, command string, commandParams []byte) ([]byte, error) {
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

	arguments := abi.Arguments{{Type: uint256Type}, {Type: bytes32Type}, {Type: stringType}, {Type: bytesType}}
	result, err := arguments.Pack(
		chainID,
		commandID,
		command,
		commandParams,
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
		bz := []byte(str)
		var byte32 [32]byte

		copy(byte32[:], bz[:32])
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

func createDeployTokenParams(tokenName string, symbol string, decimals uint8, capacity *big.Int) ([]byte, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	result, err := arguments.Pack(
		tokenName,
		symbol,
		decimals,
		capacity,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func flatTransfers(transfers []balance.CrossChainTransfer) ([]string, []string, []*big.Int) {
	var addresses []string
	var denoms []string
	var amounts []*big.Int

	for _, transfer := range transfers {
		addresses = append(addresses, transfer.Recipient.Address)
		denoms = append(denoms, transfer.Amount.Denom)
		amounts = append(amounts, transfer.Amount.Amount.BigInt())
	}

	return addresses, denoms, amounts
}

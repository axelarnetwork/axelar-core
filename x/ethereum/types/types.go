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

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
	axelarGatewayCommandMint              = "mintToken"
	axelarGatewayCommandDeployToken       = "deployToken"
	axelarGatewayCommandBurnToken         = "burnToken"
	axelarGatewayCommandTransferOwnership = "transferOwnership"
	axelarGatewayFuncExecute              = "execute"
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

// NetworkFromStr returns network given string
func NetworkFromStr(net string) (Network, error) {
	switch net {
	case "main":
		return Mainnet, nil
	case "ropsten":
		return Ropsten, nil
	case "rinkeby":
		return Rinkeby, nil
	case "goerli":
		return Goerli, nil
	case "ganache":
		return Ganache, nil
	default:
		return "", fmt.Errorf("unknown network: %s", net)
	}
}

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

// ToEthSignature transforms an Axelar generated signature into an ethereum recoverable signature
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

// DeployParams describe the parameters used to create a deploy contract transaction for Ethereum
type DeployParams struct {
	GasPrice sdk.Int
	GasLimit uint64
}

// DeployResult describes the result of the deploy contract query,
// containing the raw unsigned transaction and the address to which the contract will be deployed
type DeployResult struct {
	ContractAddress string                `json:"contract_address"`
	Tx              *ethTypes.Transaction `json:"tx"`
}

// SendTxResult describes the result of the send signed tx query,
// containing the signed transaction and the unsigned tx hash
type SendTxResult struct {
	TxID     string                `json:"tx_id"`
	SignedTx *ethTypes.Transaction `json:"tx"`
}

// CommandParams describe the parameters used to send a pre-signed command to the given contract,
// with the sender signing the transaction on the Ethereum node
type CommandParams struct {
	CommandID CommandID
	Sender    string
}

// ERC20TokenDeploy describes information about an ERC20 token
type ERC20TokenDeploy struct {
	Symbol    string
	TokenAddr string
}

// BurnerInfo describes information required to burn token at an burner address
// that is deposited by an user
type BurnerInfo struct {
	TokenAddr string
	Symbol    string
	Salt      [common.HashLength]byte
}

// ERC20Deposit contains information for an ERC20 deposit
type ERC20Deposit struct {
	TxID       common.Hash
	Amount     sdk.Uint
	Symbol     string
	BurnerAddr string
}

// DepositState is an enum for the state of a deposit
type DepositState int

// States of confirmed deposits
const (
	CONFIRMED DepositState = iota
	BURNED
)

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

func transferIDtoCommandID(transferID uint64) CommandID {
	var commandID CommandID

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, transferID)

	copy(commandID[:], common.LeftPadBytes(bz, 32)[:32])

	return commandID
}

// CreateMintCommandData returns the command data to mint tokens for the specified transfers
func CreateMintCommandData(chainID *big.Int, transfers []nexus.CrossChainTransfer) ([]byte, error) {
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	for _, transfer := range transfers {
		commandParam, err := createMintParams(transfer.Recipient.Address, transfer.Asset.Denom, transfer.Asset.Amount.BigInt())
		if err != nil {
			return nil, err
		}

		commandIDs = append(commandIDs, transferIDtoCommandID(transfer.ID))
		commands = append(commands, axelarGatewayCommandMint)
		commandParams = append(commandParams, commandParam)
	}

	return packArguments(chainID, commandIDs, commands, commandParams)
}

// CreateDeployTokenCommandData returns the command data to deploy the specified token
func CreateDeployTokenCommandData(chainID *big.Int, commandID CommandID, tokenName string, symbol string, decimals uint8, capacity sdk.Int) ([]byte, error) {
	deployParams, err := createDeployTokenParams(tokenName, symbol, decimals, capacity.BigInt())
	if err != nil {
		return nil, err
	}

	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	commandIDs = append(commandIDs, commandID)
	commands = append(commands, axelarGatewayCommandDeployToken)
	commandParams = append(commandParams, deployParams)

	return packArguments(chainID, commandIDs, commands, commandParams)
}

// CreateBurnCommandData returns the command data to burn tokens given burners' information
func CreateBurnCommandData(chainID *big.Int, height int64, burnerInfos []BurnerInfo) ([]byte, error) {
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))

	for _, burnerInfo := range burnerInfos {
		commandParam, err := createBurnTokenParams(burnerInfo.Symbol, common.BytesToHash(burnerInfo.Salt[:]))
		if err != nil {
			return nil, err
		}

		// TODO: A sequential ID for burns instead of hashing block height and salt together?
		commandID := CommandID(crypto.Keccak256Hash(append(burnerInfo.Salt[:], heightBytes...)))

		commandIDs = append(commandIDs, commandID)
		commands = append(commands, axelarGatewayCommandBurnToken)
		commandParams = append(commandParams, commandParam)
	}

	return packArguments(chainID, commandIDs, commands, commandParams)
}

// CreateTransferOwnershipCommandData returns the command data to transfer ownership of the contract
func CreateTransferOwnershipCommandData(chainID *big.Int, commandID CommandID, newOwnerAddr string) ([]byte, error) {
	transferOwnershipParams, err := createTransferOwnershipParams(newOwnerAddr)
	if err != nil {
		return nil, err
	}
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	commandIDs = append(commandIDs, commandID)
	commands = append(commands, axelarGatewayCommandTransferOwnership)
	commandParams = append(commandParams, transferOwnershipParams)

	return packArguments(chainID, commandIDs, commands, commandParams)
}

// CommandID represents the unique command identifier
type CommandID [32]byte

func packArguments(chainID *big.Int, commandIDs []CommandID, commands []string, commandParams [][]byte) ([]byte, error) {
	if len(commandIDs) != len(commands) || len(commandIDs) != len(commandParams) {
		return nil, fmt.Errorf("length mismatch for command arguments")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	bytes32ArrayType, err := abi.NewType("bytes32[]", "bytes32[]", nil)
	if err != nil {
		return nil, err
	}

	stringArrayType, err := abi.NewType("string[]", "string[]", nil)
	if err != nil {
		return nil, err
	}

	bytesArrayType, err := abi.NewType("bytes[]", "bytes[]", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: uint256Type}, {Type: bytes32ArrayType}, {Type: stringArrayType}, {Type: bytesArrayType}}
	result, err := arguments.Pack(
		chainID,
		commandIDs,
		commands,
		commandParams,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

/* This function would strip off anything in the hex strings beyond 32 bytes */
func hexToByte32(hex string) [32]byte {
	var result [32]byte
	copy(result[:], common.LeftPadBytes(common.FromHex(hex), 32)[:32])

	return result
}

func createMintParams(address string, denom string, amount *big.Int) ([]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}, {Type: uint256Type}}
	result, err := arguments.Pack(denom, hexToByte32(address), amount)
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

func createBurnTokenParams(symbol string, salt [32]byte) ([]byte, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: bytes32Type}}
	result, err := arguments.Pack(symbol, salt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createTransferOwnershipParams(newOwnerAddr string) ([]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	result, err := arguments.Pack(
		hexToByte32(newOwnerAddr),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

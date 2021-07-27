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
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

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

// Address wraps EVM Address
type Address common.Address

// Bytes returns the actual byte array of the address
func (a Address) Bytes() []byte {
	return common.Address(a).Bytes()
}

// Hex returns an EIP55-compliant hex string representation of the address
func (a Address) Hex() string {
	return common.Address(a).Hex()
}

// Marshal implements codec.ProtoMarshaler
func (a Address) Marshal() ([]byte, error) {
	return a[:], nil
}

// MarshalTo implements codec.ProtoMarshaler
func (a Address) MarshalTo(data []byte) (n int, err error) {
	bytesCopied := copy(data, a[:])
	if bytesCopied != common.AddressLength {
		return 0, fmt.Errorf("expected data size to be %d, actual %d", common.AddressLength, len(data))
	}

	return common.AddressLength, nil
}

// Unmarshal implements codec.ProtoMarshaler
func (a *Address) Unmarshal(data []byte) error {
	if len(data) != common.AddressLength {
		return fmt.Errorf("expected data size to be %d, actual %d", common.AddressLength, len(data))
	}

	*a = Address(common.BytesToAddress(data))

	return nil
}

// Size implements codec.ProtoMarshaler
func (a Address) Size() int {
	return common.AddressLength
}

// Hash wraps EVM Hash
type Hash common.Hash

// Bytes returns the actual byte array of the hash
func (h Hash) Bytes() []byte {
	return common.Hash(h).Bytes()
}

// Hex converts a hash to a hex string.
func (h Hash) Hex() string {
	return common.Hash(h).Hex()
}

// Marshal implements codec.ProtoMarshaler
func (h Hash) Marshal() ([]byte, error) {
	return h[:], nil
}

// MarshalTo implements codec.ProtoMarshaler
func (h Hash) MarshalTo(data []byte) (n int, err error) {
	bytesCopied := copy(data, h[:])
	if bytesCopied != common.HashLength {
		return 0, fmt.Errorf("expected data size to be %d, actual %d", common.HashLength, len(data))
	}

	return common.HashLength, nil
}

// Unmarshal implements codec.ProtoMarshaler
func (h *Hash) Unmarshal(data []byte) error {
	if len(data) != common.HashLength {
		return fmt.Errorf("expected data size to be %d, actual %d", common.HashLength, len(data))
	}

	*h = Hash(common.BytesToHash(data))

	return nil
}

// Size implements codec.ProtoMarshaler
func (h Hash) Size() int {
	return common.HashLength
}

// Signature encodes the parameters R,S,V in the byte format expected by an EVM chain
type Signature [crypto.SignatureLength]byte

// ToSignature transforms an Axelar generated signature into a recoverable signature
func ToSignature(sig tss.Signature, hash common.Hash, pk ecdsa.PublicKey) (Signature, error) {
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

// DeployParams describe the parameters used to create a deploy contract transaction for an EVM chain
type DeployParams struct {
	Chain    string
	GasPrice sdk.Int
	GasLimit uint64
}

// DeployResult describes the result of the deploy contract query,
// containing the raw unsigned transaction and the address to which the contract will be deployed
type DeployResult struct {
	ContractAddress string                `json:"contract_address"`
	Tx              *evmTypes.Transaction `json:"tx"`
}

// CommandParams describe the parameters used to send a pre-signed command to the given contract,
// with the sender signing the transaction on the node
type CommandParams struct {
	Chain     string
	CommandID CommandID
	Sender    string
}

// DepositState is an enum for the state of a deposit
type DepositState int

// States of confirmed deposits
const (
	CONFIRMED DepositState = iota
	BURNED
)

// CreateExecuteData wraps the specific command data and includes the command signature.
// Returns the data that goes into the data field of an EVM transaction
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

// GetSignHash returns the hash that needs to be signed so AxelarGateway accepts the given command
func GetSignHash(commandData []byte) common.Hash {
	hash := crypto.Keccak256(commandData)

	//TODO: is this the same across any EVM chain?
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
func CreateMintCommandData(chainID *big.Int, transfers []nexus.CrossChainTransfer, symbolRetriever func(denom string) (string, bool)) ([]byte, error) {
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	for _, transfer := range transfers {

		symbol, found := symbolRetriever(transfer.Asset.Denom)
		if !found {
			return nil, fmt.Errorf("could not find symbol for asset %s", transfer.Asset.Denom)
		}

		commandParam, err := createMintParams(common.HexToAddress(transfer.Recipient.Address), symbol, transfer.Asset.Amount.BigInt())
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
		commandParam, err := createBurnTokenParams(burnerInfo.Symbol, common.Hash(burnerInfo.Salt))
		if err != nil {
			return nil, err
		}

		// TODO: A sequential ID for burns instead of hashing block height and salt together?
		commandID := CommandID(crypto.Keccak256Hash(append(burnerInfo.Salt.Bytes(), heightBytes...)))

		commandIDs = append(commandIDs, commandID)
		commands = append(commands, axelarGatewayCommandBurnToken)
		commandParams = append(commandParams, commandParam)
	}

	return packArguments(chainID, commandIDs, commands, commandParams)
}

// CreateTransferOwnershipCommandData returns the command data to transfer ownership of the contract
func CreateTransferOwnershipCommandData(chainID *big.Int, commandID CommandID, newOwnerAddr common.Address) ([]byte, error) {
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

func createMintParams(address common.Address, denom string, amount *big.Int) ([]byte, error) {
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
	result, err := arguments.Pack(denom, address, amount)
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

func createBurnTokenParams(symbol string, salt common.Hash) ([]byte, error) {
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

func createTransferOwnershipParams(newOwnerAddr common.Address) ([]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	result, err := arguments.Pack(newOwnerAddr)
	if err != nil {
		return nil, err
	}

	return result, nil
}

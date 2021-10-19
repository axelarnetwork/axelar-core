package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Ethereum network labels
const (
	Mainnet = "mainnet"
	Ropsten = "ropsten"
	Rinkeby = "rinkeby"
	Goerli  = "goerli"
	Ganache = "ganache"
)

// AxelarGateway contract ABI and command selectors
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
	axelarGatewayCommandMintToken            = "mintToken"
	mintTokenMaxGasCost                      = 200000
	axelarGatewayCommandDeployToken          = "deployToken"
	deployTokenMaxGasCost                    = 1500000
	axelarGatewayCommandBurnToken            = "burnToken"
	burnTokenMaxGasCost                      = 200000
	axelarGatewayCommandTransferOwnership    = "transferOwnership"
	transferOwnershipMaxGasCost              = 150000
	axelarGatewayCommandTransferOperatorship = "transferOperatorship"
	transferOperatorshipMaxGasCost           = 150000
	axelarGatewayFuncExecute                 = "execute"
)

// ERC20Token represents an ERC20 token and its respective state
type ERC20Token struct {
	metadata ERC20TokenMetadata
	setMeta  func(meta ERC20TokenMetadata)
}

// CreateERC20Token returns an ERC20Token struct
func CreateERC20Token(setter func(meta ERC20TokenMetadata), meta ERC20TokenMetadata) ERC20Token {
	token := ERC20Token{
		metadata: meta,
		setMeta:  setter,
	}

	return token
}

// GetAsset returns the asset name
func (t *ERC20Token) GetAsset() string {
	return t.metadata.Asset
}

// GetTxID returns the tx ID set with StartConfirmation
func (t *ERC20Token) GetTxID() Hash {
	return t.metadata.TxHash
}

// GetDetails returns the details of the token
func (t *ERC20Token) GetDetails() TokenDetails {
	return t.metadata.Details
}

// Is returns true if the given status matches the token's status
func (t *ERC20Token) Is(status Status) bool {
	// this special case check is needed, because 0 & x == 0 is true for any x
	if status == NonExistent {
		return t.metadata.Status == NonExistent
	}
	return status&t.metadata.Status == status
}

// CreateDeployCommand returns a token deployment command for the token
func (t *ERC20Token) CreateDeployCommand(key tss.KeyID) (Command, error) {
	switch {
	case t.Is(NonExistent):
		return Command{}, fmt.Errorf("token %s non-existent", t.metadata.Asset)
	case t.Is(Confirmed):
		return Command{}, fmt.Errorf("token %s already confirmed", t.metadata.Asset)
	}
	if err := key.Validate(); err != nil {
		return Command{}, err
	}

	return CreateDeployTokenCommand(
		t.metadata.ChainID.BigInt(),
		key,
		t.metadata.Details,
	)
}

// CreateMintCommand returns a mint deployment command for the token
func (t *ERC20Token) CreateMintCommand(key tss.KeyID, transfer nexus.CrossChainTransfer) (Command, error) {
	if !t.Is(Confirmed) {
		return Command{}, fmt.Errorf("token %s not confirmed (current status: %s)",
			t.metadata.Asset, t.metadata.Status.String())
	}
	if err := key.Validate(); err != nil {
		return Command{}, err
	}

	return CreateMintTokenCommand(
		t.metadata.ChainID.BigInt(),
		key,
		transferIDtoCommandID(transfer.ID),
		t.metadata.Details.Symbol,
		common.HexToAddress(transfer.Recipient.Address),
		transfer.Asset.Amount.BigInt(),
	)
}

func transferIDtoCommandID(transferID uint64) CommandID {
	var commandID CommandID

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, transferID)

	copy(commandID[:], common.LeftPadBytes(bz, 32)[:32])

	return commandID
}

// GetAddress returns the token's address
func (t *ERC20Token) GetAddress() Address {
	return t.metadata.TokenAddress

}

// RecordDeployment signals that the token confirmation is underway for the given tx ID
func (t *ERC20Token) RecordDeployment(txID Hash) error {
	switch {
	case t.Is(NonExistent):
		return fmt.Errorf("token %s non-existent", t.metadata.Asset)
	case t.Is(Confirmed):
		return fmt.Errorf("token %s already confirmed", t.metadata.Asset)
	case t.Is(Pending):
		return fmt.Errorf("voting for token %s is already underway", t.metadata.Asset)
	}

	t.metadata.TxHash = txID
	t.metadata.Status |= Pending
	t.setMeta(t.metadata)

	return nil
}

// RejectDeployment reverts the token state back to Initialized
func (t *ERC20Token) RejectDeployment() error {
	switch {
	case t.Is(NonExistent):
		return fmt.Errorf("token %s non-existent", t.metadata.Asset)
	case !t.Is(Pending):
		return fmt.Errorf("token %s not waiting confirmation (current status: %s)", t.metadata.Asset, t.metadata.Status.String())
	}

	t.metadata.Status = Initialized
	t.metadata.TxHash = Hash{}
	t.setMeta(t.metadata)
	return nil
}

// ConfirmDeployment signals that the token was successfully confirmed
func (t *ERC20Token) ConfirmDeployment() error {
	switch {
	case t.Is(NonExistent):
		return fmt.Errorf("token %s non-existent", t.metadata.Asset)
	case !t.Is(Pending):
		return fmt.Errorf("token %s not waiting confirmation (current status: %s)", t.metadata.Asset, t.metadata.Status.String())
	}

	t.metadata.Status = Confirmed
	t.setMeta(t.metadata)

	return nil
}

// NilToken is a nil erc20 token
var NilToken = ERC20Token{}

// GetConfirmTokenKey creates a poll key for token confirmation
func GetConfirmTokenKey(txID Hash, asset string) vote.PollKey {
	return vote.NewPollKey(ModuleName, txID.Hex()+"_"+strings.ToLower(asset))
}

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

// CreateBurnTokenCommand creates a command to burn tokens with the given burner's information
func CreateBurnTokenCommand(chainID *big.Int, keyID tss.KeyID, height int64, burnerInfo BurnerInfo) (Command, error) {
	params, err := createBurnTokenParams(burnerInfo.Symbol, common.Hash(burnerInfo.Salt))
	if err != nil {
		return Command{}, err
	}

	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))

	return Command{
		ID:         NewCommandID(append(burnerInfo.Salt.Bytes(), heightBytes...), chainID),
		Command:    axelarGatewayCommandBurnToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: burnTokenMaxGasCost,
	}, nil
}

// CreateDeployTokenCommand creates a command to deploy a token
func CreateDeployTokenCommand(chainID *big.Int, keyID tss.KeyID, tokenDetails TokenDetails) (Command, error) {
	params, err := createDeployTokenParams(tokenDetails.TokenName, tokenDetails.Symbol, tokenDetails.Decimals, tokenDetails.Capacity.BigInt())
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         NewCommandID([]byte(tokenDetails.Symbol), chainID),
		Command:    axelarGatewayCommandDeployToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: deployTokenMaxGasCost,
	}, nil
}

// CreateMintTokenCommand creates a command to mint token to the given address
func CreateMintTokenCommand(chainID *big.Int, keyID tss.KeyID, id CommandID, symbol string, address common.Address, amount *big.Int) (Command, error) {
	params, err := createMintTokenParams(symbol, address, amount)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         id,
		Command:    axelarGatewayCommandMintToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: mintTokenMaxGasCost,
	}, nil
}

// CreateTransferOwnershipCommand creates a command to transfer ownership of the contract
func CreateTransferOwnershipCommand(chainID *big.Int, keyID tss.KeyID, newOwnerAddr common.Address) (Command, error) {
	params, err := createTransferOwnershipParams(newOwnerAddr)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         NewCommandID(newOwnerAddr.Bytes(), chainID),
		Command:    axelarGatewayCommandTransferOwnership,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: transferOwnershipMaxGasCost,
	}, nil
}

// CreateTransferOperatorshipCommand creates a command to transfer operatorship of the contract
func CreateTransferOperatorshipCommand(chainID *big.Int, keyID tss.KeyID, newOperatorAddr common.Address) (Command, error) {
	params, err := createTransferOperatorshipParams(newOperatorAddr)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         NewCommandID(newOperatorAddr.Bytes(), chainID),
		Command:    axelarGatewayCommandTransferOperatorship,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: transferOperatorshipMaxGasCost,
	}, nil
}

// GetGatewayDeploymentBytecode returns the deployment bytecode for the gateway contract
func GetGatewayDeploymentBytecode(contractBytecode []byte, admins []common.Address, threshold uint8, owner common.Address, operator common.Address) ([]byte, error) {
	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return nil, err
	}

	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{{Type: addressesType}, {Type: uint8Type}, {Type: addressType}, {Type: addressType}}
	argBytes, err := args.Pack(admins, threshold, owner, operator)
	if err != nil {
		return nil, err
	}

	return append(contractBytecode, argBytes...), nil
}

// Clone returns an exacy copy of Command
func (c Command) Clone() Command {
	var clone Command

	clone.Command = c.Command
	clone.ID = c.ID
	clone.KeyID = c.KeyID
	clone.Params = make([]byte, len(c.Params))
	copy(clone.Params, c.Params)

	return clone
}

// CommandBatch represents a batch of commands
type CommandBatch struct {
	metadata CommandBatchMetadata
	setter   func(batch CommandBatchMetadata)
}

// NewCommandBatch returns a new command batch struct
func NewCommandBatch(metadata CommandBatchMetadata, setter func(batch CommandBatchMetadata)) CommandBatch {
	return CommandBatch{
		metadata: metadata,
		setter:   setter,
	}
}

// GetPrevBatchedCommandsID returns the batch that preceeds this one
func (b CommandBatch) GetPrevBatchedCommandsID() []byte {
	return b.metadata.PrevBatchedCommandsID
}

// GetStatus returns the batch's status
func (b CommandBatch) GetStatus() BatchedCommandsStatus {
	return b.metadata.Status
}

// GetData returns the batch's data
func (b CommandBatch) GetData() []byte {
	return b.metadata.Data
}

// GetID returns the batch ID
func (b CommandBatch) GetID() []byte {
	return b.metadata.ID

}

// GetKeyID returns the batch's key ID
func (b CommandBatch) GetKeyID() tss.KeyID {
	return b.metadata.KeyID

}

// GetSigHash returns the batch's key ID
func (b CommandBatch) GetSigHash() Hash {
	return b.metadata.SigHash

}

// Is returns true if batched commands is in the given status; false otherwise
func (b CommandBatch) Is(status BatchedCommandsStatus) bool {
	return b.metadata.Status == status
}

// SetStatus sets the status for the batch, returning true if the status was updated
func (b *CommandBatch) SetStatus(status BatchedCommandsStatus) bool {
	if b.metadata.Status != BatchNonExistent && b.metadata.Status != BatchSigned {
		b.metadata.Status = status
		b.setter(b.metadata)
		return true
	}

	return false
}

// NewCommandBatchMetadata assembles a CommandBatchMetadata struct from the provided arguments
func NewCommandBatchMetadata(chainID *big.Int, keyID tss.KeyID, cmds []Command) (CommandBatchMetadata, error) {
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	for _, cmd := range cmds {
		commandIDs = append(commandIDs, cmd.ID)
		commands = append(commands, cmd.Command)
		commandParams = append(commandParams, cmd.Params)
	}

	data, err := packArguments(chainID, commandIDs, commands, commandParams)
	if err != nil {
		return CommandBatchMetadata{}, err
	}

	return CommandBatchMetadata{
		ID:         crypto.Keccak256(data),
		CommandIDs: commandIDs,
		Data:       data,
		SigHash:    Hash(GetSignHash(data)),
		Status:     BatchSigning,
		KeyID:      keyID,
	}, nil
}

const commandIDSize = 32

// CommandID represents the unique command identifier
type CommandID [commandIDSize]byte

// NewCommandID is the constructor for CommandID
func NewCommandID(data []byte, chainID *big.Int) CommandID {
	var commandID CommandID
	copy(commandID[:], crypto.Keccak256(append(data, chainID.Bytes()...))[:commandIDSize])

	return commandID
}

// Hex returns the hex representation of command ID
func (c CommandID) Hex() string {
	return hex.EncodeToString(c[:])
}

// Size implements codec.ProtoMarshaler
func (c CommandID) Size() int {
	return commandIDSize
}

// Marshal implements codec.ProtoMarshaler
func (c CommandID) Marshal() ([]byte, error) {
	return c[:], nil
}

// MarshalTo implements codec.ProtoMarshaler
func (c CommandID) MarshalTo(data []byte) (n int, err error) {
	bytesCopied := copy(data, c[:])
	if bytesCopied != commandIDSize {
		return 0, fmt.Errorf("expected data size to be %d, actual %d", commandIDSize, len(data))
	}

	return commandIDSize, nil
}

// Unmarshal implements codec.ProtoMarshaler
func (c *CommandID) Unmarshal(data []byte) error {
	bytesCopied := copy(c[:], data)
	if bytesCopied != commandIDSize {
		return fmt.Errorf("expected data size to be %d, actual %d", commandIDSize, len(data))
	}

	return nil
}

// TransferKeyTypeFromSimpleStr converts a given string into TransferKeyType
func TransferKeyTypeFromSimpleStr(str string) (TransferKeyType, error) {
	switch strings.ToLower(str) {
	case Ownership.SimpleString():
		return Ownership, nil
	case Operatorship.SimpleString():
		return Operatorship, nil
	default:
		return -1, fmt.Errorf("invalid transfer key type %s", str)
	}
}

// Validate returns an error if the TransferKeyType is invalid; nil otherwise
func (t TransferKeyType) Validate() error {
	switch t {
	case Ownership, Operatorship:
		return nil
	default:
		return fmt.Errorf("invalid transfer key type")
	}
}

// SimpleString returns a human-readable string representing the TransferKeyType
func (t TransferKeyType) SimpleString() string {
	switch t {
	case Ownership:
		return "transfer_ownership"
	case Operatorship:
		return "transfer_operatorship"
	default:
		return "unknown"
	}
}

// NewAsset returns a new Asset instance
func NewAsset(chain, name string) Asset {
	return Asset{
		Chain: chain,
		Name:  name,
	}
}

// Validate ensures that all fields are filled with sensible values
func (m Asset) Validate() error {
	if m.Chain == "" {
		return fmt.Errorf("missing asset chain")
	}
	if m.Name == "" {
		return fmt.Errorf("missing asset name")
	}
	return nil
}

// NewTokenDetails returns a new TokenDetails instance
func NewTokenDetails(tokenName, symbol string, decimals uint8, capacity sdk.Int) TokenDetails {
	return TokenDetails{
		TokenName: tokenName,
		Symbol:    symbol,
		Decimals:  decimals,
		Capacity:  capacity,
	}
}

// Validate ensures that all fields are filled with sensible values
func (m TokenDetails) Validate() error {
	if m.TokenName == "" {
		return fmt.Errorf("missing token name")
	}
	if m.Symbol == "" {
		return fmt.Errorf("missing token symbol")
	}
	if !m.Capacity.IsPositive() {
		return fmt.Errorf("token capacity must be a positive number")
	}

	return nil
}

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

func createMintTokenParams(symbol string, address common.Address, amount *big.Int) ([]byte, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}, {Type: uint256Type}}
	result, err := arguments.Pack(symbol, address, amount)
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

func createTransferOperatorshipParams(newOperatorAddr common.Address) ([]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	result, err := arguments.Pack(newOperatorAddr)
	if err != nil {
		return nil, err
	}

	return result, nil
}

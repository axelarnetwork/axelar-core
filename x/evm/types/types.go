package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
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

// Burner code hashes
const (
	// BurnerCodeHashV1 is the hash of the bytecode of burner v1
	BurnerCodeHashV1 = "0x70be6eedec1d63b7cf8b9233615e4e408c99e0753be123b605aa5d53ed4a8670"
	// BurnerCodeHashV2 is the hash of the bytecode of burner v2
	BurnerCodeHashV2 = "0xf34c56593ef4a993c05acac98bf4ae170ee322068752b49fb44ce545d29c3c6f"
)

func validateBurnerCode(burnerCode []byte) error {
	burnerCodeHash := crypto.Keccak256Hash(burnerCode).Hex()
	switch burnerCodeHash {
	case BurnerCodeHashV1:
	case BurnerCodeHashV2:
	default:
		return fmt.Errorf("unsupported burner code with hash %s", burnerCodeHash)
	}

	return nil
}

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
	AxelarGatewayCommandMintToken            = "mintToken"
	mintTokenMaxGasCost                      = 200000
	AxelarGatewayCommandDeployToken          = "deployToken"
	deployTokenMaxGasCost                    = 1500000
	AxelarGatewayCommandBurnToken            = "burnToken"
	burnTokenMaxGasCost                      = 200000
	AxelarGatewayCommandTransferOwnership    = "transferOwnership"
	transferOwnershipMaxGasCost              = 150000
	AxelarGatewayCommandTransferOperatorship = "transferOperatorship"
	transferOperatorshipMaxGasCost           = 150000
	axelarGatewayFuncExecute                 = "execute"
)

type role uint8

const (
	roleOwner    role = 1
	roleOperator role = 2
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

// IsExternal returns true if the given token is external; false otherwise
func (t ERC20Token) IsExternal() bool {
	return t.metadata.IsExternal
}

// SaveBurnerCode saves the burner code; panic if already saved since it should only be used during in-place storage migration
func (t ERC20Token) SaveBurnerCode(burnerCode []byte) {
	if len(t.metadata.BurnerCode) > 0 {
		panic(fmt.Errorf("burner code already set"))
	}

	t.metadata.BurnerCode = burnerCode
	t.setMeta(t.metadata)
}

// GetBurnerCode returns the version of the burner the token is deployed with
func (t ERC20Token) GetBurnerCode() []byte {
	return t.metadata.BurnerCode
}

// GetBurnerCodeHash returns the version of the burner the token is deployed with
func (t ERC20Token) GetBurnerCodeHash() Hash {
	return Hash(crypto.Keccak256Hash(t.metadata.BurnerCode))
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

	if t.IsExternal() {
		return CreateDeployTokenCommand(
			t.metadata.ChainID.BigInt(),
			key,
			t.metadata.Details,
			t.GetAddress(),
		)
	}

	return CreateDeployTokenCommand(
		t.metadata.ChainID.BigInt(),
		key,
		t.metadata.Details,
		ZeroAddress,
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

	return CreateMintTokenCommand(key, TransferIDtoCommandID(transfer.ID), t.metadata.Details.Symbol, common.HexToAddress(transfer.Recipient.Address), transfer.Asset.Amount.BigInt())
}

// TransferIDtoCommandID converts a transferID to a commandID
func TransferIDtoCommandID(transferID nexus.TransferID) CommandID {
	var commandID CommandID
	copy(commandID[:], common.LeftPadBytes(transferID.Bytes(), 32)[:32])

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

// GetConfirmGatewayDeploymentPollKey creates a poll key for the gateway deployment
func GetConfirmGatewayDeploymentPollKey(chain nexus.Chain, txID Hash, address Address) vote.PollKey {
	return vote.NewPollKey(ModuleName, fmt.Sprintf("%s_%s_%s", chain.Name, txID.Hex(), address.Hex()))
}

// GetConfirmTokenKey creates a poll key for token confirmation
func GetConfirmTokenKey(txID Hash, asset string) vote.PollKey {
	return vote.NewPollKey(ModuleName, txID.Hex()+"_"+strings.ToLower(asset))
}

// Address wraps EVM Address
type Address common.Address

// ZeroAddress represents an evm address with all bytes being zero
var ZeroAddress = Address{}

// IsZeroAddress returns true if the address contains only zero bytes; false otherwise
func (a Address) IsZeroAddress() bool {
	return bytes.Equal(a.Bytes(), ZeroAddress.Bytes())
}

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
func ToSignature(sig btcec.Signature, hash common.Hash, pk ecdsa.PublicKey) (Signature, error) {
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

// KeysToAddresses converts a slice of ECDSA public keys to evm addresses
func KeysToAddresses(keys ...ecdsa.PublicKey) []common.Address {
	addresses := make([]common.Address, len(keys))

	for i, key := range keys {
		addresses[i] = crypto.PubkeyToAddress(key)
	}

	return addresses
}

func toHomesteadSig(sig Signature) []byte {
	/* TODO: We have to make v 27 or 28 due to openzeppelin's implementation at https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/cryptography/ECDSA.sol
	requiring that. Consider copying and modifying it to require v to be just 0 or 1
	instead.
	*/
	bz := sig[:]
	if bz[crypto.SignatureLength-1] == 0 || bz[crypto.SignatureLength-1] == 1 {
		bz[crypto.SignatureLength-1] += 27
	}

	return bz
}

// CreateExecuteDataSinglesig wraps the specific command data and includes the command signature.
// Returns the data that goes into the data field of an EVM transaction
func CreateExecuteDataSinglesig(data []byte, sig Signature) ([]byte, error) {
	abiEncoder, err := abi.JSON(strings.NewReader(axelarGatewayABI))
	if err != nil {
		return nil, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: bytesType}, {Type: bytesType}}
	executeData, err := arguments.Pack(data, toHomesteadSig(sig))
	if err != nil {
		return nil, err
	}

	result, err := abiEncoder.Pack(axelarGatewayFuncExecute, executeData)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateExecuteDataMultisig wraps the specific command data and includes the command signatures.
// Returns the data that goes into the data field of an EVM transaction
func CreateExecuteDataMultisig(data []byte, sigs ...Signature) ([]byte, error) {
	abiEncoder, err := abi.JSON(strings.NewReader(axelarGatewayABI))
	if err != nil {
		return nil, err
	}

	var homesteadSigs [][]byte
	for _, sig := range sigs {
		homesteadSigs = append(homesteadSigs, toHomesteadSig(sig))
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	bytesArrayType, err := abi.NewType("bytes[]", "bytes[]", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: bytesType}, {Type: bytesArrayType}}
	executeData, err := arguments.Pack(data, homesteadSigs)
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

	// TODO: is this the same across any EVM chain?
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
		Command:    AxelarGatewayCommandBurnToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: burnTokenMaxGasCost,
	}, nil
}

// CreateDeployTokenCommand creates a command to deploy a token
func CreateDeployTokenCommand(chainID *big.Int, keyID tss.KeyID, tokenDetails TokenDetails, address Address) (Command, error) {
	params, err := createDeployTokenParams(tokenDetails.TokenName, tokenDetails.Symbol, tokenDetails.Decimals, tokenDetails.Capacity.BigInt(), address)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         NewCommandID([]byte(tokenDetails.Symbol), chainID),
		Command:    AxelarGatewayCommandDeployToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: deployTokenMaxGasCost,
	}, nil
}

// CreateMintTokenCommand creates a command to mint token to the given address
func CreateMintTokenCommand(keyID tss.KeyID, id CommandID, symbol string, address common.Address, amount *big.Int) (Command, error) {
	params, err := createMintTokenParams(symbol, address, amount)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         id,
		Command:    AxelarGatewayCommandMintToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: mintTokenMaxGasCost,
	}, nil
}

// CreateSinglesigTransferCommand creates a command to transfer ownership/operator of the singlesig contract
func CreateSinglesigTransferCommand(
	transferType TransferKeyType,
	chainID *big.Int,
	keyID tss.KeyID,
	address common.Address) (Command, error) {
	params, err := createTransferSinglesigParams(address)
	if err != nil {
		return Command{}, err
	}

	return createTransferCmd(NewCommandID(address.Bytes(), chainID), params, keyID, transferType)
}

// CreateMultisigTransferCommand creates a command to transfer ownership/operator of the multisig contract
func CreateMultisigTransferCommand(
	transferType TransferKeyType,
	chainID *big.Int,
	keyID tss.KeyID,
	threshold uint8,
	addresses ...common.Address) (Command, error) {

	if len(addresses) <= 0 {
		return Command{}, fmt.Errorf("transfer ownership command requires at least one key (received %d)", len(addresses))
	}

	var concat []byte
	for _, addr := range addresses {
		concat = append(concat, addr.Bytes()...)
	}

	params, err := createTransferMultisigParams(addresses, threshold)
	if err != nil {
		return Command{}, err
	}

	return createTransferCmd(NewCommandID(concat, chainID), params, keyID, transferType)
}

func createTransferCmd(id CommandID, params []byte, keyID tss.KeyID, transferType TransferKeyType) (Command, error) {
	switch transferType {
	case Ownership:
		return Command{
			ID:         id,
			Command:    AxelarGatewayCommandTransferOwnership,
			Params:     params,
			KeyID:      keyID,
			MaxGasCost: transferOwnershipMaxGasCost,
		}, nil
	case Operatorship:
		return Command{
			ID:         id,
			Command:    AxelarGatewayCommandTransferOperatorship,
			Params:     params,
			KeyID:      keyID,
			MaxGasCost: transferOperatorshipMaxGasCost,
		}, nil
	default:
		return Command{}, fmt.Errorf("invalid transfer key type %s", transferType.SimpleString())
	}
}

// GetSinglesigGatewayDeploymentBytecode returns the deployment bytecode for the singlesig gateway contract
func GetSinglesigGatewayDeploymentBytecode(contractBytecode []byte, admins []common.Address, threshold uint8, owner common.Address, operator common.Address) ([]byte, error) {
	if len(contractBytecode) == 0 {
		return nil, fmt.Errorf("contract bytecode cannot be empty bytes")
	}

	if threshold == 0 {
		return nil, fmt.Errorf("admin threshold must be >0")
	}

	if len(admins) < int(threshold) {
		return nil, fmt.Errorf("not enought admins")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
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

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{{Type: addressesType}, {Type: uint256Type}, {Type: addressType}, {Type: addressType}}
	argBytes, err := args.Pack(admins, big.NewInt(int64(threshold)), owner, operator)
	if err != nil {
		return nil, err
	}

	argBytes, err = abi.Arguments{{Type: bytesType}}.Pack(argBytes)
	if err != nil {
		return nil, err
	}

	return append(contractBytecode, argBytes...), nil
}

// GetMultisigGatewayDeploymentBytecode returns the deployment bytecode for the multisig gateway contract
func GetMultisigGatewayDeploymentBytecode(contractBytecode []byte, admins []common.Address, adminThreshold uint8, owners []common.Address, ownerThreshold uint8, operators []common.Address, operatorThreshold uint8) ([]byte, error) {
	if len(contractBytecode) == 0 {
		return nil, fmt.Errorf("contract bytecode cannot be empty bytes")
	}

	if adminThreshold == 0 {
		return nil, fmt.Errorf("admin threshold must be >0")
	}

	if len(admins) < int(adminThreshold) {
		return nil, fmt.Errorf("not enought admins")
	}

	if ownerThreshold == 0 {
		return nil, fmt.Errorf("owner threshold must be >0")
	}

	if len(owners) < int(ownerThreshold) {
		return nil, fmt.Errorf("not enought owners")
	}

	if operatorThreshold == 0 {
		return nil, fmt.Errorf("operator threshold must be >0")
	}

	if len(operators) < int(operatorThreshold) {
		return nil, fmt.Errorf("not enought operators")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return nil, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{{Type: addressesType}, {Type: uint256Type}, {Type: addressesType}, {Type: uint256Type}, {Type: addressesType}, {Type: uint256Type}}
	argBytes, err := args.Pack(admins, big.NewInt(int64(adminThreshold)), owners, big.NewInt(int64(ownerThreshold)), operators, big.NewInt(int64(operatorThreshold)))
	if err != nil {
		return nil, err
	}

	argBytes, err = abi.Arguments{{Type: bytesType}}.Pack(argBytes)
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

// NonExistentCommand can be used to represent a non-existent command
var NonExistentCommand = NewCommandBatch(CommandBatchMetadata{}, func(CommandBatchMetadata) {})

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

// GetCommandIDs returns the IDs of the commands included in the batch
func (b CommandBatch) GetCommandIDs() []CommandID {
	return b.metadata.CommandIDs
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
func NewCommandBatchMetadata(chainID *big.Int, keyID tss.KeyID, keyRole tss.KeyRole, cmds []Command) (CommandBatchMetadata, error) {
	var r role
	var commandIDs []CommandID
	var commands []string
	var commandParams [][]byte

	switch keyRole {
	case tss.MasterKey:
		r = roleOwner
	case tss.SecondaryKey:
		r = roleOperator
	default:
		return CommandBatchMetadata{}, fmt.Errorf("cannot sign command batch with a key of role %s", keyRole.SimpleString())
	}

	for _, cmd := range cmds {
		commandIDs = append(commandIDs, cmd.ID)
		commands = append(commands, cmd.Command)
		commandParams = append(commandParams, cmd.Params)
	}

	data, err := packArguments(chainID, r, commandIDs, commands, commandParams)
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

// HexToCommandID decodes an hex representation of a CommandID
func HexToCommandID(id string) (CommandID, error) {
	bz, err := hex.DecodeString(id)
	if err != nil {
		return CommandID{}, err
	}

	var commandID CommandID
	copy(commandID[:], bz)

	return commandID, nil
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
		Chain: utils.NormalizeString(chain),
		Name:  utils.NormalizeString(name),
	}
}

// Validate ensures that all fields are filled with sensible values
func (m Asset) Validate() error {
	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := utils.ValidateString(m.Name); err != nil {
		return sdkerrors.Wrap(err, "invalid name")
	}

	return nil
}

// NewTokenDetails returns a new TokenDetails instance
func NewTokenDetails(tokenName, symbol string, decimals uint8, capacity sdk.Int) TokenDetails {
	return TokenDetails{
		TokenName: utils.NormalizeString(tokenName),
		Symbol:    utils.NormalizeString(symbol),
		Decimals:  decimals,
		Capacity:  capacity,
	}
}

// Validate ensures that all fields are filled with sensible values
func (m TokenDetails) Validate() error {
	if err := utils.ValidateString(m.TokenName); err != nil {
		return sdkerrors.Wrap(err, "invalid token name")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid token symbol")
	}

	if m.Capacity.IsNil() || m.Capacity.IsNegative() {
		return fmt.Errorf("token capacity must be a non-negative number")
	}

	return nil
}

func packArguments(chainID *big.Int, r role, commandIDs []CommandID, commands []string, commandParams [][]byte) ([]byte, error) {
	if len(commandIDs) != len(commands) || len(commandIDs) != len(commandParams) {
		return nil, fmt.Errorf("length mismatch for command arguments")
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
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

	arguments := abi.Arguments{{Type: uint256Type}, {Type: uint8Type}, {Type: bytes32ArrayType}, {Type: stringArrayType}, {Type: bytesArrayType}}
	result, err := arguments.Pack(
		chainID,
		r,
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

// DecodeMintTokenParams unpacks the parameters of a mint token command
func DecodeMintTokenParams(bz []byte) (string, common.Address, *big.Int, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", common.Address{}, nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return "", common.Address{}, nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return "", common.Address{}, nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: addressType}, {Type: uint256Type}}
	params, err := arguments.Unpack(bz)
	if err != nil {
		return "", common.Address{}, nil, err
	}

	return params[0].(string), params[1].(common.Address), params[2].(*big.Int), nil
}

func createDeployTokenParams(tokenName string, symbol string, decimals uint8, capacity *big.Int, address Address) ([]byte, error) {
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

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}, {Type: addressType}}
	result, err := arguments.Pack(
		tokenName,
		symbol,
		decimals,
		capacity,
		address,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DecodeDeployTokenParams unpacks the parameters of a deploy token command
func DecodeDeployTokenParams(bz []byte) (string, string, uint8, *big.Int, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", "", 0, nil, err
	}

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return "", "", 0, nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return "", "", 0, nil, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	params, err := arguments.Unpack(bz)
	if err != nil {
		return "", "", 0, nil, err
	}

	return params[0].(string), params[1].(string), params[2].(uint8), params[3].(*big.Int), nil
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

// DecodeBurnTokenParams unpacks the parameters of a burn token command
func DecodeBurnTokenParams(bz []byte) (string, common.Hash, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", common.Hash{}, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return "", common.Hash{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: bytes32Type}}
	params, err := arguments.Unpack(bz)
	if err != nil {
		return "", common.Hash{}, err
	}

	return params[0].(string), params[1].([common.HashLength]byte), nil
}

func createTransferSinglesigParams(addr common.Address) ([]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	result, err := arguments.Pack(addr)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DecodeTransferSinglesigParams unpacks the parameters of a single sig transfer command
func DecodeTransferSinglesigParams(bz []byte) (common.Address, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return common.Address{}, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	params, err := arguments.Unpack(bz)
	if err != nil {
		return common.Address{}, err
	}

	return params[0].(common.Address), nil
}

func createTransferMultisigParams(addrs []common.Address, threshold uint8) ([]byte, error) {
	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256Type}}
	result, err := arguments.Pack(addrs, big.NewInt(int64(threshold)))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DecodeTransferMultisigParams unpacks the parameters of a multi sig transfer command
func DecodeTransferMultisigParams(bz []byte) ([]common.Address, uint8, error) {
	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return []common.Address{}, 0, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return []common.Address{}, 0, err
	}

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256Type}}
	params, err := arguments.Unpack(bz)
	if err != nil {
		return []common.Address{}, 0, err
	}

	return params[0].([]common.Address), uint8(params[1].(*big.Int).Uint64()), nil
}

// ValidateBasic does stateless validation of the object
func (m *BurnerInfo) ValidateBasic() error {
	if err := utils.ValidateString(m.DestinationChain); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if err := sdk.ValidateDenom(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	return nil
}

// ValidateBasic does stateless validation of the object
func (m *ERC20TokenMetadata) ValidateBasic() error {
	if m.Status == NonExistent {
		return fmt.Errorf("token status not set")
	}

	if err := sdk.ValidateDenom(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if m.ChainID.IsNil() || !m.ChainID.IsPositive() {
		return fmt.Errorf("chain ID not set")
	}

	if err := m.Details.Validate(); err != nil {
		return err
	}

	if err := validateBurnerCode(m.BurnerCode); err != nil {
		return err
	}

	return nil
}

// ValidateBasic does stateless validation of the object
func (m *ERC20Deposit) ValidateBasic() error {
	if err := sdk.ValidateDenom(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if err := utils.ValidateString(m.DestinationChain); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("amount must be >0")
	}

	return nil
}

// CommandIDsToStrings converts a slice of type CommandID to a slice of strings
func CommandIDsToStrings(commandIDs []CommandID) []string {
	commandList := make([]string, len(commandIDs))
	for i, commandID := range commandIDs {
		commandList[i] = commandID.Hex()
	}

	return commandList
}

// ValidateCommandQueueState checks if the keys of the given map have the correct format to be imported as command queue state.
// The expected format is {block height}_{[a-zA-Z0-9]+}
func ValidateCommandQueueState(state map[string]codec.ProtoMarshaler) error {
	for key := range state {
		keyParticles := strings.Split(key, utils.DefaultDelimiter)
		if len(keyParticles) != 2 {
			return fmt.Errorf("expected key %s to consist of two parts", key)
		}

		if _, err := strconv.ParseInt(keyParticles[0], 10, 64); err != nil {
			return fmt.Errorf("expected first key part of %s to be a block height", key)
		}
	}

	return nil
}

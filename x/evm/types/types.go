package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var _ codectypes.UnpackInterfacesMessage = CommandBatchMetadata{}

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
	BurnerCodeHashV2 = "0x49c166661e31e0bf5434d891dea1448dc35f6ecd54a0d88594df06e24effe7c2"
	// BurnerCodeHashV3 is the hash of the bytecode of burner v3
	BurnerCodeHashV3 = "0xa50851cafd39f2f61171c0c00a11bda820ed0958950df5a53ba11a047402351f"
	// BurnerCodeHashV4 is the hash of the bytecode of burner v4
	BurnerCodeHashV4 = "0x701d8db26f2d668fee8acf2346199a6b63b0173f212324d1c5a04b4d4de95666"
	// BurnerCodeHashV5 is the hash of the bytecode of burner v5
	BurnerCodeHashV5 = "0x9f217a79e864028081339cfcead3c3d1fe92e237fcbe9468d6bb4d1da7aa6352"
)

const (
	// DefaultRateLimitWindow is the default rate limit window, also used by the gateway
	DefaultRateLimitWindow = 6 * time.Hour
)

func validateBurnerCode(burnerCode []byte) error {
	burnerCodeHash := crypto.Keccak256Hash(burnerCode).Hex()
	switch burnerCodeHash {
	case BurnerCodeHashV1,
		BurnerCodeHashV2,
		BurnerCodeHashV3,
		BurnerCodeHashV4,
		BurnerCodeHashV5:
		break
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

	axelarGatewayFuncExecute = "execute"
)

// IsEVMChain returns true if a chain is an EVM chain
func IsEVMChain(chain nexus.Chain) bool {
	return chain.Module == ModuleName
}

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
func (t ERC20Token) GetAsset() string {
	return t.metadata.Asset
}

// GetTxID returns the tx ID set with StartConfirmation
func (t ERC20Token) GetTxID() Hash {
	return t.metadata.TxHash
}

// GetDetails returns the details of the token
func (t ERC20Token) GetDetails() TokenDetails {
	return t.metadata.Details
}

// Is returns true if the given status matches the token's status
func (t ERC20Token) Is(status Status) bool {
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

// GetBurnerCode returns the version of the burner the token is deployed with
func (t ERC20Token) GetBurnerCode() []byte {
	return t.metadata.BurnerCode
}

// GetBurnerCodeHash returns the version of the burner the token is deployed with if it exists
func (t ERC20Token) GetBurnerCodeHash() (Hash, bool) {
	if t.metadata.BurnerCode == nil {
		return Hash{}, false
	}

	return Hash(crypto.Keccak256Hash(t.metadata.BurnerCode)), true
}

// CreateDeployCommand returns a token deployment command for the token
func (t *ERC20Token) CreateDeployCommand(keyID multisig.KeyID, dailyMintLimit sdk.Uint) (Command, error) {
	switch {
	case t.Is(NonExistent):
		return Command{}, fmt.Errorf("token %s non-existent", t.GetAsset())
	case t.Is(Confirmed):
		return Command{}, fmt.Errorf("token %s already confirmed", t.GetAsset())
	}
	if err := keyID.ValidateBasic(); err != nil {
		return Command{}, err
	}

	if t.IsExternal() {
		return NewDeployTokenCommand(
			t.metadata.ChainID,
			keyID,
			t.GetAsset(),
			t.metadata.Details,
			t.GetAddress(),
			dailyMintLimit,
		), nil
	}

	return NewDeployTokenCommand(
		t.metadata.ChainID,
		keyID,
		t.GetAsset(),
		t.metadata.Details,
		ZeroAddress,
		dailyMintLimit,
	), nil
}

// CreateMintCommand returns a mint deployment command for the token
func (t *ERC20Token) CreateMintCommand(keyID multisig.KeyID, transfer nexus.CrossChainTransfer) (Command, error) {
	if !t.Is(Confirmed) {
		return Command{}, fmt.Errorf("token %s not confirmed (current status: %s)",
			t.metadata.Asset, t.metadata.Status.String())
	}
	if err := keyID.ValidateBasic(); err != nil {
		return Command{}, err
	}

	return NewMintTokenCommand(
		keyID,
		transfer.ID,
		t.metadata.Details.Symbol,
		common.HexToAddress(transfer.Recipient.Address),
		transfer.Asset.Amount.BigInt(),
	), nil
}

// GetAddress returns the token's address
func (t ERC20Token) GetAddress() Address {
	return t.metadata.TokenAddress

}

// RecordDeployment signals that the token confirmation is underway for the given tx ID
func (t *ERC20Token) RecordDeployment(txID Hash) error {
	switch {
	case t.Is(NonExistent):
		return fmt.Errorf("token %s non-existent", t.metadata.Asset)
	case t.Is(Confirmed):
		return fmt.Errorf("token %s already confirmed", t.metadata.Asset)
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

// ZeroHash represents an empty 32-bytes hash
var ZeroHash = common.Hash{}

// IsZero returns true if the hash is empty; otherwise false
func (h Hash) IsZero() bool {
	return bytes.Equal(h.Bytes(), ZeroHash.Bytes())
}

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

// NewSignature is the constructor of Signature
func NewSignature(bz []byte) (sig Signature, err error) {
	if len(bz) != crypto.SignatureLength {
		return Signature{}, fmt.Errorf("invalid signature length")
	}

	copy(sig[:], bz)

	return sig, nil
}

// Hex returns the hex-encoding of the given Signature
func (s Signature) Hex() string {
	return hex.EncodeToString(s[:])
}

// ToHomesteadSig converts signature to openzeppelin compatible
func (s Signature) ToHomesteadSig() []byte {
	/* TODO: We have to make v 27 or 28 due to openzeppelin's implementation at https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/cryptography/ECDSA.sol
	requiring that. Consider copying and modifying it to require v to be just 0 or 1
	instead.
	*/
	bz := s[:]
	if bz[crypto.SignatureLength-1] == 0 || bz[crypto.SignatureLength-1] == 1 {
		bz[crypto.SignatureLength-1] += 27
	}

	return bz
}

// ToSignature transforms an Axelar generated signature into a recoverable signature
func ToSignature(sig ec.Signature, hash common.Hash, pk ecdsa.PublicKey) (Signature, error) {
	s := Signature{}
	encSig := sig.Serialize()

	// read R length
	encSig = encSig[3:]
	rLen := int(encSig[0])
	encSig = encSig[1:]

	// extract R
	encR := encSig[:rLen]
	if encR[0] == 0 {
		encR = encR[1:]
	}
	copy(s[:32], common.LeftPadBytes(encR, 32))
	encSig = encSig[rLen:]

	// read S length
	encSig = encSig[1:]
	sLen := int(encSig[0])
	encSig = encSig[1:]

	// extract S
	encS := encSig[:sLen]
	if encS[0] == 0 {
		encS = encS[1:]
	}
	copy(s[32:], common.LeftPadBytes(encS, 32))

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

// KeysToAddresses converts a slice of ECDSA public keys to evm addresses
func KeysToAddresses(keys ...ecdsa.PublicKey) []common.Address {
	addresses := make([]common.Address, len(keys))

	for i, key := range keys {
		addresses[i] = crypto.PubkeyToAddress(key)
	}

	return addresses
}

// CreateExecuteDataMultisig wraps the specific command data and includes the command signatures.
// Returns the data that goes into the data field of an EVM transaction
func CreateExecuteDataMultisig(data []byte, addresses []common.Address, weights []sdk.Uint, threshold sdk.Uint, signatures [][]byte) ([]byte, error) {
	abiEncoder, err := abi.JSON(strings.NewReader(axelarGatewayABI))
	if err != nil {
		return nil, err
	}

	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}

	proof, err := getWeightedSignaturesProof(addresses, weights, threshold, signatures)
	if err != nil {
		return nil, err
	}

	executeData, err := abi.Arguments{{Type: bytesType}, {Type: bytesType}}.Pack(data, proof)
	if err != nil {
		return nil, err
	}

	return abiEncoder.Pack(axelarGatewayFuncExecute, executeData)
}

// GetSignHash returns the hash that needs to be signed so AxelarGateway accepts the given command
func GetSignHash(commandData []byte) common.Hash {
	hash := crypto.Keccak256(commandData)

	// TODO: is this the same across any EVM chain?
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(hash), hash)

	return crypto.Keccak256Hash([]byte(msg))
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
func (b CommandBatch) GetKeyID() multisig.KeyID {
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

// GetSignature returns the batch's signature
func (b CommandBatch) GetSignature() utils.ValidatedProtoMarshaler {
	if b.metadata.Signature == nil {
		return nil
	}

	return b.metadata.Signature.GetCachedValue().(utils.ValidatedProtoMarshaler)
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

// SetSigned sets the signature and signed status for the batch
func (b *CommandBatch) SetSigned(signature utils.ValidatedProtoMarshaler) error {
	if b.metadata.Status != BatchSigning {
		return fmt.Errorf("command batch %s is not being signed", hex.EncodeToString(b.GetID()))
	}

	b.metadata.Status = BatchSigned
	sig := funcs.Must(codectypes.NewAnyWithValue(signature))
	b.metadata.Signature = sig

	b.setter(b.metadata)

	return nil
}

// NewCommandBatchMetadata assembles a CommandBatchMetadata struct from the provided arguments
func NewCommandBatchMetadata(blockHeight int64, chainID sdk.Int, keyID multisig.KeyID, cmds []Command) (CommandBatchMetadata, error) {
	var commandIDs []CommandID
	var commands []CommandType
	var commandParams [][]byte

	for _, cmd := range cmds {
		commandIDs = append(commandIDs, cmd.ID)
		commands = append(commands, cmd.Type)
		commandParams = append(commandParams, cmd.Params)
	}

	data, err := packArguments(chainID, commandIDs, commands, commandParams)
	if err != nil {
		return CommandBatchMetadata{}, err
	}

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))

	return CommandBatchMetadata{
		ID:         crypto.Keccak256(bz, data),
		CommandIDs: commandIDs,
		Data:       data,
		SigHash:    Hash(GetSignHash(data)),
		Status:     BatchSigning,
		KeyID:      keyID,
	}, nil
}

// ValidateBasic returns an error if the CommandBatchMetadata is not valid
func (m CommandBatchMetadata) ValidateBasic() error {
	switch m.Status {
	case BatchNonExistent:
		return errors.New("batch does not exist")
	case BatchSigning, BatchAborted:
		if m.Signature != nil {
			return errors.New("unsigned batch must not have a signature")
		}
	case BatchSigned:
		if m.Signature == nil {
			return errors.New("signed batch must have a valid signature")
		}

		if err := m.Signature.GetCachedValue().(utils.ValidatedProtoMarshaler).ValidateBasic(); err != nil {
			return err
		}
	}

	if len(m.ID) != 32 {
		return errors.New("batch ID must be of length 32")
	}

	if len(m.CommandIDs) == 0 {
		return errors.New("command IDs must not be empty")
	}

	if len(m.Data) == 0 {
		return errors.New("batch data must not be empty")
	}

	if m.SigHash.IsZero() {
		return errors.New("batch data hash must not be empty")
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
	}

	if len(m.PrevBatchedCommandsID) != 0 && len(m.PrevBatchedCommandsID) != 32 {
		return errors.New("previous batch ID must either be nil or of length 32")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m CommandBatchMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler

	return unpacker.UnpackAny(m.Signature, &data)
}

const commandIDSize = 32

// CommandID represents the unique command identifier
type CommandID [commandIDSize]byte

// NewCommandID is the constructor for CommandID
func NewCommandID(data []byte, chainID sdk.Int) CommandID {
	var commandID CommandID
	copy(commandID[:], crypto.Keccak256(append(data, chainID.BigInt().Bytes()...))[:commandIDSize])

	return commandID
}

// CommandIDFromTransferID converts a TransferID into a CommandID
func CommandIDFromTransferID(id nexus.TransferID) CommandID {
	var commandID CommandID
	idBz := id.Bytes()

	copy(commandID[:], common.LeftPadBytes(idBz[:], commandIDSize))

	return commandID
}

// HexToCommandID decodes a hex representation of a CommandID
func HexToCommandID(id string) (CommandID, error) {
	bz, err := utils.HexDecode(id)
	if err != nil {
		return CommandID{}, err
	}

	var commandID CommandID
	copy(commandID[:], bz)

	return commandID, commandID.ValidateBasic()
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

	return c.ValidateBasic()
}

// ValidateBasic returns an error if the given command ID is invalid
func (c CommandID) ValidateBasic() error {
	return nil
}

// NewAsset returns a new Asset instance
func NewAsset(chain, name string) Asset {
	return Asset{
		Chain: nexus.ChainName(utils.NormalizeString(chain)),
		Name:  utils.NormalizeString(name),
	}
}

// Validate ensures that all fields are filled with sensible values
func (m Asset) Validate() error {
	if err := m.Chain.Validate(); err != nil {
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

func packArguments(chainID sdk.Int, commandIDs []CommandID, commands []CommandType, commandParams [][]byte) ([]byte, error) {
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
		chainID.BigInt(),
		commandIDs,
		slices.Map(commands, CommandType.String),
		commandParams,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ValidateBasic does stateless validation of the object
func (m *BurnerInfo) ValidateBasic() error {
	if err := m.DestinationChain.Validate(); err != nil {
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

	switch m.IsExternal {
	case true:
		if m.BurnerCode != nil {
			return fmt.Errorf("burner code for external tokens must be nil")
		}
	case false:
		if err := validateBurnerCode(m.BurnerCode); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic does stateless validation of the object
func (m *ERC20Deposit) ValidateBasic() error {
	if err := sdk.ValidateDenom(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if err := m.DestinationChain.Validate(); err != nil {
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

// GetID returns an unique ID for the event
func (m Event) GetID() EventID {
	return NewEventID(m.TxID, m.Index)
}

// ValidateBasic returns an error if the event is invalid
func (m Event) ValidateBasic() error {
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid source chain")
	}

	if m.TxID.IsZero() {
		return fmt.Errorf("invalid tx id")
	}

	switch event := m.GetEvent().(type) {
	case *Event_ContractCall:
		if event.ContractCall == nil {
			return fmt.Errorf("missing event ContractCall")
		}

		if err := event.ContractCall.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event ContractCall")
		}
	case *Event_ContractCallWithToken:
		if event.ContractCallWithToken == nil {
			return fmt.Errorf("missing event ContractCallWithToken")
		}

		if err := event.ContractCallWithToken.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event ContractCallWithToken")
		}
	case *Event_TokenSent:
		if event.TokenSent == nil {
			return fmt.Errorf("missing event TokenSent")
		}

		if err := event.TokenSent.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event TokenSent")
		}
	case *Event_Transfer:
		if event.Transfer == nil {
			return fmt.Errorf("missing event Transfer")
		}
		if err := event.Transfer.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event Transfer")
		}
	case *Event_TokenDeployed:
		if event.TokenDeployed == nil {
			return fmt.Errorf("missing event TokenDeployed")
		}
		if err := event.TokenDeployed.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event TokenDeployed")
		}
	case *Event_MultisigOperatorshipTransferred:
		if event.MultisigOperatorshipTransferred == nil {
			return fmt.Errorf("missing event MultisigOperatorshipTransferred")
		}
		if err := event.MultisigOperatorshipTransferred.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event MultisigOperatorshipTransferred")
		}
	default:
		return fmt.Errorf("unknown type of event")
	}

	return nil
}

// GetEventType returns the type for the event
func (m Event) GetEventType() string {
	return getType(m.GetEvent())
}

const maxReceiverLength = 128

// ValidateBasic returns an error if the event token sent is invalid
func (m EventTokenSent) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if err := utils.ValidateString(m.DestinationAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid destination address")
	}

	if len(m.DestinationAddress) > maxReceiverLength {
		return fmt.Errorf("receiver length %d is greater than %d", len(m.DestinationAddress), maxReceiverLength)
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event contract call is invalid
func (m EventContractCall) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if err := utils.ValidateString(m.ContractAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid destination address")
	}

	if len(m.ContractAddress) > maxReceiverLength {
		return fmt.Errorf("receiver length %d is greater than %d", len(m.ContractAddress), maxReceiverLength)
	}

	if m.PayloadHash.IsZero() {
		return fmt.Errorf("invalid payload hash")
	}

	return nil
}

// ValidateBasic returns an error if the event contract call with token is invalid
func (m EventContractCallWithToken) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	if err := utils.ValidateString(m.ContractAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid destination address")
	}

	if len(m.ContractAddress) > maxReceiverLength {
		return fmt.Errorf("receiver length %d is greater than %d", len(m.ContractAddress), maxReceiverLength)
	}

	if m.PayloadHash.IsZero() {
		return fmt.Errorf("invalid payload hash")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event transfer is invalid
func (m EventTransfer) ValidateBasic() error {
	if m.To.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if m.Amount.IsZero() {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// ValidateBasic returns an error if the event token deployed is invalid
func (m EventTokenDeployed) ValidateBasic() error {
	if m.TokenAddress.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := utils.ValidateString(m.Symbol); err != nil {
		return sdkerrors.Wrap(err, "invalid symbol")
	}

	return nil
}

// ValidateBasic returns an error if the event multisig operatorship transferred is invalid
func (m EventMultisigOperatorshipTransferred) ValidateBasic() error {
	if slices.Any(m.NewOperators, Address.IsZeroAddress) {
		return fmt.Errorf("invalid new operators")
	}

	if len(m.NewOperators) != len(m.NewWeights) {
		return fmt.Errorf("length of new operators does not match new weights")
	}

	totalWeight := sdk.ZeroUint()
	slices.ForEach(m.NewWeights, func(w sdk.Uint) { totalWeight = totalWeight.Add(w) })

	if m.NewThreshold.IsZero() || m.NewThreshold.GT(totalWeight) {
		return fmt.Errorf("invalid new threshold")
	}

	return nil
}

// StrictDecode performs strict decode on evm encoded data, e.g. no byte can be left after the decoding
func StrictDecode(arguments abi.Arguments, bz []byte) ([]interface{}, error) {
	params, err := arguments.Unpack(bz)
	if err != nil {
		return nil, err
	}

	if actual, err := arguments.Pack(params...); err != nil || !bytes.Equal(actual, bz) {
		return nil, fmt.Errorf("wrong data")
	}

	return params, nil
}

// NewVoteEvents is the constructor for vote events
func NewVoteEvents(chain nexus.ChainName, events ...Event) *VoteEvents {
	return &VoteEvents{
		Chain:  chain,
		Events: events,
	}
}

// ValidateBasic does stateless validation of the object
func (m VoteEvents) ValidateBasic() error {
	if err := m.Chain.Validate(); err != nil {
		return err
	}

	for _, event := range m.Events {
		if err := event.ValidateBasic(); err != nil {
			return err
		}

		if event.Chain != m.Chain {
			return fmt.Errorf("events are not from the same source chain")
		}
	}

	return nil
}

// GetMultisigAddressesAndWeights coverts a multisig key to sorted addresses, weights and threshold
func GetMultisigAddressesAndWeights(key multisig.Key) ([]common.Address, []sdk.Uint, sdk.Uint) {
	addressWeights, threshold := ParseMultisigKey(key)
	addresses := slices.Map(maps.Keys(addressWeights), common.HexToAddress)
	sort.SliceStable(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i].Bytes(), addresses[j].Bytes()) < 0
	})
	weights := slices.Map(addresses, func(address common.Address) sdk.Uint {
		return addressWeights[address.Hex()]
	})

	return addresses, weights, threshold
}

func getType(val interface{}) string {
	t := reflect.TypeOf(val)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// EventID ensures a correctly formatted event ID
type EventID string

// NewEventID returns a new event ID
func NewEventID(txID Hash, index uint64) EventID {
	return EventID(fmt.Sprintf("%s-%d", txID.Hex(), index))
}

// Validate returns an error, if the event ID is not in format of txID-index
func (id EventID) Validate() error {
	if err := utils.ValidateString(string(id)); err != nil {
		return err
	}

	arr := strings.Split(string(id), "-")
	if len(arr) != 2 {
		return fmt.Errorf("event ID should be in foramt of txID-index")
	}

	bz, err := hexutil.Decode(arr[0])
	if err != nil {
		return sdkerrors.Wrap(err, "invalid tx hash hex encoding")
	}

	if len(bz) != common.HashLength {
		return fmt.Errorf("invalid tx hash length")
	}

	_, err = strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid index")
	}

	return nil
}

// ParseMultisigKey parses the given multisig key and returns the weight for
// each particpant evm address and the threshold
func ParseMultisigKey(key multisig.Key) (map[string]sdk.Uint, sdk.Uint) {
	participants := key.GetParticipants()
	addressWeights := make(map[string]sdk.Uint, len(participants))

	for _, p := range participants {
		pubKey := funcs.MustOk(key.GetPubKey(p))
		weight := key.GetWeight(p)
		address := crypto.PubkeyToAddress(pubKey.ToECDSAPubKey())

		addressWeights[address.Hex()] = weight
	}

	return addressWeights, key.GetMinPassingWeight()
}

// NewSigMetadata is the constructor for sig metadata
func NewSigMetadata(sigType SigType, chain nexus.ChainName, commandBatchID []byte) *SigMetadata {
	return &SigMetadata{
		Type:           sigType,
		Chain:          chain,
		CommandBatchID: commandBatchID,
	}
}

func getWeightedSignaturesProof(addresses []common.Address, weights []sdk.Uint, threshold sdk.Uint, signatures [][]byte) ([]byte, error) {
	addressesType, err := abi.NewType("address[]", "address[]", nil)
	if err != nil {
		return nil, err
	}

	weightsType, err := abi.NewType("uint256[]", "uint256[]", nil)
	if err != nil {
		return nil, err
	}

	thresholdType, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	signaturesType, err := abi.NewType("bytes[]", "bytes[]", nil)
	if err != nil {
		return nil, err
	}

	proof, err := abi.Arguments{
		{Type: addressesType},
		{Type: weightsType},
		{Type: thresholdType},
		{Type: signaturesType}}.Pack(
		addresses,
		slices.Map(weights, sdk.Uint.BigInt),
		threshold.BigInt(),
		signatures,
	)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// Operator binds the signer's address, signature and weight
type Operator struct {
	Address   common.Address
	Signature []byte
	Weight    sdk.Uint
}

// ValidateBasic returns an error if the Gateway address is invalid
func (m Gateway) ValidateBasic() error {
	if m.Address.IsZeroAddress() {
		return errors.New("address must not be empty")
	}

	return nil
}

// ValidateBasic returns an error if the given command is invalid
func (m Command) ValidateBasic() error {
	if err := m.ID.ValidateBasic(); err != nil {
		return err
	}

	if err := m.Type.ValidateBasic(); err != nil {
		return err
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
	}

	if m.MaxGasCost == 0 {
		return errors.New("max gas cost must be >0")
	}

	if _, err := m.DecodeParams(); err != nil {
		return err
	}
	return nil
}

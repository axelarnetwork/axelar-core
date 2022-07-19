package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
	AxelarGatewayCommandMintToken                   = "mintToken"
	mintTokenMaxGasCost                             = 150000
	AxelarGatewayCommandDeployToken                 = "deployToken"
	deployTokenMaxGasCost                           = 1400000
	AxelarGatewayCommandBurnToken                   = "burnToken"
	burnExternalTokenMaxGasCost                     = 400000
	burnInternalTokenMaxGasCost                     = 120000
	AxelarGatewayCommandTransferOperatorship        = "transferOperatorship"
	transferOperatorshipMaxGasCost                  = 120000
	AxelarGatewayCommandApproveContractCallWithMint = "approveContractCallWithMint"
	approveContractCallWithMintMaxGasCost           = 120000
	AxelarGatewayCommandApproveContractCall         = "approveContractCall"
	approveContractCallMaxGasCost                   = 120000
	axelarGatewayFuncExecute                        = "execute"
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

// GetBurnerCodeHash returns the version of the burner the token is deployed with
func (t ERC20Token) GetBurnerCodeHash() Hash {
	return Hash(crypto.Keccak256Hash(t.metadata.BurnerCode))
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
		return CreateDeployTokenCommand(
			t.metadata.ChainID,
			keyID,
			t.GetAsset(),
			t.metadata.Details,
			t.GetAddress(),
			dailyMintLimit,
		)
	}

	return CreateDeployTokenCommand(
		t.metadata.ChainID,
		keyID,
		t.GetAsset(),
		t.metadata.Details,
		ZeroAddress,
		dailyMintLimit,
	)
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

	return CreateMintTokenCommand(keyID, transferIDtoCommandID(transfer.ID), t.metadata.Details.Symbol, common.HexToAddress(transfer.Recipient.Address), transfer.Asset.Amount.BigInt())
}

// transferIDtoCommandID converts a transferID to a commandID
func transferIDtoCommandID(transferID nexus.TransferID) CommandID {
	var commandID CommandID
	copy(commandID[:], common.LeftPadBytes(transferID.Bytes(), 32)[:32])

	return commandID
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

// CreateApproveContractCallCommand creates a command to approve contract call
func CreateApproveContractCallCommand(
	chainID sdk.Int,
	keyID multisig.KeyID,
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCall,
) (Command, error) {
	params, err := createApproveContractCallParams(sourceChain, sourceTxID, sourceEventIndex, event)
	if err != nil {
		return Command{}, err
	}

	sourceEventIndexBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sourceEventIndexBz, sourceEventIndex)

	return Command{
		ID:         NewCommandID(append(sourceTxID.Bytes(), sourceEventIndexBz...), chainID),
		Command:    AxelarGatewayCommandApproveContractCall,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallMaxGasCost),
	}, nil
}

// CreateApproveContractCallWithMintCommand creates a command to approve contract call with token being minted
func CreateApproveContractCallWithMintCommand(
	chainID sdk.Int,
	keyID multisig.KeyID,
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCallWithToken,
	amount sdk.Uint,
	symbol string,
) (Command, error) {
	params, err := createApproveContractCallWithMintParams(sourceChain, sourceTxID, sourceEventIndex, event, amount, symbol)
	if err != nil {
		return Command{}, err
	}

	sourceEventIndexBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sourceEventIndexBz, sourceEventIndex)

	return Command{
		ID:         NewCommandID(append(sourceTxID.Bytes(), sourceEventIndexBz...), chainID),
		Command:    AxelarGatewayCommandApproveContractCallWithMint,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallWithMintMaxGasCost),
	}, nil
}

// decodeApproveContractCallWithMintParams unpacks the parameters of a approve contract call with mint command
func decodeApproveContractCallWithMintParams(bz []byte) (string, string, common.Address, common.Hash, string, *big.Int, common.Hash, *big.Int, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, "", nil, common.Hash{}, nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, "", nil, common.Hash{}, nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, "", nil, common.Hash{}, nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, "", nil, common.Hash{}, nil, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: stringType},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: uint256Type},
	}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, "", nil, common.Hash{}, nil, err
	}

	payloadHash := params[3].([common.HashLength]byte)
	sourceTxID := params[6].([common.HashLength]byte)

	return params[0].(string),
		params[1].(string),
		params[2].(common.Address),
		common.BytesToHash(payloadHash[:]),
		params[4].(string),
		params[5].(*big.Int),
		common.BytesToHash(sourceTxID[:]),
		params[7].(*big.Int),
		nil
}

// decodeApproveContractCallParams unpacks the parameters of a approve contract call command
func decodeApproveContractCallParams(bz []byte) (string, string, common.Address, common.Hash, common.Hash, *big.Int, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, common.Hash{}, nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, common.Hash{}, nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, common.Hash{}, nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, common.Hash{}, nil, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
	}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return "", "", common.Address{}, common.Hash{}, common.Hash{}, nil, err
	}

	payloadHash := params[3].([common.HashLength]byte)
	sourceTxID := params[4].([common.HashLength]byte)

	return params[0].(string),
		params[1].(string),
		params[2].(common.Address),
		common.BytesToHash(payloadHash[:]),
		common.BytesToHash(sourceTxID[:]),
		params[5].(*big.Int),
		nil
}

func createApproveContractCallParams(
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCall) ([]byte, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
	}

	result, err := arguments.Pack(
		sourceChain,
		event.Sender.Hex(),
		common.HexToAddress(event.ContractAddress),
		common.Hash(event.PayloadHash),
		common.Hash(sourceTxID),
		new(big.Int).SetUint64(sourceEventIndex),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func createApproveContractCallWithMintParams(
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCallWithToken,
	amount sdk.Uint,
	symbol string) ([]byte, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}

	arguments := abi.Arguments{
		{Type: stringType},
		{Type: stringType},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: stringType},
		{Type: uint256Type},
		{Type: bytes32Type},
		{Type: uint256Type},
	}
	result, err := arguments.Pack(
		sourceChain,
		event.Sender.Hex(),
		common.HexToAddress(event.ContractAddress),
		common.Hash(event.PayloadHash),
		symbol,
		amount.BigInt(),
		common.Hash(sourceTxID),
		new(big.Int).SetUint64(sourceEventIndex),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateBurnTokenCommand creates a command to burn tokens with the given burner's information
func CreateBurnTokenCommand(chainID sdk.Int, keyID multisig.KeyID, height int64, burnerInfo BurnerInfo, isTokenExternal bool) (Command, error) {
	params, err := createBurnTokenParams(burnerInfo.Symbol, common.Hash(burnerInfo.Salt))
	if err != nil {
		return Command{}, err
	}

	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))

	burnTokenMaxGasCost := burnInternalTokenMaxGasCost
	if isTokenExternal {
		burnTokenMaxGasCost = burnExternalTokenMaxGasCost
	}

	return Command{
		ID:         NewCommandID(append(burnerInfo.Salt.Bytes(), heightBytes...), chainID),
		Command:    AxelarGatewayCommandBurnToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: uint32(burnTokenMaxGasCost),
	}, nil
}

// CreateDeployTokenCommand creates a command to deploy a token
func CreateDeployTokenCommand(chainID sdk.Int, keyID multisig.KeyID, asset string, tokenDetails TokenDetails, address Address, dailyMintLimit sdk.Uint) (Command, error) {
	params, err := createDeployTokenParams(tokenDetails.TokenName, tokenDetails.Symbol, tokenDetails.Decimals, tokenDetails.Capacity, address, dailyMintLimit)
	if err != nil {
		return Command{}, err
	}

	return Command{
		ID:         NewCommandID([]byte(fmt.Sprintf("%s_%s", asset, tokenDetails.Symbol)), chainID),
		Command:    AxelarGatewayCommandDeployToken,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: deployTokenMaxGasCost,
	}, nil
}

// CreateMintTokenCommand creates a command to mint token to the given address
func CreateMintTokenCommand(keyID multisig.KeyID, id CommandID, symbol string, address common.Address, amount *big.Int) (Command, error) {
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

// CreateMultisigTransferCommand creates a command to transfer ownership/operator of the multisig contract
func CreateMultisigTransferCommand(chainID sdk.Int, keyID multisig.KeyID, nextKey multisig.Key) Command {
	addresses, weights, threshold := GetMultisigAddressesAndWeights(nextKey)
	params := createTransferMultisigParams(addresses, slices.Map(weights, sdk.Uint.BigInt), threshold.BigInt())

	var concat []byte
	for _, addr := range addresses {
		concat = append(concat, addr.Bytes()...)
	}

	return Command{
		ID:         NewCommandID(concat, chainID),
		Command:    AxelarGatewayCommandTransferOperatorship,
		Params:     params,
		KeyID:      keyID,
		MaxGasCost: transferOperatorshipMaxGasCost,
	}
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
func (b CommandBatch) GetSignature() codec.ProtoMarshaler {
	if b.metadata.Signature == nil {
		return nil
	}

	return b.metadata.Signature.GetCachedValue().(codec.ProtoMarshaler)
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
func (b *CommandBatch) SetSigned(signature codec.ProtoMarshaler) error {
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

func packArguments(chainID sdk.Int, commandIDs []CommandID, commands []string, commandParams [][]byte) ([]byte, error) {
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

// decodeMintTokenParams unpacks the parameters of a mint token command
func decodeMintTokenParams(bz []byte) (string, common.Address, *big.Int, error) {
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
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return "", common.Address{}, nil, err
	}

	return params[0].(string), params[1].(common.Address), params[2].(*big.Int), nil
}

func createDeployTokenParams(tokenName string, symbol string, decimals uint8, capacity sdk.Int, address Address, dailyMintLimit sdk.Uint) ([]byte, error) {
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

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}, {Type: addressType}, {Type: uint256Type}}
	result, err := arguments.Pack(
		tokenName,
		symbol,
		decimals,
		capacity.BigInt(),
		address,
		dailyMintLimit.BigInt(),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// decodeDeployTokenParams unpacks the parameters of a deploy token command
func decodeDeployTokenParams(bz []byte) (string, string, uint8, *big.Int, common.Address, sdk.Uint, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", "", 0, nil, common.Address{}, sdk.OneUint(), err
	}

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return "", "", 0, nil, common.Address{}, sdk.OneUint(), err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return "", "", 0, nil, common.Address{}, sdk.OneUint(), err
	}

	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return "", "", 0, nil, common.Address{}, sdk.OneUint(), err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}, {Type: addressType}, {Type: uint256Type}}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return "", "", 0, nil, common.Address{}, sdk.OneUint(), err
	}

	dailyMintLimit := sdk.NewUintFromBigInt(params[5].(*big.Int))

	return params[0].(string), params[1].(string), params[2].(uint8), params[3].(*big.Int), params[4].(common.Address), dailyMintLimit, nil
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

// decodeBurnTokenParams unpacks the parameters of a burn token command
func decodeBurnTokenParams(bz []byte) (string, common.Hash, error) {
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return "", common.Hash{}, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return "", common.Hash{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: bytes32Type}}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return "", common.Hash{}, err
	}

	return params[0].(string), params[1].([common.HashLength]byte), nil
}

// decodeTransferSinglesigParams unpacks the parameters of a single sig transfer command
func decodeTransferSinglesigParams(bz []byte) (common.Address, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return common.Address{}, err
	}

	arguments := abi.Arguments{{Type: addressType}}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return common.Address{}, err
	}

	return params[0].(common.Address), nil
}

func createTransferMultisigParams(addresses []common.Address, weights []*big.Int, threshold *big.Int) []byte {
	addressesType := funcs.Must(abi.NewType("address[]", "address[]", nil))
	uint256ArrayType := funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))
	uint256Type := funcs.Must(abi.NewType("uint256", "uint256", nil))

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256ArrayType}, {Type: uint256Type}}

	return funcs.Must(arguments.Pack(addresses, weights, threshold))
}

// decodeTransferMultisigParams unpacks the parameters of a multi sig transfer command
func decodeTransferMultisigParams(bz []byte) ([]common.Address, []*big.Int, *big.Int, error) {
	addressesType := funcs.Must(abi.NewType("address[]", "address[]", nil))
	uint256ArrayType := funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))
	uint256Type := funcs.Must(abi.NewType("uint256", "uint256", nil))

	arguments := abi.Arguments{{Type: addressesType}, {Type: uint256ArrayType}, {Type: uint256Type}}
	params, err := StrictDecode(arguments, bz)
	if err != nil {
		return nil, nil, nil, err
	}

	return params[0].([]common.Address), params[1].([]*big.Int), params[2].(*big.Int), nil
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
	return EventID(fmt.Sprintf("%s-%d", m.TxId.Hex(), m.Index))
}

// ValidateBasic returns an error if the event is invalid
func (m Event) ValidateBasic() error {
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid source chain")
	}

	if m.TxId.IsZero() {
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
	case *Event_MultisigOwnershipTransferred:
		if event.MultisigOwnershipTransferred == nil {
			return fmt.Errorf("missing event MultisigOwnershipTransferred")
		}
		if err := event.MultisigOwnershipTransferred.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "invalid event MultisigOwnershipTransferred")
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

// ValidateBasic returns an error if the event token sent is invalid
func (m EventTokenSent) ValidateBasic() error {
	if m.Sender.IsZeroAddress() {
		return fmt.Errorf("invalid sender")
	}

	if err := m.DestinationChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid destination chain")
	}

	err := utils.ValidateString(m.DestinationAddress)
	if err != nil || common.IsHexAddress(m.DestinationAddress) {
		return sdkerrors.Wrap(err, "invalid destination address")
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

	err := utils.ValidateString(m.ContractAddress)
	if err != nil || common.IsHexAddress(m.ContractAddress) {
		return sdkerrors.Wrap(err, "invalid contract address")
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

	err := utils.ValidateString(m.ContractAddress)
	if err != nil || common.IsHexAddress(m.ContractAddress) {
		return sdkerrors.Wrap(err, "invalid contract address")
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

// ValidateBasic returns an error if the event multisig ownership transferred is invalid
func (m EventMultisigOwnershipTransferred) ValidateBasic() error {
	NonzeroAddress := func(addr Address) bool { return !addr.IsZeroAddress() }

	if !slices.All(m.PreOwners, NonzeroAddress) {
		return fmt.Errorf("invalid pre owners")
	}
	if m.PrevThreshold.IsZero() {
		return fmt.Errorf("invalid pre threshold")
	}
	if !slices.All(m.NewOwners, NonzeroAddress) {
		return fmt.Errorf("invalid new owners")
	}
	if m.NewThreshold.IsZero() {
		return fmt.Errorf("invalid new threshold")
	}

	return nil
}

// ValidateBasic returns an error if the event multisig ownership transferred is invalid
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

// DecodeParams returns the decoded parameters in the given command
func (c Command) DecodeParams() (map[string]string, error) {
	params := make(map[string]string)

	switch c.Command {
	case AxelarGatewayCommandApproveContractCallWithMint:
		sourceChain, sourceAddress, contractAddress, payloadHash, symbol, amount, sourceTxID, sourceEventIndex, err := decodeApproveContractCallWithMintParams(c.Params)
		if err != nil {
			return nil, err
		}

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["symbol"] = symbol
		params["amount"] = amount.String()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case AxelarGatewayCommandApproveContractCall:
		sourceChain, sourceAddress, contractAddress, payloadHash, sourceTxID, sourceEventIndex, err := decodeApproveContractCallParams(c.Params)
		if err != nil {
			return nil, err
		}

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case AxelarGatewayCommandDeployToken:
		name, symbol, decs, cap, tokenAddress, dailyMintLimit, err := decodeDeployTokenParams(c.Params)
		if err != nil {
			return nil, err
		}

		params["name"] = name
		params["symbol"] = symbol
		params["decimals"] = strconv.FormatUint(uint64(decs), 10)
		params["cap"] = cap.String()
		params["tokenAddress"] = tokenAddress.Hex()
		params["dailyMintLimit"] = dailyMintLimit.String()
	case AxelarGatewayCommandMintToken:
		symbol, addr, amount, err := decodeMintTokenParams(c.Params)
		if err != nil {
			return nil, err
		}

		params["symbol"] = symbol
		params["account"] = addr.Hex()
		params["amount"] = amount.String()
	case AxelarGatewayCommandBurnToken:
		symbol, salt, err := decodeBurnTokenParams(c.Params)
		if err != nil {
			return nil, err
		}

		params["symbol"] = symbol
		params["salt"] = salt.Hex()
	case AxelarGatewayCommandTransferOperatorship:
		address, decodeSinglesigErr := decodeTransferSinglesigParams(c.Params)
		addresses, weights, threshold, decodeMultisigErr := decodeTransferMultisigParams(c.Params)

		switch {
		case decodeSinglesigErr == nil:
			params["newOperator"] = address.Hex()
		case decodeMultisigErr == nil:
			params["newOperators"] = strings.Join(slices.Map(addresses, common.Address.Hex), ";")
			params["newWeights"] = strings.Join(slices.Map(weights, func(w *big.Int) string { return w.String() }), ";")
			params["newThreshold"] = threshold.String()
		default:
			return nil, fmt.Errorf("unsupported type of transfer key")
		}
	default:
		return nil, fmt.Errorf("unknown command type '%s'", c.Command)
	}

	return params, nil
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
func NewVoteEvents(chain nexus.ChainName, events []Event) *VoteEvents {
	return &VoteEvents{
		Chain:  chain,
		Events: events,
	}
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
	if err != nil || len(bz) != common.HashLength {
		return sdkerrors.Wrap(err, "invalid tx hash")
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

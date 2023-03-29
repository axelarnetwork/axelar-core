package types

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/stoewer/go-strcase"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

const (
	mintTokenMaxGasCost                   = 100000
	deployTokenMaxGasCost                 = 1400000
	burnExternalTokenMaxGasCost           = 400000
	burnInternalTokenMaxGasCost           = 120000
	transferOperatorshipMaxGasCost        = 120000
	approveContractCallWithMintMaxGasCost = 100000
	approveContractCallMaxGasCost         = 100000
)

func (c CommandType) String() string {
	return strcase.LowerCamelCase(strings.TrimPrefix(proto.EnumName(CommandType_name, int32(c)), "COMMAND_TYPE_"))
}

// ValidateBasic returns an error if the given command type is invalid
func (c CommandType) ValidateBasic() error {
	if _, ok := CommandType_name[int32(c)]; !ok || c == COMMAND_TYPE_UNSPECIFIED {
		return fmt.Errorf("%s is not a valid command type", c.String())
	}
	return nil
}

var (
	stringType       = funcs.Must(abi.NewType("string", "string", nil))
	addressType      = funcs.Must(abi.NewType("address", "address", nil))
	addressesType    = funcs.Must(abi.NewType("address[]", "address[]", nil))
	bytes32Type      = funcs.Must(abi.NewType("bytes32", "bytes32", nil))
	uint8Type        = funcs.Must(abi.NewType("uint8", "uint8", nil))
	uint256Type      = funcs.Must(abi.NewType("uint256", "uint256", nil))
	uint256ArrayType = funcs.Must(abi.NewType("uint256[]", "uint256[]", nil))

	deployTokenArguments                 = abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}, {Type: addressType}, {Type: uint256Type}}
	mintTokenArguments                   = abi.Arguments{{Type: stringType}, {Type: addressType}, {Type: uint256Type}}
	burnTokenArguments                   = abi.Arguments{{Type: stringType}, {Type: bytes32Type}}
	transferMultisigArguments            = abi.Arguments{{Type: addressesType}, {Type: uint256ArrayType}, {Type: uint256Type}}
	approveContractCallArguments         = abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: addressType}, {Type: bytes32Type}, {Type: bytes32Type}, {Type: uint256Type}}
	approveContractCallWithMintArguments = abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: addressType}, {Type: bytes32Type}, {Type: stringType}, {Type: uint256Type}, {Type: bytes32Type}, {Type: uint256Type}}
)

// NewBurnTokenCommand creates a command to burn tokens with the given burner's information
func NewBurnTokenCommand(chainID sdk.Int, keyID multisig.KeyID, height int64, burnerInfo BurnerInfo, isTokenExternal bool) Command {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))

	burnTokenMaxGasCost := burnInternalTokenMaxGasCost
	if isTokenExternal {
		burnTokenMaxGasCost = burnExternalTokenMaxGasCost
	}

	return Command{
		ID:         NewCommandID(append(burnerInfo.Salt.Bytes(), heightBytes...), chainID),
		Type:       COMMAND_TYPE_BURN_TOKEN,
		Params:     createBurnTokenParams(burnerInfo.Symbol, common.Hash(burnerInfo.Salt)),
		KeyID:      keyID,
		MaxGasCost: uint32(burnTokenMaxGasCost),
	}
}

// NewDeployTokenCommand creates a command to deploy a token
func NewDeployTokenCommand(chainID sdk.Int, keyID multisig.KeyID, asset string, tokenDetails TokenDetails, address Address, dailyMintLimit sdk.Uint) Command {
	return Command{
		ID:         NewCommandID([]byte(fmt.Sprintf("%s_%s", asset, tokenDetails.Symbol)), chainID),
		Type:       COMMAND_TYPE_DEPLOY_TOKEN,
		Params:     createDeployTokenParams(tokenDetails.TokenName, tokenDetails.Symbol, tokenDetails.Decimals, tokenDetails.Capacity, address, dailyMintLimit),
		KeyID:      keyID,
		MaxGasCost: deployTokenMaxGasCost,
	}
}

// NewMintTokenCommand creates a command to mint token to the given address
func NewMintTokenCommand(keyID multisig.KeyID, id nexus.TransferID, symbol string, address common.Address, amount *big.Int) Command {
	return Command{
		ID:         CommandIDFromTransferID(id),
		Type:       COMMAND_TYPE_MINT_TOKEN,
		Params:     createMintTokenParams(symbol, address, amount),
		KeyID:      keyID,
		MaxGasCost: mintTokenMaxGasCost,
	}
}

// NewMultisigTransferCommand creates a command to transfer operator of the multisig contract
func NewMultisigTransferCommand(chainID sdk.Int, keyID multisig.KeyID, nextKey multisig.Key) Command {
	addresses, weights, threshold := GetMultisigAddressesAndWeights(nextKey)

	var concat []byte
	for _, addr := range addresses {
		concat = append(concat, addr.Bytes()...)
	}

	return Command{
		ID:         NewCommandID(concat, chainID),
		Type:       COMMAND_TYPE_TRANSFER_OPERATORSHIP,
		Params:     createTransferMultisigParams(addresses, slices.Map(weights, sdk.Uint.BigInt), threshold.BigInt()),
		KeyID:      keyID,
		MaxGasCost: transferOperatorshipMaxGasCost,
	}
}

// NewApproveContractCallCommand creates a command to approve contract call
func NewApproveContractCallCommand(
	chainID sdk.Int,
	keyID multisig.KeyID,
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCall,
) Command {
	sourceEventIndexBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sourceEventIndexBz, sourceEventIndex)

	return Command{
		ID:         NewCommandID(append(sourceTxID.Bytes(), sourceEventIndexBz...), chainID),
		Type:       COMMAND_TYPE_APPROVE_CONTRACT_CALL,
		Params:     createApproveContractCallParams(sourceChain, sourceTxID, sourceEventIndex, event),
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallMaxGasCost),
	}
}

// NewApproveContractCallCommandGeneric creates a command to approve contract call
func NewApproveContractCallCommandGeneric(
	chainID sdk.Int,
	keyID multisig.KeyID,
	contractAddress common.Address,
	payloadHash common.Hash,
	sourceTxID common.Hash,
	sourceChain nexus.ChainName,
	sender string,
	sourceEventIndex uint64,
	ID string,
) Command {
	commandID := NewCommandID([]byte(ID), chainID)
	return Command{
		ID:         commandID,
		Type:       COMMAND_TYPE_APPROVE_CONTRACT_CALL,
		Params:     createApproveContractCallParamsGeneric(contractAddress, payloadHash, sourceTxID, string(sourceChain), sender, sourceEventIndex),
		KeyID:      keyID,
		MaxGasCost: approveContractCallMaxGasCost,
	}
}

// NewApproveContractCallWithMintCommand creates a command to approve contract call with token being minted
func NewApproveContractCallWithMintCommand(
	chainID sdk.Int,
	keyID multisig.KeyID,
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCallWithToken,
	amount sdk.Uint,
	symbol string,
) Command {
	sourceEventIndexBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sourceEventIndexBz, sourceEventIndex)

	return Command{
		ID:         NewCommandID(append(sourceTxID.Bytes(), sourceEventIndexBz...), chainID),
		Type:       COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT,
		Params:     createApproveContractCallWithMintParams(sourceChain, sourceTxID, sourceEventIndex, event, amount, symbol),
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallWithMintMaxGasCost),
	}
}

// NewApproveContractCallWithMintGeneric creates a command to approve contract call with mint
func NewApproveContractCallWithMintGeneric(
	chainID sdk.Int,
	keyID multisig.KeyID,
	sourceTxID common.Hash,
	sourceEventIndex uint64,
	message nexus.GeneralMessage,
	symbol string,
) Command {
	commandID := NewCommandID([]byte(message.ID), chainID)
	contractAddress := common.HexToAddress(message.GetDestinationAddress())
	payloadHash := common.BytesToHash(message.PayloadHash)

	return Command{
		ID:         commandID,
		Type:       COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT,
		Params:     createApproveContractCallWithMintParamsGeneric(contractAddress, payloadHash, sourceTxID, message.Sender, sourceEventIndex, message.Asset.Amount.BigInt(), symbol),
		KeyID:      keyID,
		MaxGasCost: approveContractCallWithMintMaxGasCost,
	}
}

// DecodeParams returns the decoded parameters in the given command
func (m Command) DecodeParams() (map[string]string, error) {
	params := make(map[string]string)

	switch m.Type {
	case COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT:
		sourceChain, sourceAddress, contractAddress, payloadHash, symbol, amount, sourceTxID, sourceEventIndex := DecodeApproveContractCallWithMintParams(m.Params)

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["symbol"] = symbol
		params["amount"] = amount.String()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case COMMAND_TYPE_APPROVE_CONTRACT_CALL:
		sourceChain, sourceAddress, contractAddress, payloadHash, sourceTxID, sourceEventIndex := DecodeApproveContractCallParams(m.Params)

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case COMMAND_TYPE_DEPLOY_TOKEN:
		name, symbol, decs, cap, tokenAddress, dailyMintLimit := DecodeDeployTokenParams(m.Params)

		params["name"] = name
		params["symbol"] = symbol
		params["decimals"] = strconv.FormatUint(uint64(decs), 10)
		params["cap"] = cap.String()
		params["tokenAddress"] = tokenAddress.Hex()
		params["dailyMintLimit"] = dailyMintLimit.String()
	case COMMAND_TYPE_MINT_TOKEN:
		symbol, addr, amount := DecodeMintTokenParams(m.Params)

		params["symbol"] = symbol
		params["account"] = addr.Hex()
		params["amount"] = amount.String()
	case COMMAND_TYPE_BURN_TOKEN:
		symbol, salt := DecodeBurnTokenParams(m.Params)

		params["symbol"] = symbol
		params["salt"] = salt.Hex()
	case COMMAND_TYPE_TRANSFER_OPERATORSHIP:
		addresses, weights, threshold := DecodeTransferMultisigParams(m.Params)

		params["newOperators"] = strings.Join(slices.Map(addresses, common.Address.Hex), ";")
		params["newWeights"] = strings.Join(slices.Map(weights, func(w *big.Int) string { return w.String() }), ";")
		params["newThreshold"] = threshold.String()
	default:
		return nil, fmt.Errorf("unknown command type '%s'", m.Type)
	}

	return params, nil
}

// Clone returns an exacy copy of Command
func (m Command) Clone() Command {
	clone := Command{
		ID:         m.ID,
		Type:       m.Type,
		KeyID:      m.KeyID,
		MaxGasCost: m.MaxGasCost,
		Params:     make([]byte, len(m.Params)),
	}
	copy(clone.Params, m.Params)

	return clone
}

func createBurnTokenParams(symbol string, salt common.Hash) []byte {
	return funcs.Must(burnTokenArguments.Pack(symbol, salt))
}

func createDeployTokenParams(tokenName string, symbol string, decimals uint8, capacity sdk.Int, address Address, dailyMintLimit sdk.Uint) []byte {
	return funcs.Must(deployTokenArguments.Pack(
		tokenName,
		symbol,
		decimals,
		capacity.BigInt(),
		address,
		dailyMintLimit.BigInt(),
	))
}

func createMintTokenParams(symbol string, address common.Address, amount *big.Int) []byte {
	return funcs.Must(mintTokenArguments.Pack(symbol, address, amount))
}

func createTransferMultisigParams(addresses []common.Address, weights []*big.Int, threshold *big.Int) []byte {
	return funcs.Must(transferMultisigArguments.Pack(addresses, weights, threshold))
}

func createApproveContractCallParams(
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCall) []byte {
	return funcs.Must(approveContractCallArguments.Pack(
		sourceChain,
		event.Sender.Hex(),
		common.HexToAddress(event.ContractAddress),
		common.Hash(event.PayloadHash),
		common.Hash(sourceTxID),
		new(big.Int).SetUint64(sourceEventIndex),
	))
}

func createApproveContractCallParamsGeneric(
	contractAddress common.Address,
	payloadHash common.Hash,
	txID common.Hash,
	sourceChain string,
	sender string,
	sourceEventIndex uint64) []byte {

	return funcs.Must(approveContractCallArguments.Pack(
		sourceChain,
		sender,
		contractAddress,
		payloadHash,
		txID,
		new(big.Int).SetUint64(sourceEventIndex),
	))
}

func createApproveContractCallWithMintParams(
	sourceChain nexus.ChainName,
	sourceTxID Hash,
	sourceEventIndex uint64,
	event EventContractCallWithToken,
	amount sdk.Uint,
	symbol string) []byte {
	return funcs.Must(approveContractCallWithMintArguments.Pack(
		sourceChain,
		event.Sender.Hex(),
		common.HexToAddress(event.ContractAddress),
		common.Hash(event.PayloadHash),
		symbol,
		amount.BigInt(),
		common.Hash(sourceTxID),
		new(big.Int).SetUint64(sourceEventIndex),
	))
}

func createApproveContractCallWithMintParamsGeneric(
	contractAddress common.Address,
	payloadHash common.Hash,
	txID common.Hash,
	sender nexus.CrossChainAddress,
	sourceEventIndex uint64,
	amount *big.Int,
	symbol string) []byte {

	return funcs.Must(approveContractCallWithMintArguments.Pack(
		sender.Chain.Name,
		sender.Address,
		contractAddress,
		payloadHash,
		symbol,
		amount,
		txID,
		new(big.Int).SetUint64(sourceEventIndex),
	))
}

// DecodeApproveContractCallWithMintParams decodes the call arguments from the given contract call
func DecodeApproveContractCallWithMintParams(bz []byte) (string, string, common.Address, common.Hash, string, *big.Int, common.Hash, *big.Int) {
	params := funcs.Must(StrictDecode(approveContractCallWithMintArguments, bz))

	payloadHash := params[3].([common.HashLength]byte)
	sourceTxID := params[6].([common.HashLength]byte)

	return params[0].(string),
		params[1].(string),
		params[2].(common.Address),
		common.BytesToHash(payloadHash[:]),
		params[4].(string),
		params[5].(*big.Int),
		common.BytesToHash(sourceTxID[:]),
		params[7].(*big.Int)
}

// DecodeApproveContractCallParams decodes the call arguments from the given contract call
func DecodeApproveContractCallParams(bz []byte) (string, string, common.Address, common.Hash, common.Hash, *big.Int) {
	params := funcs.Must(StrictDecode(approveContractCallArguments, bz))

	payloadHash := params[3].([common.HashLength]byte)
	sourceTxID := params[4].([common.HashLength]byte)

	return params[0].(string),
		params[1].(string),
		params[2].(common.Address),
		common.BytesToHash(payloadHash[:]),
		common.BytesToHash(sourceTxID[:]),
		params[5].(*big.Int)
}

// DecodeMintTokenParams decodes the call arguments from the given contract call
func DecodeMintTokenParams(bz []byte) (string, common.Address, *big.Int) {
	params := funcs.Must(StrictDecode(mintTokenArguments, bz))

	return params[0].(string), params[1].(common.Address), params[2].(*big.Int)
}

// DecodeDeployTokenParams decodes the call arguments from the given contract call
func DecodeDeployTokenParams(bz []byte) (string, string, uint8, *big.Int, common.Address, sdk.Uint) {
	params := funcs.Must(StrictDecode(deployTokenArguments, bz))

	return params[0].(string), params[1].(string), params[2].(uint8), params[3].(*big.Int), params[4].(common.Address), sdk.NewUintFromBigInt(params[5].(*big.Int))
}

// DecodeBurnTokenParams decodes the call arguments from the given contract call
func DecodeBurnTokenParams(bz []byte) (string, common.Hash) {
	params := funcs.Must(StrictDecode(burnTokenArguments, bz))

	return params[0].(string), params[1].([common.HashLength]byte)
}

// DecodeTransferMultisigParams decodes the call arguments from the given contract call
func DecodeTransferMultisigParams(bz []byte) ([]common.Address, []*big.Int, *big.Int) {
	params := funcs.Must(StrictDecode(transferMultisigArguments, bz))

	return params[0].([]common.Address), params[1].([]*big.Int), params[2].(*big.Int)
}

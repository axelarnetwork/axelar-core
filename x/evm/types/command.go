package types

import (
	"encoding/binary"
	fmt "fmt"
	"math/big"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

const (
	axelarGatewayCommandMintToken                   = "mintToken"
	mintTokenMaxGasCost                             = 100000
	axelarGatewayCommandDeployToken                 = "deployToken"
	deployTokenMaxGasCost                           = 1400000
	axelarGatewayCommandBurnToken                   = "burnToken"
	burnExternalTokenMaxGasCost                     = 400000
	burnInternalTokenMaxGasCost                     = 120000
	axelarGatewayCommandTransferOperatorship        = "transferOperatorship"
	transferOperatorshipMaxGasCost                  = 120000
	axelarGatewayCommandApproveContractCallWithMint = "approveContractCallWithMint"
	approveContractCallWithMintMaxGasCost           = 100000
	axelarGatewayCommandApproveContractCall         = "approveContractCall"
	approveContractCallMaxGasCost                   = 100000
)

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
		Command:    axelarGatewayCommandBurnToken,
		Params:     createBurnTokenParams(burnerInfo.Symbol, common.Hash(burnerInfo.Salt)),
		KeyID:      keyID,
		MaxGasCost: uint32(burnTokenMaxGasCost),
	}
}

// NewDeployTokenCommand creates a command to deploy a token
func NewDeployTokenCommand(chainID sdk.Int, keyID multisig.KeyID, asset string, tokenDetails TokenDetails, address Address, dailyMintLimit sdk.Uint) Command {
	return Command{
		ID:         NewCommandID([]byte(fmt.Sprintf("%s_%s", asset, tokenDetails.Symbol)), chainID),
		Command:    axelarGatewayCommandDeployToken,
		Params:     createDeployTokenParams(tokenDetails.TokenName, tokenDetails.Symbol, tokenDetails.Decimals, tokenDetails.Capacity, address, dailyMintLimit),
		KeyID:      keyID,
		MaxGasCost: deployTokenMaxGasCost,
	}
}

// NewMintTokenCommand creates a command to mint token to the given address
func NewMintTokenCommand(keyID multisig.KeyID, id CommandID, symbol string, address common.Address, amount *big.Int) Command {
	return Command{
		ID:         id,
		Command:    axelarGatewayCommandMintToken,
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
		Command:    axelarGatewayCommandTransferOperatorship,
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
		Command:    axelarGatewayCommandApproveContractCall,
		Params:     createApproveContractCallParams(sourceChain, sourceTxID, sourceEventIndex, event),
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallMaxGasCost),
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
		Command:    axelarGatewayCommandApproveContractCallWithMint,
		Params:     createApproveContractCallWithMintParams(sourceChain, sourceTxID, sourceEventIndex, event, amount, symbol),
		KeyID:      keyID,
		MaxGasCost: uint32(approveContractCallWithMintMaxGasCost),
	}
}

// DecodeParams returns the decoded parameters in the given command
func (c Command) DecodeParams() (map[string]string, error) {
	params := make(map[string]string)

	switch c.Command {
	case axelarGatewayCommandApproveContractCallWithMint:
		sourceChain, sourceAddress, contractAddress, payloadHash, symbol, amount, sourceTxID, sourceEventIndex := decodeApproveContractCallWithMintParams(c.Params)

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["symbol"] = symbol
		params["amount"] = amount.String()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case axelarGatewayCommandApproveContractCall:
		sourceChain, sourceAddress, contractAddress, payloadHash, sourceTxID, sourceEventIndex := decodeApproveContractCallParams(c.Params)

		params["sourceChain"] = sourceChain
		params["sourceAddress"] = sourceAddress
		params["contractAddress"] = contractAddress.Hex()
		params["payloadHash"] = payloadHash.Hex()
		params["sourceTxHash"] = sourceTxID.Hex()
		params["sourceEventIndex"] = sourceEventIndex.String()
	case axelarGatewayCommandDeployToken:
		name, symbol, decs, cap, tokenAddress, dailyMintLimit := decodeDeployTokenParams(c.Params)

		params["name"] = name
		params["symbol"] = symbol
		params["decimals"] = strconv.FormatUint(uint64(decs), 10)
		params["cap"] = cap.String()
		params["tokenAddress"] = tokenAddress.Hex()
		params["dailyMintLimit"] = dailyMintLimit.String()
	case axelarGatewayCommandMintToken:
		symbol, addr, amount := decodeMintTokenParams(c.Params)

		params["symbol"] = symbol
		params["account"] = addr.Hex()
		params["amount"] = amount.String()
	case axelarGatewayCommandBurnToken:
		symbol, salt := decodeBurnTokenParams(c.Params)

		params["symbol"] = symbol
		params["salt"] = salt.Hex()
	case axelarGatewayCommandTransferOperatorship:
		addresses, weights, threshold := decodeTransferMultisigParams(c.Params)

		params["newOperators"] = strings.Join(slices.Map(addresses, common.Address.Hex), ";")
		params["newWeights"] = strings.Join(slices.Map(weights, func(w *big.Int) string { return w.String() }), ";")
		params["newThreshold"] = threshold.String()
	default:
		return nil, fmt.Errorf("unknown command type '%s'", c.Command)
	}

	return params, nil
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

func decodeApproveContractCallWithMintParams(bz []byte) (string, string, common.Address, common.Hash, string, *big.Int, common.Hash, *big.Int) {
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

func decodeApproveContractCallParams(bz []byte) (string, string, common.Address, common.Hash, common.Hash, *big.Int) {
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

func decodeMintTokenParams(bz []byte) (string, common.Address, *big.Int) {
	params := funcs.Must(StrictDecode(mintTokenArguments, bz))

	return params[0].(string), params[1].(common.Address), params[2].(*big.Int)
}

func decodeDeployTokenParams(bz []byte) (string, string, uint8, *big.Int, common.Address, sdk.Uint) {
	params := funcs.Must(StrictDecode(deployTokenArguments, bz))

	return params[0].(string), params[1].(string), params[2].(uint8), params[3].(*big.Int), params[4].(common.Address), sdk.NewUintFromBigInt(params[5].(*big.Int))
}

func decodeBurnTokenParams(bz []byte) (string, common.Hash) {
	params := funcs.Must(StrictDecode(burnTokenArguments, bz))

	return params[0].(string), params[1].([common.HashLength]byte)
}

func decodeTransferMultisigParams(bz []byte) ([]common.Address, []*big.Int, *big.Int) {
	params := funcs.Must(StrictDecode(transferMultisigArguments, bz))

	return params[0].([]common.Address), params[1].([]*big.Int), params[2].(*big.Int)
}

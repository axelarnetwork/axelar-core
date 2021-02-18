package keeper

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/tendermint/libs/log"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	rawPrefix           = "raw_"
	verifiedTxPrefix    = "verified_tx_"
	pendingTxPrefix     = "pending_tx_"
	verifiedTokenPrefix = "verified_token_"
	pendingTokenPrefix  = "pending_token_"
	commandPrefix       = "command_"
	symbolPrefix        = "symbol_"
	burnerPrefix        = "burner_"
)

// Keeper represents the ethereum keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

// NewEthKeeper returns a new ethereum keeper
func NewEthKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// SetParams sets the eth module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// GetParams gets the eth module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// GetNetwork returns the Ethereum network Axelar-Core is expected to connect to
func (k Keeper) GetNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

// GetERC20TransferSignature returns the signature of the ERC20 transfer method
func (k Keeper) GetERC20TransferSignature(ctx sdk.Context) common.Hash {
	var transferSig []byte
	k.params.Get(ctx, types.KeyNetwork, &transferSig)
	return common.BytesToHash(transferSig)
}

// Codec returns the codec
func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k Keeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

// SetBurnerInfo saves the burner info for a given address
func (k Keeper) SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *types.BurnerInfo) {
	key := append([]byte(burnerPrefix), burnerAddr.Bytes()...)
	bz := k.cdc.MustMarshalJSON(burnerInfo)

	ctx.KVStore(k.storeKey).Set(key, bz)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k Keeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := append([]byte(burnerPrefix), burnerAddr.Bytes()...)

	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return nil
	}

	var result *types.BurnerInfo
	k.cdc.MustUnmarshalJSON(bz, &result)

	return result
}

// GetTokenAddress calculates the token address given symbol and axelar gateway address
func (k Keeper) GetTokenAddress(ctx sdk.Context, symbol string, gatewayAddr common.Address) (common.Address, error) {
	tokenInfo := k.getTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return common.Address{}, fmt.Errorf("symbol not found/verified")
	}

	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(symbol)).Bytes())

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return common.Address{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return common.Address{}, err
	}

	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return common.Address{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	packed, err := arguments.Pack(tokenInfo.TokenName, symbol, tokenInfo.Decimals, tokenInfo.Capacity.BigInt())
	if err != nil {
		return common.Address{}, err
	}

	tokenInitCode := append(k.getTokenBC(ctx), packed...)
	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)

	return crypto.CreateAddress2(gatewayAddr, saltToken, tokenInitCodeHash.Bytes()), nil
}

// GetBurnerAddressAndSalt calculates a burner address and the corresponding salt for the given token address, recipient and axelar gateway address
func (k Keeper) GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, [32]byte, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return common.Address{}, [32]byte{}, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return common.Address{}, [32]byte{}, err
	}

	var saltBurn [32]byte
	copy(saltBurn[:], crypto.Keccak256Hash([]byte(recipient)).Bytes())

	arguments := abi.Arguments{{Type: addressType}, {Type: bytes32Type}}
	packed, err := arguments.Pack(tokenAddr, saltBurn)
	if err != nil {
		return common.Address{}, [32]byte{}, err
	}

	burnerInitCode := append(k.getBurnerBC(ctx), packed...)
	burnerInitCodeHash := crypto.Keccak256Hash(burnerInitCode)

	return crypto.CreateAddress2(gatewayAddr, saltBurn, burnerInitCodeHash.Bytes()), saltBurn, nil
}

func (k Keeper) getBurnerBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyBurnable, &b)
	return b
}

func (k Keeper) getTokenBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyToken, &b)
	return b
}

// SetUnverifiedErc20TokenDeploy stores and unverified erc20 token
func (k Keeper) SetUnverifiedErc20TokenDeploy(ctx sdk.Context, token *types.Erc20TokenDeploy) {
	bz := k.cdc.MustMarshalJSON(token)
	ctx.KVStore(k.storeKey).Set([]byte(pendingTokenPrefix+token.TxID.String()), bz)
}

// SaveTokenInfo stores the token info
func (k Keeper) SaveTokenInfo(ctx sdk.Context, msg types.MsgSignDeployToken) {
	bz := k.cdc.MustMarshalJSON(msg)
	ctx.KVStore(k.storeKey).Set([]byte(symbolPrefix+msg.Symbol), bz)
}

func (k Keeper) getTokenInfo(ctx sdk.Context, symbol string) *types.MsgSignDeployToken {
	bz := ctx.KVStore(k.storeKey).Get([]byte(symbolPrefix + symbol))
	if bz == nil {
		return nil
	}
	var msg *types.MsgSignDeployToken
	k.cdc.MustUnmarshalJSON(bz, &msg)

	return msg
}

// SetCommandData stores command data by ID
func (k Keeper) SetCommandData(ctx sdk.Context, commandID types.CommandID, commandData []byte) {
	key := append([]byte(commandPrefix), commandID[:]...)

	ctx.KVStore(k.storeKey).Set(key, commandData)
}

// GetCommandData retrieves command data by ID
func (k Keeper) GetCommandData(ctx sdk.Context, commandID types.CommandID) []byte {
	key := append([]byte(commandPrefix), commandID[:]...)

	return ctx.KVStore(k.storeKey).Get(key)
}

func (k Keeper) getRawTx(ctx sdk.Context, txID string) *ethTypes.Transaction {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txID))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalJSON(bz, &tx)

	return tx
}

// SetRawTx stores a raw transaction by hash
func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

// HasVerifiedTx returns true if a raw transaction has been stored
func (k Keeper) HasVerifiedTx(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(verifiedTxPrefix + txID))
}

// SetUnverifiedTx stores and unverified transaction
func (k Keeper) SetUnverifiedTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	ctx.KVStore(k.storeKey).Set([]byte(pendingTxPrefix+txID), tx.Hash().Bytes())
}

// HasUnverifiedTx returns true if an unverified transaction has been stored
func (k Keeper) HasUnverifiedTx(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingTxPrefix + txID))
}

// ProcessVerificationResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, txID, pollType string, verified bool) {

	var pendingKey []byte
	var verifiedKey []byte

	switch pollType {
	case types.PollVerifyToken:
		pendingKey = []byte(pendingTokenPrefix + txID)
		verifiedKey = []byte(verifiedTokenPrefix + txID)
	case types.PollVerifyTx:
		pendingKey = []byte(pendingTxPrefix + txID)
		verifiedKey = []byte(verifiedTxPrefix + txID)
	default:
		k.Logger(ctx).Debug(fmt.Sprintf("unknown verification type: %s", pollType))
		return
	}

	bz := ctx.KVStore(k.storeKey).Get(pendingKey)
	if bz == nil {
		k.Logger(ctx).Debug(fmt.Sprintf("tx %s not found", txID))
		return
	}
	if verified {
		ctx.KVStore(k.storeKey).Set(verifiedKey, bz)
	}
	ctx.KVStore(k.storeKey).Delete(pendingKey)
	return
}

// AssembleEthTx sets a signature for a previously stored raw transaction
func (k Keeper) AssembleEthTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*ethTypes.Transaction, error) {
	rawTx := k.getRawTx(ctx, txID)
	if rawTx == nil {
		return nil, fmt.Errorf("raw tx for ID %s has not been prepared yet", txID)
	}

	signer := k.getSigner(ctx)

	recoverableSig, err := types.ToEthSignature(sig, signer.Hash(rawTx), pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	return rawTx.WithSignature(signer, recoverableSig[:])
}

// GetHashToSign returns the hash to sign of a previously stored raw transaction
func (k Keeper) GetHashToSign(ctx sdk.Context, txID string) (common.Hash, error) {
	rawTx := k.getRawTx(ctx, txID)
	if rawTx == nil {
		return common.Hash{}, fmt.Errorf("raw tx with id %s not found", txID)
	}
	signer := k.getSigner(ctx)
	return signer.Hash(rawTx), nil
}

func (k Keeper) getSigner(ctx sdk.Context) ethTypes.EIP155Signer {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return ethTypes.NewEIP155Signer(network.Params().ChainID)
}

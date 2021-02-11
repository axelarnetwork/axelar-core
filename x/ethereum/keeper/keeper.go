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
	rawPrefix     = "raw_"
	txPrefix      = "tx_"
	pendingPrefix = "pend_"
	commandPrefix = "command_"
	symbolPrefix  = "symbol_"
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

// GetBurnerAddress calculates a burner address for the given symbol and recipient
func (k Keeper) GetBurnerAddress(ctx sdk.Context, symbol, recipient string, gatewayAddr common.Address) (common.Address, error) {
	tokenInfo := k.getTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, "symbol not found/verified")

	}

	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(symbol)).Bytes())

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}
	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}
	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}
	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	packed, err := arguments.Pack(tokenInfo.TokenName, symbol, tokenInfo.Decimals, tokenInfo.Capacity.BigInt())
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	tokenInitCode := k.getTokenBC(ctx)
	tokenInitCode = append(tokenInitCode, packed...)

	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)
	tokenAddr := crypto.CreateAddress2(gatewayAddr, saltToken, tokenInitCodeHash.Bytes())

	var saltBurn [32]byte
	copy(saltBurn[:], crypto.Keccak256Hash([]byte(recipient)).Bytes())

	arguments = abi.Arguments{{Type: addressType}, {Type: bytes32Type}}
	packed, err = arguments.Pack(tokenAddr, saltBurn)
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	burnerInitCode := k.getBurnerBC(ctx)
	burnerInitCode = append(burnerInitCode, packed...)

	burnerInitCodeHash := crypto.Keccak256Hash(burnerInitCode)
	return crypto.CreateAddress2(gatewayAddr, saltBurn, burnerInitCodeHash.Bytes()), nil

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
	return ctx.KVStore(k.storeKey).Has([]byte(txPrefix + txID))
}

// SetUnverifiedTx stores and unverified transaction
func (k Keeper) SetUnverifiedTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+txID), tx.Hash().Bytes())
}

// HasUnverifiedTx returns true if an unverified transaction has been stored
func (k Keeper) HasUnverifiedTx(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingPrefix + txID))
}

// ProcessVerificationResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, txID string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + txID))
	if bz == nil {
		return fmt.Errorf("tx %s not found", txID)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(txPrefix+txID), bz)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + txID))
	return nil
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

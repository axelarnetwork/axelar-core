package keeper

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	rawPrefix           = "raw_"
	symbolPrefix        = "symbol_"
	txPrefix            = "tx_"
	pendingSymbolPrefix = "pend_symbol_"
	pendingTXPrefix     = "pend_tx_"
	commandPrefix       = "command_"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

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

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

func (k Keeper) SetCommandData(ctx sdk.Context, commandID types.CommandID, commandData []byte) {
	key := append([]byte(commandPrefix), commandID[:]...)

	ctx.KVStore(k.storeKey).Set(key, commandData)
}

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

func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

func (k Keeper) HasVerifiedTx(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(txPrefix + txID))
}

func (k Keeper) SetUnverifiedTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	ctx.KVStore(k.storeKey).Set([]byte(pendingTXPrefix+txID), tx.Hash().Bytes())
}

func (k Keeper) HasUnverifiedTx(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingTXPrefix + txID))
}

func (k Keeper) GetContractAddress(ctx sdk.Context, symbol string) (common.Address, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingSymbolPrefix + symbol))
	if bz == nil {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}

func (k Keeper) HasVerifiedSymbol(ctx sdk.Context, symbol string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(symbolPrefix + symbol))
}

func (k Keeper) SetUnverifiedSymbol(ctx sdk.Context, symbol string, addr common.Address) {
	ctx.KVStore(k.storeKey).Set([]byte(pendingSymbolPrefix+symbol), addr.Bytes())
}

func (k Keeper) HasUnverifiedSymbol(ctx sdk.Context, symbol string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingSymbolPrefix + symbol))
}

// ProcessVerificationResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessTxVerificationResult(ctx sdk.Context, txID string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingTXPrefix + txID))
	if bz == nil {
		return fmt.Errorf("tx %s not found", txID)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(txPrefix+txID), bz)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pendingTXPrefix + txID))
	return nil
}

func (k Keeper) ProcessSymbolVerificationResult(ctx sdk.Context, symbol string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingSymbolPrefix + symbol))
	if bz == nil {
		return fmt.Errorf("symbol %s not found", symbol)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(symbolPrefix+symbol), bz)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pendingSymbolPrefix + symbol))
	return nil
}

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

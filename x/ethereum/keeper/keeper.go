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
	symbolPrefix        = "symbol_"
	txPrefix            = "tx_"
	pendingSymbolPrefix = "pend_symbol_"
	pendingTXPrefix     = "pend_tx_"
	commandPrefix       = "command_"

	gatewayKey = "gateway"
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

func (k Keeper) GetBurnerAddress(ctx sdk.Context, symbol string, recipient string) (common.Address, error) {
	tokenInfo := k.getTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, "symbol not found/verified")

	}

	gatewayAddr, ok := k.getGatewayAddress(ctx)
	if !ok {
		return common.Address{}, sdkerrors.Wrap(types.ErrEthereum, "gateway not set")

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

	burnerInitCode := k.getBurnerBC(ctx)
	burnerInitCode = append(burnerInitCode, common.LeftPadBytes(tokenAddr.Bytes(), 32)...)
	burnerInitCode = append(burnerInitCode, common.LeftPadBytes(saltBurn[:], 32)...)

	burnerInitCodeHash := crypto.Keccak256Hash(burnerInitCode)
	return crypto.CreateAddress2(gatewayAddr, saltBurn, burnerInitCodeHash.Bytes()), nil

}

func (k Keeper) getBurnerBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyBurneable, &b)
	return b
}

func (k Keeper) getTokenBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyToken, &b)
	return b
}

func (k Keeper) getGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(gatewayKey))
	if bz == nil {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}

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

func (k Keeper) getContractAddress(ctx sdk.Context, symbol string) (common.Address, bool) {
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

// ProcessTxVerificationResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessTxVerificationResult(ctx sdk.Context, txID string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingTXPrefix + txID))
	if bz == nil {
		return fmt.Errorf("tx %s not found", txID)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(txPrefix+txID), bz)

		// calculate contract address of the verified tx and store it
		// as the address for the axelar gateway
		var tx *ethTypes.Transaction
		k.cdc.MustUnmarshalJSON(bz, &tx)

		_, r, s := tx.RawSignatureValues()
		sig := types.Signature{}
		copy(sig[:32], common.LeftPadBytes(r.Bytes(), 32))
		copy(sig[32:], common.LeftPadBytes(s.Bytes(), 32))

		pubKey, err := crypto.SigToPub(tx.Hash().Bytes(), sig[:])
		if err != nil {
			return err
		}

		contractAddr := crypto.CreateAddress(crypto.PubkeyToAddress(*pubKey), tx.Nonce())
		ctx.KVStore(k.storeKey).Set([]byte(gatewayKey+txID), contractAddr.Bytes())
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

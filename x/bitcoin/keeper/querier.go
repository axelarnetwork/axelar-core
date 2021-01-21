package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	QueryDepositAddress       = "depositAddr"
	QueryConsolidationAddress = "consolidationAddr"
	// QueryOutInfo is the route to query for a transaction's outPoint information
	QueryOutInfo = "outPointInfo"
	// QueryRawTx is the route to query for an unsigned raw transaction
	QueryRawTx = "rawTx"
	SendTx     = "sendTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, b types.Balancer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryDepositAddress:
			return queryDepositAddress(ctx, k, s, req.Data)
		case QueryConsolidationAddress:
			return queryConsolidationAddress(ctx, k, b, s, req.Data)
		case QueryOutInfo:
			return queryTxOutInfo(rpc, req.Data)
		case QueryRawTx:
			return createRawTx(ctx, k, req.Data)
		case SendTx:
			return sendTx(ctx, k, b, rpc, s, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}
	}
}

func queryDepositAddress(ctx sdk.Context, k Keeper, s types.Signer, data []byte) ([]byte, error) {
	var recipient balance.CrossChainAddress
	err := types.ModuleCdc.UnmarshalJSON(data, &recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse the recipient")
	}

	pk, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "key not found")
	}

	addr, err := k.GenerateDepositAddress(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryConsolidationAddress(ctx sdk.Context, k Keeper, b types.Balancer, s types.Signer, data []byte) ([]byte, error) {
	var oldAddr balance.CrossChainAddress
	err := types.ModuleCdc.UnmarshalJSON(data, &oldAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "could not parse the current address")
	}
	if oldAddr.Chain != balance.Bitcoin {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "the current address must be a bitcoin address")
	}

	recipient, ok := b.GetRecipient(ctx, oldAddr)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "the current address is not linked to any cross-chain recipient")
	}

	pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr, err := k.GenerateDepositAddress(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryTxOutInfo(rpc types.RPCClient, data []byte) ([]byte, error) {
	var out *wire.OutPoint
	err := types.ModuleCdc.UnmarshalJSON(data, &out)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse the outpoint")
	}
	info, err := rpc.GetOutPointInfo(out)
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MustMarshalJSON(info), nil
}

func createRawTx(ctx sdk.Context, k Keeper, data []byte) ([]byte, error) {
	var params types.RawTxParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	recipient, err := btcutil.DecodeAddress(params.DepositAddr, k.getNetwork(ctx).Params())
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	outPoint, ok := k.GetVerifiedOutPointInfo(ctx, params.OutPoint)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "no outpoint of tx %s not found")
	}
	tx, err := types.CreateTx(outPoint.OutPoint, params.Satoshi, recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}
	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

func sendTx(ctx sdk.Context, k Keeper, b types.Balancer, rpc types.RPCClient, s types.Signer, txID string) ([]byte, error) {
	rawTx := k.GetRawTx(ctx, txID)
	if rawTx == nil {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "withdraw tx for ID %s has not been prepared yet", txID)
	}

	h, err := k.GetHashToSign(ctx, rawTx)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}
	sigID := hex.EncodeToString(h)
	key, ok := s.GetKeyForSigID(ctx, sigID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "could not find a corresponding key for tx ID %s", txID)
	}
	pk := btcec.PublicKey(key)

	sig, ok := s.GetSig(ctx, sigID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "signature not found")
	}
	btcSig := btcec.Signature{
		R: sig.R,
		S: sig.S,
	}

	info, ok := k.GetVerifiedOutPointInfo(ctx, rawTx.TxIn[0].PreviousOutPoint)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "verified outpoint info not found")
	}

	recipient, ok := b.GetRecipient(ctx, balance.CrossChainAddress{Chain: balance.Bitcoin, Address: info.DepositAddr})
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "recipient for deposit not found")
	}

	tx, err := k.AssembleBtcTx(ctx, rawTx, pk, btcSig, recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	// This is beyond axelar's control, so we can only log the error and move on regardless
	hash, err := rpc.SendRawTransaction(tx, false)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "sending transaction to Bitcoin failed").Error())
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return k.Codec().MustMarshalJSON(fmt.Sprintf("successfully sent transaction %s to Bitcoin", hash.String())), nil
}

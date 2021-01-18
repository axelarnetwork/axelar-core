package keeper

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	QueryMasterAddress = "masterAddr"
	// QueryOutInfo is the route to query for a transaction's outPoint information
	QueryOutInfo = "outPointInfo"
	// QueryRawTx is the route to query for an unsigned raw transaction
	QueryRawTx = "rawTx"
	SendTx     = "sendTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryMasterAddress:
			return queryMasterAddress(ctx, k, s)
		case QueryOutInfo:
			return queryTxOutInfo(rpc, path[1], path[2])
		case QueryRawTx:
			return createRawTx(ctx, k, s, req.Data)
		case SendTx:
			return sendTx(ctx, k, rpc, s, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}
	}
}

func queryMasterAddress(ctx sdk.Context, k Keeper, s types.Signer) ([]byte, error) {
	pk, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr, err := k.GetAddress(ctx, btcec.PublicKey(pk), balance.CrossChainAddress{})
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryTxOutInfo(rpc types.RPCClient, txID string, voutIdx string) ([]byte, error) {
	v, err := strconv.ParseUint(voutIdx, 10, 32)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse voutIdx")
	}
	hash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return nil, err
	}

	info, err := rpc.GetOutPointInfo(wire.NewOutPoint(hash, uint32(v)))
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MustMarshalJSON(info), nil
}

func createRawTx(ctx sdk.Context, k Keeper, s types.Signer, data []byte) ([]byte, error) {
	var params types.RawParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	var recipient btcutil.Address
	if params.Recipient != "" {
		recipient, err = btcutil.DecodeAddress(params.Recipient, k.getNetwork(ctx).Params())
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
	} else {
		pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "next master key not set")
		}
		recipient, err = k.GetAddress(ctx, btcec.PublicKey(pk), balance.CrossChainAddress{})
	}

	tx, err := k.CreateTx(ctx, params.TxID, params.Satoshi, recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}
	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

func sendTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, txID string) ([]byte, error) {
	h, err := k.GetHashToSign(ctx, txID)
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

	tx, err := k.AssembleBtcTx(ctx, txID, pk, btcSig)
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

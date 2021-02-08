package keeper

import (
	"encoding/hex"
	"fmt"

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

// Query paths
const (
	QueryDepositAddress       = "depositAddr"
	QueryConsolidationAddress = "consolidationAddr"
	QueryOutInfo              = "outPointInfo"
	QueryRawTx                = "rawTx"
	SendTx                    = "sendTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, b types.Balancer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryDepositAddress:
			return queryDepositAddress(ctx, k, s, req.Data)
		case QueryConsolidationAddress:
			return queryConsolidationAddress(ctx, k, b, s, path[1])
		case QueryOutInfo:
			blockHash, err := chainhash.NewHashFromStr(path[1])
			if err != nil {
				return nil, sdkerrors.Wrapf(types.ErrBitcoin, "could not parse block hash: %s", err.Error())
			}
			return queryTxOutInfo(rpc, blockHash, req.Data)
		case QueryRawTx:
			return createRawTx(ctx, k, req.Data)
		case SendTx:
			return sendTx(ctx, k, rpc, s, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}
	}
}

func queryDepositAddress(ctx sdk.Context, k Keeper, s types.Signer, data []byte) ([]byte, error) {
	var recipient balance.CrossChainAddress
	if err := types.ModuleCdc.UnmarshalJSON(data, &recipient); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "could not parse the recipient")
	}

	pk, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "key not found")
	}

	addr, _, err := k.GenerateDepositAddressAndRedeemScript(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryConsolidationAddress(ctx sdk.Context, k Keeper, b types.Balancer, s types.Signer, currAddr string) ([]byte, error) {
	recipient, ok := b.GetRecipient(ctx, balance.CrossChainAddress{Chain: balance.Bitcoin, Address: currAddr})
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "the current address is not linked to any cross-chain recipient")
	}

	pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr, _, err := k.GenerateDepositAddressAndRedeemScript(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryTxOutInfo(rpc types.RPCClient, blockHash *chainhash.Hash, data []byte) ([]byte, error) {
	var out *wire.OutPoint
	err := types.ModuleCdc.UnmarshalJSON(data, &out)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, sdkerrors.Wrap(err, "could not parse the outpoint").Error())
	}
	info, err := rpc.GetOutPointInfo(blockHash, out)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return types.ModuleCdc.MustMarshalJSON(info), nil
}

func createRawTx(ctx sdk.Context, k Keeper, data []byte) ([]byte, error) {
	var params types.RawTxParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	recipient, err := btcutil.DecodeAddress(params.DepositAddr, k.getNetwork(ctx).Params)
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

func sendTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, data []byte) ([]byte, error) {
	var out *wire.OutPoint
	err := types.ModuleCdc.UnmarshalJSON(data, &out)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, sdkerrors.Wrap(err, "could not parse the outpoint").Error())
	}
	rawTx := k.GetRawTx(ctx, out)
	if rawTx == nil {
		return nil, sdkerrors.Wrapf(types.ErrBitcoin, "withdraw tx for outpoint %s has not been prepared yet", out)
	}

	h, err := k.GetHashToSign(ctx, rawTx)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	sigID := hex.EncodeToString(h)
	sig, ok := s.GetSig(ctx, sigID)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "signature not found")
	}
	btcSig := btcec.Signature{
		R: sig.R,
		S: sig.S,
	}

	tx, err := k.AssembleBtcTx(ctx, rawTx, btcSig)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	// This is beyond axelar's control, so we can only log the error and move on regardless
	hash, err := rpc.SendRawTransaction(tx, false)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "sending transaction to Bitcoin failed").Error())
		return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
	}

	return k.Codec().MustMarshalJSON(hash), nil
}

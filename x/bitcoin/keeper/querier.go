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
	SendTransfers             = "sendTransfers"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, b types.Balancer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, req.Data)
		case QueryConsolidationAddress:
			return queryConsolidationAddress(ctx, k, b, s, path[1])
		case QueryOutInfo:
			blockHash, err := chainhash.NewHashFromStr(path[1])
			if err != nil {
				return nil, sdkerrors.Wrapf(types.ErrBitcoin, "could not parse block hash: %s", err.Error())
			}
			res, err = queryTxOutInfo(rpc, blockHash, req.Data)
		case QueryRawTx:
			res, err = createRawTx(ctx, k, req.Data)
		case SendTx:
			res, err = sendTx(ctx, k, rpc, s, req.Data)
		case SendTransfers:
			res, err = sendTransferTx(ctx, k, rpc, s)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryDepositAddress(ctx sdk.Context, k Keeper, s types.Signer, data []byte) ([]byte, error) {
	var recipient balance.CrossChainAddress
	if err := types.ModuleCdc.UnmarshalJSON(data, &recipient); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	pk, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr, _, err := k.GenerateDepositAddressAndRedeemScript(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, err
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryConsolidationAddress(ctx sdk.Context, k Keeper, b types.Balancer, s types.Signer, currAddr string) ([]byte, error) {
	recipient, ok := b.GetRecipient(ctx, balance.CrossChainAddress{Chain: balance.Bitcoin, Address: currAddr})
	if !ok {
		return nil, fmt.Errorf("the current address is not linked to any cross-chain recipient")
	}

	pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	addr, _, err := k.GenerateDepositAddressAndRedeemScript(ctx, btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, err
	}

	return []byte(addr.EncodeAddress()), nil
}

func queryTxOutInfo(rpc types.RPCClient, blockHash *chainhash.Hash, data []byte) ([]byte, error) {
	var out *wire.OutPoint
	err := types.ModuleCdc.UnmarshalJSON(data, &out)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse the outpoint")
	}
	info, err := rpc.GetOutPointInfo(blockHash, out)
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MustMarshalJSON(info), nil
}

func createRawTx(ctx sdk.Context, k Keeper, data []byte) ([]byte, error) {
	var params types.RawTxParams
	err := types.ModuleCdc.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, err
	}

	recipient, err := btcutil.DecodeAddress(params.DepositAddr, k.GetNetwork(ctx).Params)
	if err != nil {
		return nil, err
	}

	outPoint, ok := k.GetVerifiedOutPointInfo(ctx, params.OutPoint)
	if !ok {
		return nil, fmt.Errorf("verified outpoint %s not found", params.OutPoint.String())
	}
	tx, err := types.CreateTx([]*wire.OutPoint{outPoint.OutPoint}, []types.Output{{btcutil.Amount(params.Satoshi.Amount.Int64()), recipient}})
	if err != nil {
		return nil, err
	}
	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

func sendTransferTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer) ([]byte, error) {
	rawTx := k.GetRawConsolidationTx(ctx)
	if rawTx == nil {
		return nil, fmt.Errorf("no consolidation transaction found")
	}

	hash, err := send(ctx, k, rpc, s, rawTx)
	if err != nil {
		return nil, err
	}

	return k.Codec().MustMarshalJSON(hash), nil
}

func send(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, rawTx *wire.MsgTx) (*chainhash.Hash, error) {
	hashes, err := k.GetHashesToSign(ctx, rawTx)
	if err != nil {
		return nil, err
	}

	var sigs []btcec.Signature
	for _, hash := range hashes {
		sigID := hex.EncodeToString(hash)
		sig, ok := s.GetSig(ctx, sigID)
		if !ok {
			return nil, fmt.Errorf("signature not found")
		}
		sigs = append(sigs, btcec.Signature{R: sig.R, S: sig.S})
	}

	tx, err := k.AssembleBtcTx(ctx, rawTx, sigs)
	if err != nil {
		return nil, err
	}

	// This is beyond axelar's control, so we can only log the error and move on regardless
	hash, err := rpc.SendRawTransaction(tx, false)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, "sending transaction to Bitcoin failed").Error())
		return nil, err
	}
	return hash, nil
}

func sendTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer, data []byte) ([]byte, error) {
	var out *wire.OutPoint
	err := types.ModuleCdc.UnmarshalJSON(data, &out)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, sdkerrors.Wrap(err, "could not parse the outpoint").Error())
	}
	rawTx := k.GetRawTx(ctx, out)
	if rawTx == nil {
		return nil, fmt.Errorf("withdraw tx for outpoint %s has not been prepared yet", out)
	}
	hash, err := send(ctx, k, rpc, s, rawTx)
	if err != nil {
		return nil, err
	}
	return k.Codec().MustMarshalJSON(hash), nil
}

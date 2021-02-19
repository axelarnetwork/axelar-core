package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// Query paths
const (
	QueryDepositAddress = "depositAddr"
	QueryOutInfo        = "outPointInfo"
	SendTx              = "sendTransfers"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, b types.Balancer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, b, req.Data)
		case QueryOutInfo:
			blockHash, err := chainhash.NewHashFromStr(path[1])
			if err != nil {
				return nil, sdkerrors.Wrapf(types.ErrBitcoin, "could not parse block hash: %s", err.Error())
			}
			res, err = queryTxOutInfo(rpc, blockHash, req.Data)
		case SendTx:
			res, err = sendTx(ctx, k, rpc, s)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryDepositAddress(ctx sdk.Context, k Keeper, s types.Signer, b types.Balancer, data []byte) ([]byte, error) {
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	chain, ok := b.GetChain(ctx, params.Chain)
	if !ok {
		return nil, fmt.Errorf("recipient chain not found")
	}

	recipient := balance.CrossChainAddress{Chain: chain, Address: params.Address}

	pk, ok := s.GetCurrentMasterKey(ctx, exported.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	script, err := types.CreateCrossChainRedeemScript(btcec.PublicKey(pk), recipient)
	if err != nil {
		return nil, err
	}
	addr, err := types.CreateDepositAddress(k.GetNetwork(ctx), script)
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

func sendTx(ctx sdk.Context, k Keeper, rpc types.RPCClient, s types.Signer) ([]byte, error) {
	rawTx := k.GetRawTx(ctx)
	if rawTx == nil {
		return nil, fmt.Errorf("no consolidation transaction found")
	}

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

	return k.Codec().MustMarshalJSON(hash), nil
}

package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Query paths
const (
	QueryDepositAddress = "depositAddr"
	SendTx              = "sendTransfers"
	GetTx               = "getTransferTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, n types.Nexus, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, n, req.Data)
		case SendTx:
			res, err = sendTx(ctx, k, rpc, s)
		case GetTx:
			res, err = getConsolidationTx(ctx, k, s)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryDepositAddress(ctx sdk.Context, k Keeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	chain, ok := n.GetChain(ctx, params.Chain)
	if !ok {
		return nil, fmt.Errorf("recipient chain not found")
	}

	recipient := nexus.CrossChainAddress{Chain: chain, Address: params.Address}

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

	return k.Codec().MustMarshalJSON(hash.String()), nil
}

func getConsolidationTx(ctx sdk.Context, k Keeper, s types.Signer) ([]byte, error) {
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

	return k.Codec().MustMarshalJSON(tx), nil
}

package keeper

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	// QueryOutInfo is the route to query for a transaction's outPoint information
	QueryOutInfo = "outPointInfo"
	// QueryRawTx is the route to query for an unsigned raw transaction
	QueryRawTx = "rawTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k Keeper, s types.Signer, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryOutInfo:
			return queryTxInfo(rpc, path[1], path[2])
		case QueryRawTx:
			if len(path) == 4 {
				return createRawTx(ctx, k, s, rpc, path[1], path[2], path[3])
			} else {
				return createRawTx(ctx, k, s, rpc, path[1], path[2], "")
			}

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}
	}
}

func queryTxInfo(rpc types.RPCClient, txID string, voutIdx string) ([]byte, error) {
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

func createRawTx(ctx sdk.Context, k Keeper, s types.Signer, rpc types.RPCClient, txID string, amountStr string, recipientAddr string) ([]byte, error) {
	out, ok := k.GetVerifiedOutPoint(ctx, txID)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}

	/*
		Creating a Bitcoin transaction one step at a time:
			1. Create the transaction message
			2. Get the output of the deposit transaction and convert it into the transaction input
			3. Create a new output
		See https://blog.hlongvu.com/post/t0xx5dejn3-Understanding-btcd-Part-4-Create-and-Sign-a-Bitcoin-transaction-with-btcd
	*/

	tx := wire.NewMsgTx(wire.TxVersion)

	// The signature script will be set later and we have no witness
	txIn := wire.NewTxIn(out.OutPoint, nil, nil)
	tx.AddTxIn(txIn)

	var recipient btcutil.Address
	var err error
	if recipientAddr != "" {
		addr, err := types.ParseBtcAddress(recipientAddr, rpc.Network())
		if err != nil {
			return nil, err
		}
		recipient, err = addr.Convert()
		if err != nil {
			return nil, err
		}
	} else {
		pk, ok := s.GetNextMasterKey(ctx, balance.Bitcoin)
		if !ok {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, "next master key not set")
		}
		recipient, err = types.PKHashFromKey(pk, rpc.Network())
	}
	addrScript, err := txscript.PayToAddrScript(recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}

	sat, err := denom.ParseSatoshi(amountStr)
	if err != nil {
		return nil, err
	}
	txOut := wire.NewTxOut(sat.Amount.Int64(), addrScript)
	tx.AddTxOut(txOut)

	return types.ModuleCdc.MustMarshalJSON(tx), nil
}

package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	QueryTxInfo = "txInfo"
)

func NewQuerier(_ Keeper, rpc types.RPCClient) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryTxInfo:
			return queryTxInfo(rpc, path[1], path[2])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}
	}
}

func queryTxInfo(rpc types.RPCClient, txID string, voutIdx string) ([]byte, error) {
	v, err := types.ParseVoutIdx(voutIdx)
	if err != nil {
		return nil, err
	}
	hash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return nil, err
	}

	info, err := rpc.GetOutPointInfo(wire.NewOutPoint(hash, v))
	if err != nil {
		return nil, err
	}

	return types.ModuleCdc.MustMarshalJSON(info), nil
}

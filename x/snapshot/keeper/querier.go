package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// Query labels
const (
	QProxy    = "proxy"
	QOperator = "operator"
)

// NewQuerier returns a new querier for the evm module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QProxy:
			return queryProxy(ctx, k, path[1])
		case QOperator:
			return queryOperator(ctx, k, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown snapshot query endpoint: %s", path[0]))
		}
	}
}

func queryProxy(ctx sdk.Context, k Keeper, address string) ([]byte, error) {
	addr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "address invalid")
	}

	proxy, active := k.GetProxy(ctx, addr)
	if proxy == nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no proxy set for operator address")
	}

	statusStr := "inactive"
	if active {
		statusStr = "active"
	}

	reply := struct {
		Address string `json:"address"`
		Status  string `json:"status"`
	}{
		Address: proxy.String(),
		Status:  statusStr,
	}

	bz, err := json.Marshal(reply)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
	}

	return bz, nil
}

func queryOperator(ctx sdk.Context, k Keeper, proxy string) ([]byte, error) {
	proxyAddr, err := sdk.AccAddressFromBech32(proxy)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "invalid proxy address")
	}

	operator := k.GetOperator(ctx, proxyAddr)
	if operator == nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, "no operator associated to the proxy address")
	}

	return []byte(operator.String()), nil
}

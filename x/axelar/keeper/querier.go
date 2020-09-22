package keeper

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryTrackedAddress = "address"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryTrackedAddress:
			return queryAddress(ctx, path[1], k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown axelar query endpoint: %s", path[0]))
		}
	}
}

func queryAddress(ctx sdk.Context, addr string, k Keeper) ([]byte, error) {
	address := k.GetTrackedAddress(ctx, addr)
	if !address.IsValid() {
		return nil, types.ErrAddressNotTracked
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

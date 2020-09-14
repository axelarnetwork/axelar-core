package keeper

import (
	"fmt"
	"github.com/axelarnetwork/axelar-net/x/axelar/internal/types"
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
			return queryAddress(ctx, req, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown axelar query endpoint: %s", path[0]))
		}
	}
}

func queryAddress(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	address := k.GetTrackedAddress(ctx, req.Data)
	if !address.IsValid() {
		return nil, types.ErrAddressNotTracked
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

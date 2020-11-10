package keeper

import (
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryGetKey = "get-key"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryGetKey:
			return queryGetKey(ctx, path[1], k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown axelar query endpoint: %s", path[0]))
		}
	}
}

func queryGetKey(ctx sdk.Context, keyID string, k Keeper) ([]byte, error) {
	pk, err := k.GetKey(ctx, keyID)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "GetKey error for key [%s]", keyID)
	}

	// pk is of type ecdsa.PublicKey, which is inherently un-marshalable
	// convert pk to tss-libs crypto.ECPoint, which implements json.Marshaler
	pkMarshalable := convert.PubkeyToPoint(pk)
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, pkMarshalable)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

package keeper

import (
	"bytes"
	"encoding/gob"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

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
	pubkey, err := k.GetKey(ctx, keyID)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "GetKey error for key [%s]", keyID)
	}

	var bz bytes.Buffer
	if err := gob.NewEncoder(&bz).Encode(pubkey); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to re-serialize pubkey for query [%s]", keyID)
	}

	return bz.Bytes(), nil
}

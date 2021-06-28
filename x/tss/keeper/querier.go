package keeper

import (
	"fmt"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Query paths
const (
	QueryCurrentKey = "current-key"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k tssTypes.TSSKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case Query:
			res, err = queryCurrentKey(ctx, k, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryCurrentKey(ctx sdk.Context, k tssTypes.TSSKeeper, data []byte) ([]byte, error) {
	var params QueryKeyParams
	err := types.ModuleCdc.LegacyAmino.UnmarshalJSON(data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
	}

	key, ok := k.GetCurrentKey(ctx, chain, role)
	if !ok {
		return nil, fmt.Errorf("masterKey not found")
	}

	return []byte(addr.Address), nil
}

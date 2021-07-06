package keeper

import (
	"fmt"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// Query paths
const (
	QueryGetSig = "get-sig"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k tssTypes.TSSKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryGetSig:
			res, err = queryGetSig(ctx, k, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func queryGetSig(ctx sdk.Context, k tssTypes.TSSKeeper, data []byte) ([]byte, error) {
	sigID := string(data)

	var sigResponse tssTypes.QuerySigResponse
	sig, ok := k.GetSig(ctx, sigID)
	if !ok {
		return sigResponse.Marshal()
	}

	sigResponse = tssTypes.QuerySigResponse{
		R: sig.R.Bytes(),
		S: sig.S.Bytes(),
	}

	return sigResponse.Marshal()
}

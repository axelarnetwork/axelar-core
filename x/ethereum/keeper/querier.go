package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"

	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	QueryMasterKey = "masterkey"
)

func NewQuerier(k Keeper, s ethTypes.Signer) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {

		case QueryMasterKey:

			return queryMasterAddress(ctx, k, s)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown eth-bridge query endpoint: %s", path[0]))
		}
	}
}

func queryMasterAddress(ctx sdk.Context, k Keeper, s ethTypes.Signer) ([]byte, error) {

	pk, ok := s.GetCurrentMasterKey(ctx, "ethereum")
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	fromAddress := crypto.PubkeyToAddress(pk)

	bz := fromAddress.Bytes()

	return bz, nil
}

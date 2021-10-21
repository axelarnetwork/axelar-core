package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// Query labels
const (
	QueryChainMaintainers = "chain-maintainers"
)

// NewQuerier returns a new querier for the nexus module
func NewQuerier(k types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryChainMaintainers:
			return QueryChainMaintainersByChain(ctx, k, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown nexus query endpoint: %s", path[0]))
		}
	}
}

// QueryChainMaintainersByChain returns the maintainers for the given chain
func QueryChainMaintainersByChain(ctx sdk.Context, k types.Nexus, chainStr string) ([]byte, error) {
	var resp types.QueryChainMaintainersResponse

	chain, ok := k.GetChain(ctx, chainStr)
	if !ok {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("%s is not a registered chain", chainStr))
	}

	resp.Maintainers = k.GetChainMaintainers(ctx, chain)

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

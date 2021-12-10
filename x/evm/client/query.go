package client

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// QueryChains returns all EVM chains
func QueryChains(clientCtx client.Context) (types.QueryChainsResponse, error) {
	path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QChains)
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return types.QueryChainsResponse{}, sdkerrors.Wrap(err, "could not get EVM chains")
	}

	var res types.QueryChainsResponse
	err = res.Unmarshal(bz)
	if err != nil {
		return types.QueryChainsResponse{}, sdkerrors.Wrap(err, "could not get EVM chains")
	}
	return res, nil
}

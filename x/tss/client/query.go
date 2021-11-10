package client

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// QueryNextKeyID returns a response that contains the next assigned key ID for the given chain and role
func QueryNextKeyID(clientCtx client.Context, chain string, roleStr string) (types.QueryNextKeyIDResponse, error) {
	request, err := types.NewQueryNextKeyIDRequest(chain, roleStr)
	if err != nil {
		return types.QueryNextKeyIDResponse{}, err
	}
	path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryNextKeyID)
	bz, err := request.Marshal()
	if err != nil {
		return types.QueryNextKeyIDResponse{}, err
	}
	bz, _, err = clientCtx.QueryWithData(path, bz)
	if err != nil {
		return types.QueryNextKeyIDResponse{}, sdkerrors.Wrap(err, "could not get the next key ID")
	}

	var res types.QueryNextKeyIDResponse
	err = res.Unmarshal(bz)
	if err != nil {
		return types.QueryNextKeyIDResponse{}, sdkerrors.Wrap(err, "could not get the next key ID")
	}
	return res, nil
}

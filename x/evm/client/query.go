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

// QueryPendingCommands returns all pending commands for the given chain
func QueryPendingCommands(clientCtx client.Context, chain string) (types.QueryPendingCommandsResponse, error) {
	path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QPendingCommands, chain)
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return types.QueryPendingCommandsResponse{}, sdkerrors.Wrapf(err, "could not get the pending commands for chain %s", chain)
	}

	var res types.QueryPendingCommandsResponse
	err = res.Unmarshal(bz)
	if err != nil {
		return types.QueryPendingCommandsResponse{}, sdkerrors.Wrap(err, "could not get pending commands")
	}
	return res, nil
}

// QueryCommand returns the specified command for the given chain
func QueryCommand(clientCtx client.Context, chain, id string) (types.QueryCommandResponse, error) {
	path := fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QCommand, chain, id)
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return types.QueryCommandResponse{}, sdkerrors.Wrapf(err, "could not get command for chain %s", chain)
	}

	var res types.QueryCommandResponse
	err = res.Unmarshal(bz)

	if err != nil {
		return types.QueryCommandResponse{}, sdkerrors.Wrap(err, "could not get pending commands")
	}
	return res, nil
}

// QueryBurnerInfo returns the specified command for the given chain
func QueryBurnerInfo(clientCtx client.Context, chain, address string) (types.QueryBurnerInfoResponse, error) {
	path := fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QBurnerInfo, chain, address)
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return types.QueryBurnerInfoResponse{}, sdkerrors.Wrapf(err, "could not get address for chain %s", chain)
	}

	var res types.QueryBurnerInfoResponse
	err = res.Unmarshal(bz)

	if err != nil {
		return types.QueryBurnerInfoResponse{}, sdkerrors.Wrap(err, "could not get deposit address")
	}
	return res, nil
}

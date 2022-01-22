package client

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"

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

// QueryBurnerInfo returns the burner information for the given address
func QueryBurnerInfo(clientCtx client.Context, chain string, burnerAddress common.Address) (types.BurnerInfo, error) {
	path := fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QBurnerInfo, chain, burnerAddress.Hex())
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return types.BurnerInfo{}, sdkerrors.Wrapf(err, "could not get burner info for chain %s", chain)
	}

	if len(bz) == 0 {
		return types.BurnerInfo{}, fmt.Errorf("burner info not found")
	}

	var res types.BurnerInfo
	err = res.Unmarshal(bz)

	if err != nil {
		return types.BurnerInfo{}, sdkerrors.Wrap(err, "could not get burner info")
	}
	return res, nil
}

// QueryBurnerExists returns whether or not the given address is a burner address
func QueryBurnerExists(clientCtx client.Context, chain string, burnerAddress common.Address) (bool, error) {
	path := fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QBurnerInfo, chain, burnerAddress.Hex())
	bz, _, err := clientCtx.Query(path)
	if err != nil {
		return false, sdkerrors.Wrapf(err, "could not get burner info for chain %s", chain)
	}

	if len(bz) == 0 {
		return false, nil
	}

	var res types.BurnerInfo
	err = res.Unmarshal(bz)
	if err != nil {
		return false, sdkerrors.Wrap(err, "could not get burner info")
	}

	return true, nil
}

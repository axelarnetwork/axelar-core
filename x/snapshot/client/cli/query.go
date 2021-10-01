package cli

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	evmQueryCmd := &cobra.Command{
		Use:                        "snapshot",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmQueryCmd.AddCommand(
		GetCmdGetProxy(queryRoute),
		GetCmdGetOperator(queryRoute),
		GetCmdGetSnapshot(queryRoute),
		GetCmdGetValidators(queryRoute),
	)

	return evmQueryCmd

}

// GetCmdGetProxy returns the proxy address associated to some operator address
func GetCmdGetProxy(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy [operator address]",
		Short: "Fetch the proxy address associated with [operator address] and status (active/inactive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QProxy, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFProxyAddress)
			}

			fmt.Println(string(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetOperator returns the operator address associated to some proxy address
func GetCmdGetOperator(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator [proxy address]",
		Short: "Fetch the operator address associated with [proxy address]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QOperator, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFOperatorAddress)
			}

			fmt.Println(string(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetSnapshot returns the snapshot for a given counter
func GetCmdGetSnapshot(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [counter]",
		Short: "Fetch the snapshot for a given counter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QInfo, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFSnapshot)
			}

			fmt.Println(string(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetValidators returns the validators
func GetCmdGetValidators(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Fetch the validators and their information",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QValidators))
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrValidators)
			}

			var res types.QueryValidatorsResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return cliCtx.PrintProto(&res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

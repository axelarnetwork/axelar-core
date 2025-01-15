package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
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
		GetProxyByOperator(),
		GetOperatorByProxy(),
		GetParams(),
	)

	return evmQueryCmd

}

// GetProxyByOperator returns the proxy address associated to some operator address
func GetProxyByOperator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy [operator address]",
		Short: "Fetch the proxy address associated with [operator address] and status (active/inactive)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Proxy(cmd.Context(), &types.ProxyByOperatorRequest{OperatorAddress: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetOperatorByProxy returns the operator address associated to some proxy address
func GetOperatorByProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator [proxy address]",
		Short: "Fetch the operator address associated with [proxy address]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Operator(cmd.Context(), &types.OperatorByProxyRequest{ProxyAddress: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetParams returns the snapshot params
func GetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Returns the params for the snapshot module",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.ParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

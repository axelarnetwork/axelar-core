package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "axelarnet",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCmdPendingIBCTransfersCount(),
		getParams(),
		getCmdIBCPath(),
		getCmdChainByIBCPath(),
	)

	return queryCmd

}

// GetCmdPendingIBCTransfersCount returns the command for a pending ibc transfer query
func GetCmdPendingIBCTransfersCount() *cobra.Command {
	return &cobra.Command{
		Use:   "ibc-transfer-count",
		Short: "returns the number of pending IBC transfers per chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.PendingIBCTransferCount(cmd.Context(), &types.PendingIBCTransferCountRequest{})
			if err != nil {
				return errors.Wrap(err, "failed to query pending IBC transfers")
			}

			return clientCtx.PrintProto(res)
		},
	}
}

func getParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Returns the params for the axelarnet module",
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

func getCmdIBCPath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-path [chain]",
		Short: "Returns the registered IBC path for the given Cosmos chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.IBCPath(cmd.Context(), &types.IBCPathRequest{
				Chain: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getCmdChainByIBCPath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-by-ibc-path [ibc path]",
		Short: "Returns the Cosmos chain for the given IBC path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.ChainByIBCPath(cmd.Context(), &types.ChainByIBCPathRequest{
				IbcPath: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

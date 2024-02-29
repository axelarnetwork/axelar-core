package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexusexported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
)

const (
	activated   = "activated"
	deactivated = "deactivated"
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
		getCmdChains(),
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

// getCmdChains returns the query to get all Cosmos chains
func getCmdChains() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chains",
		Short: "Return the supported Cosmos chains by status",
		Args:  cobra.ExactArgs(0),
	}

	status := cmd.Flags().String("status", "", fmt.Sprintf("the chain status [%s|%s]", activated, deactivated))

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		nexusQueryClient := nexustypes.NewQueryServiceClient(clientCtx)

		var chainStatus nexusexported.ChainStatus
		switch *status {
		case "":
			chainStatus = nexusexported.StatusUnspecified
		case activated:
			chainStatus = nexusexported.Activated
		case deactivated:
			chainStatus = nexusexported.Deactivated
		default:
			return fmt.Errorf("unrecognized chain status %s", *status)
		}

		res, err := nexusQueryClient.Chains(cmd.Context(), &nexustypes.ChainsRequest{
			Status: chainStatus,
			Module: exported.ModuleName,
		})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

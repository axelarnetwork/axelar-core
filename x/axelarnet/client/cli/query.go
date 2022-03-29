package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
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

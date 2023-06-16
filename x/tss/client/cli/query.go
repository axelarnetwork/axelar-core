package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	tssQueryCmd := &cobra.Command{
		Use:                        "tss",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	tssQueryCmd.AddCommand(
		GetParams(),
	)

	return tssQueryCmd

}

// GetParams returns the tss params
func GetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Returns the params for the tss module",
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

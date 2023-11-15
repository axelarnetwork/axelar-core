package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	rewardQueryCmd := &cobra.Command{
		Use:                        "reward",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	rewardQueryCmd.AddCommand(
		GetCmdInflationRate(),
		GetParams(),
	)

	return rewardQueryCmd
}

// GetCmdInflationRate returns the inflation on the network
func GetCmdInflationRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-rate",
		Short: "Returns the inflation rate on the network. If a validator is provided, query the inflation rate for that validator.",
		Args:  cobra.ExactArgs(0),
	}

	validator := cmd.Flags().String("validator", "", "the validator to retrieve the inflation rate for")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		if _, err := sdk.ValAddressFromBech32(*validator); *validator != "" && err != nil {
			return sdkerrors.Wrap(err, "invalid validator address")
		}

		queryClient := types.NewQueryServiceClient(clientCtx)
		res, err := queryClient.InflationRate(cmd.Context(), &types.InflationRateRequest{
			Validator: *validator,
		})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetParams returns the reward params
func GetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Returns the params for the reward module",
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

package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	evmQueryCmd := &cobra.Command{
		Use:                        "tss",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmQueryCmd.AddCommand(
		GetCmdGetSig(queryRoute),
	)

	return evmQueryCmd

}

// GetCmdGetSig returns the query for a signature by its sigID
func GetCmdGetSig(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signature [sig ID]",
		Short: "Query a signature by sig ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			sigID := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QuerySigStatus, sigID), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			var sigResponse types.QuerySigResponse
			err = sigResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			hexSig := types.NewHexSignatureFromQuerySigResponse(&sigResponse)
			return cliCtx.PrintObjectLegacy(hexSig)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

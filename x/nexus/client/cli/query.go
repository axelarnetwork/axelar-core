package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCommandChainMaintainers(queryRoute),
	)

	return queryCmd
}

// GetCommandChainMaintainers returns the query for getting chain maintainers for the given chain
func GetCommandChainMaintainers(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-maintainers [chain]",
		Short: "Returns the chain maintainers for the given chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryChainMaintainers, args[0]))
			if err != nil {
				return sdkerrors.Wrap(err, "couldn't resolve chain maintainers")
			}

			var res types.QueryChainMaintainersResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

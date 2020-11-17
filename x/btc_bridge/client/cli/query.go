package cli

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	btcQueryCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	btcQueryCmd.AddCommand(flags.GetCommands(GetCmdTrackedAddress(queryRoute, cdc))...)

	return btcQueryCmd

}

func GetCmdTrackedAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "address [addressString]",
		Short: "Query info about a tracked address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			address := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryTrackedAddress, address), nil)
			if err != nil {
				fmt.Printf("could not resolve address %s \n%s\n", address, err.Error())

				return nil
			}

			var out exported.ExternalChainAddress
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

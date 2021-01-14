package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	tssQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	tssQueryCmd.AddCommand(flags.GetCommands(GetCmdMasterAddress(queryRoute, cdc))...)

	return tssQueryCmd

}

func GetCmdMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-masteraddress [chain] [arg 1]...[arg n]",
		Short: "Query master address by chain.",
		Long:  "Query master address by chain. Each chain may require its own specific arguments.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			chain := args[0]
			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryMasterKey, chain)

			for i := 1; i < len(args); i++ {

				path = path + "/" + args[i]
			}

			res, _, err := cliCtx.QueryWithData(path, nil)
			if err != nil {
				fmt.Printf("could not resolve master key: %s\n", err.Error())

				return nil
			}

			return cliCtx.PrintOutput(string(res))
		},
	}

	return cmd
}

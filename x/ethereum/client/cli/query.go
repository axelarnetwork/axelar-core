package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ethQueryCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ethQueryCmd.AddCommand(flags.GetCommands(GetCmdMasterAddress(queryRoute, cdc))...)

	return ethQueryCmd

}

func GetCmdMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Query an address by key ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterKey), nil)
			if err != nil {
				fmt.Printf("could not resolve master key: %s\n", err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintOutput(out.Hex())
		},
	}

	return cmd
}

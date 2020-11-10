package cli

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/binance-chain/tss-lib/crypto"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
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

	tssQueryCmd.AddCommand(flags.GetCommands(GetCmdGetKey(queryRoute, cdc))...)

	return tssQueryCmd

}

func GetCmdGetKey(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-key [id]", // TODO should this use keeper.QueryGetKey constant?
		Short: "Get a threshold pubkey",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			keyID := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryGetKey, keyID), nil)
			if err != nil {
				return err
			}

			var out crypto.ECPoint
			cdc.MustUnmarshalJSON(res, &out)

			// crypto.ECPoint supports only json marshalling
			if cliCtx.OutputFormat != "json" {
				fmt.Printf("warning: output format [%s] not supported. use '-o json'", cliCtx.OutputFormat)
			}

			return cliCtx.PrintOutput(out)
		},
	}
}

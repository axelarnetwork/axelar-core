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

	masterAddressCmd := &cobra.Command{
		Use:                        "get-masteraddress",
		Short:                      "get master address subcommand",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	masterAddressCmd.AddCommand(flags.GetCommands(
		GetCmdBitcoinMasterAddress(queryRoute, cdc),
		GetCmdEthereumMasterAddress(queryRoute, cdc))...)

	tssQueryCmd.AddCommand(masterAddressCmd)

	return tssQueryCmd

}

func GetCmdBitcoinMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bitcoin [network]",
		Short: "Query bitcoin master key.",
		Long:  "Query bitcoin master key. Network should be `mainnet`, `testnet3`, or `regtest`",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			network := args[0]

			path := fmt.Sprintf("custom/%s/%s/bitcoin/%s", queryRoute, keeper.QueryMasterKey, network)

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

func GetCmdEthereumMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum",
		Short: "Query ethereum master key.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := fmt.Sprintf("custom/%s/%s/ethereum/", queryRoute, keeper.QueryMasterKey)

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

package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s query subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(flags.GetCommands(GetCmdTxInfo(queryRoute, cdc), GetCmdRawTx(queryRoute, cdc))...)

	return btcTxCmd
}

// GetCmdTxInfo returns the tx info query command
func GetCmdTxInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "txInfo [txID]",
		Short: "Query the transaction info of a transaction with [txID] on Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txID := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryOutInfo, txID), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not resolve txID %s", txID)
			}

			// Ensure the output can be unmarshalled
			cdc.MustUnmarshalJSON(res, &types.OutPointInfo{})
			return cliCtx.PrintOutput(res)
		},
	}
}

// GetCmdRawTx returns the raw tx creation command
func GetCmdRawTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	var useMasterKey bool
	var recipient string
	rawTxCmd := &cobra.Command{
		Use:   "rawTx [txID] [amount] [-r <recipient> | -m]",
		Short: "Get a raw transaction that spends [amount] of the utxo of [txID] to <recipient> or the next master key in rotation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txID := args[0]
			amount := args[1]

			if (recipient == "" && !useMasterKey) || (recipient != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the recipient or to use the master key, not both\"")
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s/%s", queryRoute, keeper.QueryRawTx, txID, amount, recipient), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not resolve txID %s", txID)
			}

			// Ensure the output can be unmarshalled
			cdc.MustUnmarshalJSON(res, &types.OutPointInfo{})
			return cliCtx.PrintOutput(res)
		},
	}
	addRecipientFlag(rawTxCmd, &recipient)
	addMasterKeyFlag(rawTxCmd, &useMasterKey)
	return rawTxCmd
}

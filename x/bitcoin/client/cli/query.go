package cli

import (
	"fmt"
	"strings"

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

	btcTxCmd.AddCommand(flags.GetCommands(
		GetCmdDepositAddress(queryRoute, cdc),
		GetCmdTxInfo(queryRoute, cdc),
		GetCmdSendTransfers(queryRoute, cdc),
	)...)

	return btcTxCmd
}

// GetCmdDepositAddress returns the deposit address command
func GetCmdDepositAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-addr [chain] [recipient address]",
		Short: "Returns a bitcoin deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryDepositAddress)

			res, _, err := cliCtx.QueryWithData(path, cdc.MustMarshalJSON(types.DepositQueryParams{Chain: args[0], Address: args[1]}))
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFDepositAddress)
			}

			return cliCtx.PrintOutput(string(res))
		},
	}

	return cmd
}

// GetCmdTxInfo returns the tx info query command
func GetCmdTxInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "txInfo [blockHash] [txID:voutIdx]",
		Short: "Query the info of the outpoint at index [voutIdx] of transaction [txID] on Bitcoin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			out, err := types.OutPointFromStr(args[1])
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryOutInfo, args[0]), cdc.MustMarshalJSON(out))
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFTxInfo, out.Hash.String(), out.Index)
			}

			var info types.OutPointInfo
			cdc.MustUnmarshalJSON(res, &info)
			fmt.Println(strings.ReplaceAll(string(res), "\"", "\\\""))
			return cliCtx.PrintOutput(info)
		},
	}
}

// GetCmdSendTransfers sends a transaction containing all pending transfers to Bitcoin
func GetCmdSendTransfers(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send",
		Short: "Send a transaction to Bitcoin that consolidates deposits and withdrawals",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.SendTx), nil)
			if err != nil {
				return sdkerrors.Wrap(err, "could not send the consolidation transaction")
			}

			var out string
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

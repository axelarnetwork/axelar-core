package cli

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
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
		GetCmdConsolidationAddress(queryRoute, cdc),
		GetCmdTxInfo(queryRoute, cdc),
		GetCmdRawTx(queryRoute, cdc),
		GetCmdSendTx(queryRoute, cdc),
	)...)

	return btcTxCmd
}

func GetCmdDepositAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-addr [blockchain] [recipient addr]",
		Short: "Returns a bitcoin deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryDepositAddress)

			chain := exported.ChainFromString(args[0])
			if err := chain.Validate(); err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(path, cdc.MustMarshalJSON(exported.CrossChainAddress{Chain: chain, Address: args[1]}))
			if err != nil {
				return sdkerrors.Wrap(err, "could not resolve master key")
			}

			return cliCtx.PrintOutput(string(res))
		},
	}

	return cmd
}

func GetCmdConsolidationAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consolidation-addr [deposit addr]",
		Short: "Returns a new consolidation address for an old deposit address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryConsolidationAddress)

			chain := exported.ChainFromString(args[0])
			if err := chain.Validate(); err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(path, cdc.MustMarshalJSON(exported.CrossChainAddress{Chain: chain, Address: args[1]}))
			if err != nil {
				return sdkerrors.Wrap(err, "could not resolve master key")
			}

			return cliCtx.PrintOutput(string(res))
		},
	}

	return cmd
}

// GetCmdTxInfo returns the tx info query command
func GetCmdTxInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "txInfo [txID:voutIdx]",
		Short: "Query the info of the outpoint at index [voutIdx] of transaction [txID] on Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			out, err := types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryOutInfo), cdc.MustMarshalJSON(out))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not resolve txID %s and vout index %d", out.Hash.String(), out.Index)
			}

			var info types.OutPointInfo
			cdc.MustUnmarshalJSON(res, &info)
			fmt.Println(strings.ReplaceAll(string(res), "\"", "\\\""))
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdRawTx returns the raw tx creation command
func GetCmdRawTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawTx [txID:voutIdx] [amount] [recipient]",
		Short: "Get a raw transaction that spends [amount] of the outpoint [voutIdx] of [txID] to <recipient> or the next master key in rotation",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			amount, err := denom.ParseSatoshi(args[2])
			if err != nil {
				return err
			}

			out, err := types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}

			params := types.RawTxParams{
				DepositAddr: args[3],
				OutPoint:    out,
				Satoshi:     amount,
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryRawTx), cdc.MustMarshalJSON(params))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not create a new transaction spending transaction %s", out.String())
			}

			var tx *wire.MsgTx
			cdc.MustUnmarshalJSON(res, &tx)
			fmt.Println(strings.ReplaceAll(string(res), "\"", "\\\""))
			return cliCtx.PrintOutput(tx)
		},
	}
}

// GetCmdSendTx sends a transaction to Bitcoin
func GetCmdSendTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [txID:voutIdx]",
		Short: "Send a transaction to Bitcoin that spends output [voutIdx] of tx [txID]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			outpoint, err := types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.SendTx), cdc.MustMarshalJSON(outpoint))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send the transaction spending transaction %s", args[0])
			}

			var out string
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

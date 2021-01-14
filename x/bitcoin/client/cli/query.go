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
		GetCmdMasterAddress(queryRoute, cdc),
		GetCmdTxInfo(queryRoute, cdc),
		GetCmdRawTx(queryRoute, cdc),
		GetCmdSendTx(queryRoute, cdc),
	)...)

	return btcTxCmd
}

func GetCmdMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Query bitcoin master key.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterAddress)

			res, _, err := cliCtx.QueryWithData(path, nil)
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
		Use:   "txInfo [txID] [voutIdx]",
		Short: "Query the info of the outpoint at index [voutIdx] of transaction [txID] on Bitcoin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txID := args[0]
			voutIdx := args[1]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QueryOutInfo, txID, voutIdx), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not resolve txID %s and vout %s", txID, voutIdx)
			}

			var out types.OutPointInfo
			cdc.MustUnmarshalJSON(res, &out)
			fmt.Println(strings.ReplaceAll(string(res), "\"", "\\\""))
			return cliCtx.PrintOutput(out)
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

			if (recipient == "" && !useMasterKey) || (recipient != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the recipient or to use the master key, not both\"")
			}

			amount, err := denom.ParseSatoshi(args[1])
			if err != nil {
				return err
			}

			params := types.RawParams{
				Recipient: recipient,
				TxID:      args[0],
				Satoshi:   amount,
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryRawTx), cdc.MustMarshalJSON(params))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not create a new transaction spending transaction %s", params.TxID)
			}

			var tx *wire.MsgTx
			cdc.MustUnmarshalJSON(res, &tx)
			fmt.Println(strings.ReplaceAll(string(res), "\"", "\\\""))
			return cliCtx.PrintOutput(tx)
		},
	}
	addRecipientFlag(rawTxCmd, &recipient)
	addMasterKeyFlag(rawTxCmd, &useMasterKey)
	return rawTxCmd
}

// GetCmdSendTx sends a transaction to Bitcoin
func GetCmdSendTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [txID]",
		Short: "Send a transaction that spends tx [txID] to Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.SendTx, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send the transaction spending transaction %s", args[0])
			}

			var out string
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

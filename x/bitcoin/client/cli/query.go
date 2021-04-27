package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s query subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(
		GetCmdDepositAddress(queryRoute),
		GetCmdConsolidationTx(queryRoute),
	)

	return btcTxCmd
}

// GetCmdDepositAddress returns the deposit address command
func GetCmdDepositAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-addr [chain] [recipient address]",
		Short: "Returns a bitcoin deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryDepositAddress)

			res, _, err := clientCtx.QueryWithData(path, types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{Chain: args[0], Address: args[1]}))
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFDepositAddress)
			}

			return clientCtx.PrintObjectLegacy(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdMasterAddress returns the master address command
func GetCmdMasterAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-addr",
		Short: "Returns the bitcoin address of the current master key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterAddress)

			res, _, err := clientCtx.QueryWithData(path, nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFDepositAddress)
			}

			return clientCtx.PrintString(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdConsolidationTx returns a transaction containing all pending transfers to Bitcoin
func GetCmdConsolidationTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rawTx",
		Short: "Returns the encoded hex string of a fully signed transfer and consolidation transaction",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.GetTx), nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFGetTransfers)
			}

			return clientCtx.PrintObjectLegacy(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"encoding/binary"
	"fmt"
	"strconv"

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
		GetCmdConsolidationTxState(queryRoute),
		GetCmdConsolidationTx(queryRoute),
		GetCmdPayForConsolidationTx(queryRoute),
		GetCmdMasterAddress(queryRoute),
		GetCmdNextMasterKeyID(queryRoute),
		GetCmdMinimumWithdrawAmount(queryRoute),
		GetCmdTxState(queryRoute),
	)

	return btcTxCmd
}

// GetCmdDepositAddress returns the deposit address command
func GetCmdDepositAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-address [chain] [recipient address]",
		Short: "Returns a bitcoin deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QDepositAddress)

			res, _, err := clientCtx.QueryWithData(path, types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{Chain: args[0], Address: args[1]}))
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFMasterKey)
			}

			return clientCtx.PrintObjectLegacy(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdMasterAddress returns the master address command
func GetCmdMasterAddress(queryRoute string) *cobra.Command {
	var IncludeKeyID bool
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Returns the bitcoin address of the current master key, and optionally the key's ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QMasterAddress)

			res, _, err := clientCtx.QueryWithData(path, nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFMasterKey)
			}

			var resp types.QueryMasterAddressResponse
			err = resp.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFMasterKey)
			}

			if IncludeKeyID {
				return clientCtx.PrintObjectLegacy(resp)
			}

			return clientCtx.PrintObjectLegacy(resp.Address)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().BoolVar(&IncludeKeyID, "include-key-id", false, "include the current master key ID in the output")
	return cmd
}

// GetCmdNextMasterKeyID returns the the assigned master key ID
func GetCmdNextMasterKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nextMasterKeyID",
		Short: "Returns the next assigned master key ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QNextMasterKeyID)

			res, _, err := clientCtx.QueryWithData(path, nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFNextMasterKey)
			}

			return clientCtx.PrintString(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdConsolidationTxState returns the state of the bitcoin consolidation transaction
func GetCmdConsolidationTxState(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consolidationTxState",
		Short: "Returns the state of the consolidation transaction as seen by Axelar network",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QConsolidationTxState)

			res, _, err := clientCtx.QueryWithData(path, nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFConsolidationState)
			}

			return clientCtx.PrintObjectLegacy(string(res))
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

			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QConsolidationTx), nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFGetRawTx)
			}

			var response types.QueryRawTxResponse
			err = response.Unmarshal(res)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(&response)

		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdPayForConsolidationTx returns a transaction that pays for the consolidation transaction
func GetCmdPayForConsolidationTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rawPayForConsolidationTx",
		Short: "Returns the encoded hex string of a fully signed transaction that pays for the consolidation transaction",
		Args:  cobra.ExactArgs(0),
	}

	feeRate := cmd.Flags().Int64("fee-rate", 0, "fee rate to be set for the child-pay-for-parent transaction")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		bz := make([]byte, 8)
		binary.LittleEndian.PutUint64(bz, uint64(*feeRate))

		res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QPayForConsolidationTx), bz)
		if err != nil {
			return sdkerrors.Wrap(err, types.ErrFGetPayForRawTx)
		}

		return clientCtx.PrintObjectLegacy(string(res))
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdMinimumWithdrawAmount returns the minimum amount to withdraw
func GetCmdMinimumWithdrawAmount(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minWithdraw",
		Short: "Returns the minimum withdraw amount in satoshi",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QMinimumWithdrawAmount)

			res, _, err := clientCtx.QueryWithData(path, nil)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFMinWithdraw)
			}

			response := int64(binary.LittleEndian.Uint64(res))

			return clientCtx.PrintString(strconv.FormatInt(response, 10))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdTxState returns the state of the bitcoin transaction
func GetCmdTxState(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txState [txID:voutIdx]",
		Short: "Returns the state of a bitcoin transaction as seen by Axelar network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			outpointBytes := []byte(args[0])
			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QTxState)

			res, _, err := clientCtx.QueryWithData(path, outpointBytes)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFTxState)
			}

			return clientCtx.PrintObjectLegacy(string(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

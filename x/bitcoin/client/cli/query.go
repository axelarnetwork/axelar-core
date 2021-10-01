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
	cmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s query subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdDepositAddress(queryRoute),
		GetCmdDepositStatus(queryRoute),
		GetCmdConsolidationAddress(queryRoute),
		GetCmdNextKeyID(queryRoute),
		GetCmdMinOutputAmount(queryRoute),
		GetCmdLatestTx(queryRoute),
		GetCmdSignedTx(queryRoute),
	)

	return cmd
}

// GetCmdDepositAddress returns a bitcoin deposit address for a recipient address on another blockchain
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
			params := types.DepositQueryParams{Chain: args[0], Address: args[1]}

			bz, _, err := clientCtx.QueryWithData(path, types.ModuleCdc.MustMarshalLengthPrefixed(&params))
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrDepositAddr)
			}

			var res types.QueryAddressResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdDepositStatus returns the status of a bitcoin deposit given the outpoint
func GetCmdDepositStatus(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-status [txID:voutIdx]",
		Short: "Returns the status of the bitcoin deposit with the given outpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			_, err = types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}
			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QDepositStatus, args[0])
			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrDepositStatus)
			}

			var res types.QueryDepositStatusResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdConsolidationAddress returns the consolidation address
func GetCmdConsolidationAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consolidation-address",
		Short: "Returns the bitcoin consolidation address",
		Args:  cobra.ExactArgs(0),
	}
	keyRole := cmd.Flags().String("key-role", "", "the role of the key to get the consolidation address for")
	keyID := cmd.Flags().String("key-id", "", "the ID of the key to get the consolidation address for")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		var query string
		var param string
		switch {
		case *keyRole != "" && *keyID == "":
			query = keeper.QConsolidationAddressByKeyRole
			param = *keyRole
		case *keyRole == "" && *keyID != "":
			query = keeper.QConsolidationAddressByKeyID
			param = *keyID
		default:
			return fmt.Errorf("one and only one of the two flags key-role and key-id has to be set")
		}

		path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, query, param)

		bz, _, err := clientCtx.Query(path)
		if err != nil {
			return sdkerrors.Wrap(err, types.ErrConsolidationAddr)
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		return clientCtx.PrintProto(&res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdNextKeyID returns the ID of the next assigned key
func GetCmdNextKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-key-id [keyRole]",
		Short: "Returns the ID of the next assigned key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QNextKeyID, args[0])

			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrNextKeyID)
			}

			keyID := string(bz)

			return clientCtx.PrintString(keyID)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdMinOutputAmount returns the minimum amount allowed for any transaction output in satoshi
func GetCmdMinOutputAmount(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-output-amount",
		Short: "Returns the minimum amount allowed for any transaction output in satoshi",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QMinOutputAmount)

			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrMinOutputAmount)
			}

			minOutputAmount := int64(binary.LittleEndian.Uint64(bz))

			return clientCtx.PrintString(strconv.FormatInt(minOutputAmount, 10))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdLatestTx returns the latest consolidation transaction of the given key role
func GetCmdLatestTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-tx [keyRole]",
		Short: "Returns the latest consolidation transaction of the given key role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QLatestTxByTxType, args[0])

			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrLatestTx)
			}

			var res types.QueryTxResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignedTx returns the signed consolidation transaction of the given transaction hash
func GetCmdSignedTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signed-tx [txHash]",
		Short: "Returns the signed consolidation transaction of the given transaction hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QSignedTx, args[0])

			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrSignedTx)
			}

			var res types.QueryTxResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

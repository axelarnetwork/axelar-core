package cli

import (
	"fmt"
	"math/rand"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(
		GetCmdConfirmTxOut(),
		GetCmdLink(),
		GetCmdSignPendingTransfersTx(),
	)

	return btcTxCmd
}

// GetCmdConfirmTxOut returns the transaction confirmation command
func GetCmdConfirmTxOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirmTxOut [txID:voutIdx] [amount] [address]",
		Short: "Confirm a Bitcoin transaction",
		Long:  "Confirm that a transaction happened on the Bitcoin network so it can be processed on axelar.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			outPoint, err := types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}

			satoshi, err := types.ParseSatoshi(args[1])
			if err != nil {
				return err
			}

			outInfo := types.NewOutPointInfo(outPoint, btcutil.Amount(satoshi.Amount.Int64()), args[2])

			msg := types.NewConfirmOutpointRequest(clientCtx.GetFromAddress(), outInfo)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdLink links a cross chain address to a bitcoin address created by Axelar
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [chain] [address]",
		Short: "Link a cross chain address to a bitcoin address created by Axelar",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewLinkRequest(clientCtx.GetFromAddress(), args[1], args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignPendingTransfersTx returns the cli command to sign all pending token transfers from other chains
func GetCmdSignPendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-pending-transfers",
		Short: "Create a Bitcoin transaction for all pending transfers and sign it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewSignPendingTransfersRequest(clientCtx.GetFromAddress(), btcutil.Amount(rand.Int63()))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	axelarTxCmd := &cobra.Command{
		Use:                        "axelarnet",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	axelarTxCmd.AddCommand(
		GetCmdLink(),
		GetCmdConfirmDeposit(),
		GetCmdExecutePendingTransfersTx(),
	)

	return axelarTxCmd
}

// GetCmdLink links a cross chain address to an Axelar chain address
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [recipient chain] [recipient address] [symbol]",
		Short: "Link a cross chain address to an Axelar address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewLinkRequest(clientCtx.GetFromAddress(), args[0], args[1], args[2])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmDeposit returns the cli command to confirm a deposit
func GetCmdConfirmDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-deposit [chain] [txID] [amount] [burnerAddr]",
		Short: "Confirm a deposit to Axelar chain that sent given amount of token to a burner address",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID, err := hex.DecodeString(args[1])

			coin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			burnerAddr, err := sdk.AccAddressFromBech32(args[3])
			if err != nil {
				return err
			}

			msg := types.NewConfirmDepositRequest(cliCtx.GetFromAddress(), chain, txID, coin, burnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdExecutePendingTransfersTx returns the cli command to transfer all pending token transfers to Axelar chain
func GetCmdExecutePendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-pending-transfers",
		Short: "Send all pending transfers to Axelar chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewExecutePendingTransfersRequest(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

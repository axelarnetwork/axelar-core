package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
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
		GetCmdSignMasterConsolidationTx(),
		GetCmdRegisterExternalKey(),
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
		Use:   "sign-pending-transfers [keyID]",
		Short: "Create a Bitcoin transaction for all pending transfers and sign it",
		Args:  cobra.ExactArgs(1),
	}

	masterKeyAmountStr := cmd.Flags().String("master-key-amount", "0btc", "amount of satoshi to send to the master key")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		masterKeyAmount, err := types.ParseSatoshi(*masterKeyAmountStr)
		if err != nil {
			return err
		}

		msg := types.NewSignPendingTransfersRequest(clientCtx.GetFromAddress(), args[0], btcutil.Amount(masterKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignMasterConsolidationTx returns the cli command to sign the master key consolidation transaction
func GetCmdSignMasterConsolidationTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-master-consolidation [keyID]",
		Short: "Create a Bitcoin transaction for consolidating master key UTXOs, and send the change to an address controlled by [keyID]",
		Args:  cobra.ExactArgs(1),
	}

	secondaryKeyAmountStr := cmd.Flags().String("secondary-key-amount", "0btc", "amount of satoshi to send to the secondary key")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		secondaryKeyAmount, err := types.ParseSatoshi(*secondaryKeyAmountStr)
		if err != nil {
			return err
		}

		msg := types.NewSignMasterConsolidationTransactionRequest(clientCtx.GetFromAddress(), args[0], btcutil.Amount(secondaryKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterExternalKey returns the cli command to register an external key
func GetCmdRegisterExternalKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-external-key [keyID] [pubKeyHex]",
		Short: "Register the external key for bitcoin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]
			pubKeyHex := args[1]

			pubKeyBytes, err := hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
			if err != nil {
				return err
			}

			msg := types.NewRegisterExternalKeyRequest(clientCtx.GetFromAddress(), keyID, pubKey)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

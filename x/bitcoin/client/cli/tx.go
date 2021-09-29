package cli

import (
	"encoding/hex"
	"fmt"

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
		GetCmdCreatePendingTransfersTx(),
		GetCmdCreateMasterConsolidationTx(),
		GetCmdCreateRescueTx(),
		GetCmdSignTx(),
		GetCmdSubmitExternalSignature(),
	)

	return btcTxCmd
}

// GetCmdConfirmTxOut returns the transaction confirmation command
func GetCmdConfirmTxOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-tx-out [txID:voutIdx] [amount] [address]",
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

// GetCmdCreatePendingTransfersTx returns the cli command to create a secondary key consolidation transaction handling all pending transfers
func GetCmdCreatePendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pending-transfers-tx [keyID]",
		Short: "Create a Bitcoin transaction for all pending transfers",
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

		msg := types.NewCreatePendingTransfersTxRequest(clientCtx.GetFromAddress(), args[0], btcutil.Amount(masterKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateMasterConsolidationTx returns the cli command to create a master key consolidation transaction
func GetCmdCreateMasterConsolidationTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-master-tx [keyID]",
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

		msg := types.NewCreateMasterTxRequest(clientCtx.GetFromAddress(), args[0], btcutil.Amount(secondaryKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateRescueTx returns the cli command to create a rescue transaction
func GetCmdCreateRescueTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-rescue-tx",
		Short: "Create a Bitcoin transaction for rescuing the outpoints that were sent to old keys",
		Args:  cobra.ExactArgs(0),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		msg := types.NewCreateRescueTxRequest(clientCtx.GetFromAddress())
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignTx returns the cli command to sign a consolidation transaction
func GetCmdSignTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-tx [keyRole]",
		Short: "Sign a consolidation transaction with the current key of given key role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txType, err := types.TxTypeFromSimpleStr(args[0])
			if err != nil {
				return err
			}

			msg := types.NewSignTxRequest(clientCtx.FromAddress, txType)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSubmitExternalSignature returns the cli command to submit a signature from an external key
func GetCmdSubmitExternalSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-external-signature [keyID] [signatureHex] [sigHashHex]",
		Short: "Submit a signature of the given external key signing the given sig hash",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]

			signature, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			sigHash, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			msg := types.NewSubmitExternalSignatureRequest(clientCtx.GetFromAddress(), keyID, signature, sigHash)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

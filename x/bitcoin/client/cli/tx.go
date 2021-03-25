package cli

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(flags.PostCommands(
		GetCmdConfirmTxOut(cdc),
		GetCmdLink(cdc),
		GetCmdSignPendingTransfersTx(cdc),
	)...)

	return btcTxCmd
}

// GetCmdConfirmTxOut returns the transaction confirmation command
func GetCmdConfirmTxOut(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "confirmTxOut [txID:voutIdx] [amount] [address]",
		Short: "Confirm a Bitcoin transaction",
		Long: fmt.Sprintf(
			"Confirm that a transaction happened on the Bitcoin network so it can be processed on axelar."),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			outPoint, err := types.OutPointFromStr(args[0])
			if err != nil {
				return err
			}

			satoshi, err := types.ParseSatoshi(args[1])
			if err != nil {
				return err
			}

			outInfo, err := types.NewOutPointInfo(outPoint, btcutil.Amount(satoshi.Amount.Int64()), args[2])

			msg := types.MsgConfirmOutpoint{Sender: cliCtx.GetFromAddress(), OutPointInfo: outInfo}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdLink links a cross chain address to a bitcoin address created by Axelar
func GetCmdLink(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "link [chain] [address]",
		Short: "Link a cross chain address to a bitcoin address created by Axelar",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.MsgLink{Sender: cliCtx.GetFromAddress(), RecipientAddr: args[1], RecipientChain: args[0]}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSignPendingTransfersTx returns the cli command to sign all pending token transfers from Ethereum
func GetCmdSignPendingTransfersTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sign-pending-transfers [fee]",
		Short: "Create a Bitcoin transaction for all pending transfers and sign it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			satoshi, err := types.ParseSatoshi(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgSignPendingTransfers(cliCtx.GetFromAddress(), btcutil.Amount(satoshi.Amount.Int64()))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

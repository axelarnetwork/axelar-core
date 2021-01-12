package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	ethTxCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	rawTxCmd := makeCommand("raw")
	rawTxCmd.AddCommand(flags.PostCommands(GetCmdRawTx(cdc))...)

	verifyTxCmd := makeCommand("verify")
	verifyTxCmd.AddCommand(flags.PostCommands(GetCmdVerifyMintTx(cdc), GetCmdVerifyDeployTx(cdc))...)

	sendCmd := GetCmdSend(cdc)
	ethTxCmd.AddCommand(rawTxCmd, verifyTxCmd)
	ethTxCmd.AddCommand(flags.PostCommands(sendCmd)...)

	return ethTxCmd
}

func makeCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:                        name,
		Short:                      fmt.Sprintf("%s transactions subcommands", name),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
}

func GetCmdSend(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [txHash] [sigID]",
		Short: "Submit the specified transaction to ethereum with the specified signature",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgSendTx(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdRawTx(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw [tx json]",
		Short: "upload a raw (unsigned) Ethereum transaction to the acelar network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON([]byte(args[0]), &tx)

			msg := types.NewMsgRawTx(cliCtx.GetFromAddress(), tx)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdVerifyMintTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mint [tx json]",
		Short: "Verify an Ethereum transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON([]byte(args[0]), &tx)

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), tx)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVerifyDeployTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy [tx json] [contract ID] ",
		Short: "Verify an Ethereum transaction",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON([]byte(args[0]), &tx)

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), tx)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

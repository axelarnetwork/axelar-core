package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
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

	ethTxCmd.AddCommand(
		flags.PostCommands(
			GetCmdSignTx(cdc),
			GetCmdVerifyTx(cdc),
			GetCmdSignPendingTransfersTx(cdc),
		)...,
	)

	return ethTxCmd
}

func GetCmdSignTx(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [tx json file path]",
		Short: "sign a raw Ethereum transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			json, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON(json, &tx)

			msg := types.NewMsgSignTx(cliCtx.GetFromAddress(), json)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdVerifyTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verify [tx json file path]",
		Short: "Verify an Ethereum transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			json, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON(json, &tx)

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), json)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSignPendingTransfersTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sign-pending-transfers [chain]",
		Short: "Sign all pending transfers to Ethereum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)
			chainStr := args[0]

			chain := exported.ChainFromString(chainStr)
			if err := chain.Validate(); err != nil {
				return err
			}

			msg := types.NewMsgSignPendingTransfersTx(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

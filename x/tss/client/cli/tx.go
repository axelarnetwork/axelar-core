package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	cliUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	tssTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	tssTxCmd.AddCommand(flags.PostCommands(
		getCmdKeygenStart(cdc),
		getCmdMasterKeyAssignNext(cdc),
		getCmdRotateMasterKey(cdc),
		getCmdDeregister(cdc),
	)...)

	return tssTxCmd
}

func getCmdKeygenStart(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-keygen",
		Short: "Initiate threshold key generation protocol",
		Args:  cobra.NoArgs,
	}
	newKeyID := cmd.Flags().String("id", "", "unique ID for new key (required)")
	if cmd.MarkFlagRequired("id") != nil {
		panic("flag not set")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.NewMsgKeygenStart(cliCtx.FromAddress, *newKeyID, 0)
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

func getCmdMasterKeyAssignNext(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mk-assign-next [chain] [keyID]",
		Short: "Assigns a previously created key with [keyID] as the next master key for [chain]",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.MsgAssignNextMasterKey{
			Sender: cliCtx.FromAddress,
			Chain:  args[0],
			KeyID:  args[1],
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

func getCmdRotateMasterKey(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mk-rotate [chain]",
		Short: "Rotate the given chain from the old master key to the previously created one (see mk-refresh)",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.MsgRotateMasterKey{
			Sender: cliCtx.FromAddress,
			Chain:  args[0],
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

func getCmdDeregister(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister from participating in any future key generation",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgDeregister(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

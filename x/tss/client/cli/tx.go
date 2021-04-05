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
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
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
		getCmdAssignNextKey(cdc),
		getCmdRotateKey(cdc),
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

	subsetSize := cmd.Flags().Int64("subset-size", 0, "number of top validators to participate in the key generation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.NewMsgKeygenStart(cliCtx.FromAddress, *newKeyID, *subsetSize)
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}

	return cmd
}

func getCmdAssignNextKey(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-next [chain] [role] [keyID]",
		Short: "Assigns a previously created key with [keyID] as the next key for [chain]",
		Args:  cobra.ExactArgs(3),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		chain := args[0]
		keyRole, err := exported.KeyRoleFromStr(args[1])
		if err != nil {
			return err
		}
		keyID := args[2]

		msg := types.NewMsgAssignNextKey(cliCtx.FromAddress, chain, keyID, keyRole)

		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}

	return cmd
}

func getCmdRotateKey(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate [chain] [role]",
		Short: "Rotate the given chain from the old key to the previously assigned one",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		chain := args[0]
		keyRole, err := exported.KeyRoleFromStr(args[1])
		if err != nil {
			return err
		}

		msg := types.NewMsgRotateKey(cliCtx.FromAddress, chain, keyRole)
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

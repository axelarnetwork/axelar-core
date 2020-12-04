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
		getCmdSignStart(cdc),
		getCmdMasterKeyAssignNext(cdc),
		getCmdRotateMasterKey(cdc),
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
	threshold := cmd.Flags().IntP("threshold", "t", 2, "number of corruptions to withstand (required)")
	if cmd.MarkFlagRequired("threshold") != nil {
		panic("flag not set")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.MsgKeygenStart{
			Sender:    cliCtx.FromAddress,
			NewKeyID:  *newKeyID,
			Threshold: *threshold,
		}
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

func getCmdSignStart(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-sign [message]",
		Short: "Initiate threshold signature protocol",
		Args:  cobra.ExactArgs(1),
	}
	newSigID := cmd.Flags().String("new-sig-id", "", "unique ID for new signature (required)")
	if cmd.MarkFlagRequired("new-sig-id") != nil {
		panic("flag not set")
	}
	keyID := cmd.Flags().String("key-id", "", "unique ID for signature pubkey (required)")
	if cmd.MarkFlagRequired("key-id") != nil {
		panic("flag not set")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		var toSign []byte
		cdc.MustUnmarshalJSON([]byte(args[0]), &toSign)
		msg := types.MsgSignStart{
			Sender:    cliCtx.FromAddress,
			NewSigID:  *newSigID,
			KeyID:     *keyID,
			MsgToSign: toSign,
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

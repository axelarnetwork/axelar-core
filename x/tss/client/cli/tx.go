package cli

import (
	"bufio"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
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
	cmd.MarkFlagRequired("id")
	threshold := cmd.Flags().IntP("threshold", "t", 2, "number of corruptions to withstand (required)")
	cmd.MarkFlagRequired("threshold")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx := context.NewCLIContext().WithCodec(cdc)
		inBuf := bufio.NewReader(cmd.InOrStdin())
		txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

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

func getCmdSignStart(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-sign [message]",
		Short: "Initiate threshold signature protocol",
		Args:  cobra.ExactArgs(1),
	}
	newSigID := cmd.Flags().String("new-sig-id", "", "unique ID for new signature (required)")
	cmd.MarkFlagRequired("new-sig-id")
	keyID := cmd.Flags().String("key-id", "", "unique ID for signature pubkey (required)")
	cmd.MarkFlagRequired("key-id")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx := context.NewCLIContext().WithCodec(cdc)
		inBuf := bufio.NewReader(cmd.InOrStdin())
		txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

		msg := types.MsgSignStart{
			Sender:    cliCtx.FromAddress,
			NewSigID:  *newSigID,
			KeyID:     *keyID,
			MsgToSign: []byte(args[0]),
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

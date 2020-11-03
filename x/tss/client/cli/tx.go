package cli

import (
	"bufio"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssd "github.com/axelarnetwork/tssd/pb"
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
		getCmdTSS(cdc),
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
			Sender:   cliCtx.FromAddress,
			NewSigID: *newSigID,
			KeyID:    *keyID,
			Msg:      []byte(args[0]),
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

// TODO hide this command; it should only be used for testing, never in production
func getCmdTSS(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-tss-message",
		Short: "Relay a message in an in-progress instance of a threshold cryptography protocol",
		Args:  cobra.NoArgs,
	}
	sessionID := cmd.Flags().StringP("session-id", "i", "", "unique ID for protocol (required)")
	cmd.MarkFlagRequired("session-id")
	toParty := cmd.Flags().StringP("to", "t", "", "destination validator address (non-broadcast only)")
	isBroadcast := cmd.Flags().Bool("is-broadcast", false, "is this a broacast message?")
	cmd.MarkFlagRequired("is-broadcast")
	payload := cmd.Flags().BytesBase64P("payload", "p", nil, "message payload")
	cmd.MarkFlagRequired("payload")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx := context.NewCLIContext().WithCodec(cdc)
		inBuf := bufio.NewReader(cmd.InOrStdin())
		txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

		msg := types.MsgTSS{
			Sender:    cliCtx.GetFromAddress(),
			SessionID: *sessionID,
			Payload: &tssd.MessageOut{
				ToPartyUid:  []byte(*toParty),
				IsBroadcast: *isBroadcast,
				Payload:     *payload,
			},
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

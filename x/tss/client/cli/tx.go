package cli

import (
	"bufio"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	// tssd "github.com/axelarnetwork/tssd/pb"
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
		getCmdKeygenMsg(cdc),
	)...)

	return tssTxCmd
}

func getCmdKeygenStart(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "keygenStart [complicated-usage] ", // TODO usage info
		Short: "Initiate threshold key generation protocol",
		Args:  cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// TODO parse keygen metadata into a types.MsgKeygenStart
			msg := types.MsgKeygenStart{
				Sender:  cliCtx.GetFromAddress(),
				Payload: nil, // TODO
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdKeygenMsg(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "keygenMsg [complicated-usage] ", // TODO usage info
		Short: "Relay a message in an in-progress instance of the threshold key generation protocol",
		Args:  cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// TODO parse keygen message into a types.MsgIn
			msg := types.MsgIn{
				Sender:  cliCtx.GetFromAddress(),
				Payload: nil, // TODO
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

package cli

import (
	"bufio"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	broadcastTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	broadcastTxCmd.AddCommand(flags.PostCommands(
		GetCmdRegisterProxy(cdc),
		GetCmdSendStake(cdc),
	)...)

	return broadcastTxCmd
}

func GetCmdRegisterProxy(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "registerProxy [proxy] ",
		Short: "Register a proxy account for a specific validator principal to broadcast transactions in its stead",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			voter, _, err := context.GetFromFields(inBuf, args[0], false)
			if err != nil {
				return sdkerrors.Wrap(types.ErrBroadcast, "proxy invalid")
			}

			msg := types.NewMsgRegisterProxy(sdk.ValAddress(cliCtx.FromAddress), voter)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSendStake(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sendStake [amount] [address 1] ... [address n]",
		Short: "Sends the specified amount of stake to the designated addresses",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			coins, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			if coins.Len() != 1 {
				return fmt.Errorf("Only a single amount is permitted")
			}

			if coins.GetDenomByIndex(0) != "stake" {
				return fmt.Errorf("Invalid denomination")

			}

			inputs := make([]bank.Input, 0)
			outputs := make([]bank.Output, 0)

			for i, addr := range args {

				// ignore the amount argument
				if i == 0 {
					continue
				}

				to, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				inputs = append(inputs, bank.NewInput(cliCtx.FromAddress, coins))
				outputs = append(outputs, bank.NewOutput(to, coins))

			}

			msg := bank.NewMsgMultiSend(inputs, outputs)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

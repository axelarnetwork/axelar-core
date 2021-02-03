package cli

import (
	"bufio"
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
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
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authUtils.GetTxEncoder(cdc))

			voter, _, err := context.GetFromFields(inBuf, args[0], false)
			if err != nil {
				return sdkerrors.Wrap(types.ErrBroadcast, "proxy invalid")
			}

			msg := types.NewMsgRegisterProxy(sdk.ValAddress(cliCtx.FromAddress), voter)
			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSendStake(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sendStake [amount] [address 1] ... [address n]",
		Short: "Sends the specified amount of stake to the designated addresses",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			coins, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			if coins.Len() != 1 {
				return fmt.Errorf("Only a single amount is permitted")
			}

			inputs := make([]bank.Input, 0)
			outputs := make([]bank.Output, 0)

			for _, addr := range args[1:] {

				to, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				inputs = append(inputs, bank.NewInput(cliCtx.FromAddress, coins))
				outputs = append(outputs, bank.NewOutput(to, coins))

			}

			msg := bank.NewMsgMultiSend(inputs, outputs)
			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

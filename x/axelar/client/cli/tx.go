package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	axelarTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	axelarTxCmd.AddCommand(flags.PostCommands(
		GetCmdTrackAddress(cdc),
		GetCmdVerifyTx(cdc),
	)...)

	return axelarTxCmd
}

func GetCmdTrackAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trackAddress [chain] [address] ",
		Short: "Make the axelar network aware of a specific address on another blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			addr := exported.ExternalChainAddress{
				Chain:   args[0],
				Address: args[1],
			}
			msg := types.NewMsgTrackAddress(cliCtx.GetFromAddress(), addr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVerifyTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [chain] [txId] [amount] ",
		Short: "Verify a transaction happened on another chain",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			rawCoin := args[2]
			if strings.Contains(rawCoin, ",") || strings.Contains(rawCoin, ".") {
				return fmt.Errorf("choose denomination so that amount is an integer value")
			}
			coin, err := sdk.ParseCoin(args[2])
			if err != nil {
				return err
			}
			tx := exported.ExternalTx{
				Chain:  args[0],
				TxID:   args[1],
				Amount: coin,
			}
			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), tx)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

package cli

import (
	"bufio"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

const (
	sat     = "sat"
	satoshi = "satoshi"
	btc     = "btc"
	bitcoin = "bitcoin"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        bitcoin,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(flags.PostCommands(
		GetCmdTrackAddress(cdc),
		GetCmdVerifyTx(cdc),
	)...)

	return btcTxCmd
}

func GetCmdTrackAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trackAddress [address] ",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgTrackAddress(cliCtx.GetFromAddress(), args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVerifyTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [txId] [amount] ",
		Short: "Verify a Bitcoin transaction",
		Long: `Verify that a transaction happened on the Bitcoin chain so it can be processed on axelar.
Accepted denominations (case-insensitive): satoshi (sat), bitcoin (btc)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			rawCoin := args[1]

			var decCoin sdk.DecCoin
			decCoin, err := sdk.ParseDecCoin(rawCoin)
			if err != nil {
				coin, err := sdk.ParseCoin(rawCoin)
				if err != nil {
					return fmt.Errorf("could not parse coin string")
				}
				decCoin = sdk.NewDecCoinFromCoin(coin)
			}

			switch decCoin.Denom {
			case sat, satoshi:
				if !decCoin.Amount.IsInteger() {
					return fmt.Errorf("satoshi must be an integer value")
				}
			case btc, bitcoin:
				break
			default:
				return fmt.Errorf("choose a correct denomination: satoshi (sat), bitcoin (btc)")
			}

			tx := exported.ExternalTx{
				Chain:  "bitcoin",
				TxID:   args[0],
				Amount: decCoin,
			}
			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), tx)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

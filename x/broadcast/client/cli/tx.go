package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	legacyTx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	broadcastTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	broadcastTxCmd.AddCommand(
		GetCmdRegisterProxy(),
		GetCmdSendStake(),
	)

	return broadcastTxCmd
}

// GetCmdRegisterProxy returns the command to register a proxy
func GetCmdRegisterProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registerProxy [proxy] ",
		Short: "Register a proxy account for a specific validator principal to broadcast transactions in its stead",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proxyKey, err := clientCtx.Keyring.Key(args[0])
			if err != nil {
				return sdkerrors.Wrap(types.ErrBroadcast, "proxy invalid")
			}

			msg := types.NewRegisterProxyRequest(sdk.ValAddress(clientCtx.FromAddress), proxyKey.GetAddress())
			return legacyTx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSendStake returns the command to send stake to a number of addresses
func GetCmdSendStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendStake [amount] [address 1] ... [address n]",
		Short: "Sends the specified amount of stake to the designated addresses",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			decCoins, err := sdk.ParseDecCoins(args[0])
			if err != nil {
				return err
			}

			if decCoins.Len() != 1 {
				return fmt.Errorf("only a single amount is permitted")
			}

			coins, decimals := decCoins.TruncateDecimal()
			if !decimals.IsZero() {
				return fmt.Errorf("amount must be an integer value")
			}

			inputs := make([]banktypes.Input, 0)
			outputs := make([]banktypes.Output, 0)

			for _, addr := range args[1:] {

				to, err := sdk.AccAddressFromBech32(addr)
				if err != nil {
					return err
				}

				inputs = append(inputs, banktypes.NewInput(clientCtx.FromAddress, coins))
				outputs = append(outputs, banktypes.NewOutput(to, coins))

			}

			msg := banktypes.NewMsgMultiSend(inputs, outputs)
			return legacyTx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

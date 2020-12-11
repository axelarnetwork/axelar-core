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
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(flags.PostCommands(
		getCmdSnapshot(cdc),
	)...)

	return stakingTxCmd
}

func getCmdSnapshot(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Trigger a new snapshot of the current validator set",
		Args:  cobra.NoArgs,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

		msg := types.MsgSnapshot{
			Sender: cliCtx.FromAddress,
		}
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
	}
	return cmd
}

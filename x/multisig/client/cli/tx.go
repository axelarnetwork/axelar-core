package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdStartKeygen(),
	)

	return txCmd
}

// GetCmdStartKeygen returns the cli command to start keygen
func GetCmdStartKeygen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-keygen",
		Short: "Initiate key generation protocol",
		Args:  cobra.NoArgs,
	}

	keyID := cmd.Flags().String("id", "", "unique ID for new key (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic("id flag not set")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		msg := types.NewStartKeygenRequest(cliCtx.FromAddress, exported.KeyID(*keyID))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

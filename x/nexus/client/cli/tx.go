package cli

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
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
		GetCmdRegisterChainMaintainer(),
		GetCmdDeregisterChainMaintainer(),
	)

	return txCmd
}

// GetCmdRegisterChainMaintainer returns the cli command to register a validator as a chain maintainer for the given chains
func GetCmdRegisterChainMaintainer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-chain-maintainer [chains]",
		Short: "register a validator as a chain maintainer for the given chains",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewRegisterChainMaintainerRequest(cliCtx.GetFromAddress(), args...)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeregisterChainMaintainer returns the cli command to deregister a validator as a chain maintainer for the given chains
func GetCmdDeregisterChainMaintainer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister-chain-maintainer [chains]",
		Short: "deregister a validator as a chain maintainer for the given chains",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewDeregisterChainMaintainerRequest(cliCtx.GetFromAddress(), args...)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

const (
	flagLevel                  = "level"
	flagOverridePendingUpgrade = "override-pending-upgrade"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govTxCmd.AddCommand(
		GetCmdRegisterController(),
		GetCmdDeregisterController(),
		GetCmdUpdateGovernanceKey(),
		GetCmdPause(),
		GetCmdUnpause(),
		GetCmdHaltChain(),
	)

	return govTxCmd
}

// GetCmdUpdateGovernanceKey returns the cli command to update the multisig governance key
func GetCmdUpdateGovernanceKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-governance-key [threshold] [[pubKey]...]",
		Short: "Update the multisig governance key for axelar network",
		Args:  cobra.MinimumNArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		threshold, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		var pubKeys []crypto.PubKey
		for i := 1; i < len(args); i++ {
			var pk crypto.PubKey
			err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[i]), &pk)
			if err != nil {
				return err
			}

			pubKeys = append(pubKeys, pk)
		}

		msg := types.NewUpdateGovernanceKeyRequest(clientCtx.GetFromAddress(), threshold, pubKeys...)

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterController returns the cli command to register a controller account
func GetCmdRegisterController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-controller [controller]",
		Short: "Register controller account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			controller, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewRegisterControllerRequest(cliCtx.GetFromAddress(), controller)

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeregisterController returns the cli command to deregister a controller account
func GetCmdDeregisterController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister-controller [controller]",
		Short: "Deregister controller account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			controller, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewDeregisterControllerRequest(cliCtx.GetFromAddress(), controller)

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdPause returns the cli command to pause the chain
func GetCmdPause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause [reason]",
		Short: "Pause the chain: only the messages allowlisted for the pause level are accepted until it is unpaused",
		Long: `Pause the chain at the given level:

- all: only pause, unpause and halt-chain are allowed
- except-validators: like all, plus validator votes, keygen and signing submissions, and permission key rotation

Running it again with a different level changes the level of the ongoing pause.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			levelStr, err := cmd.Flags().GetString(flagLevel)
			if err != nil {
				return err
			}

			level, err := parsePauseLevel(levelStr)
			if err != nil {
				return err
			}

			msg := types.NewPauseRequest(cliCtx.GetFromAddress(), level, args[0])

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagLevel, "all", "pause level, one of: all, except-validators")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUnpause returns the cli command to unpause the chain
func GetCmdUnpause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpause [reason]",
		Short: "Lift the emergency pause",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewUnpauseRequest(cliCtx.GetFromAddress(), args[0])

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdHaltChain returns the cli command to halt the chain at the next block height
func GetCmdHaltChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "halt-chain [reason]",
		Short: "Schedule an upgrade plan at the next block height so that every node stops before executing it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			override, err := cmd.Flags().GetBool(flagOverridePendingUpgrade)
			if err != nil {
				return err
			}

			msg := types.NewHaltChainRequest(cliCtx.GetFromAddress(), args[0], override)

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().Bool(flagOverridePendingUpgrade, false, "replace an already scheduled upgrade plan instead of failing")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parsePauseLevel(level string) (exported.PauseLevel, error) {
	switch strings.ToLower(level) {
	case "all":
		return exported.PAUSE_LEVEL_ALL, nil
	case "except-validators":
		return exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, nil
	default:
		return exported.PAUSE_LEVEL_UNSPECIFIED, fmt.Errorf("unknown pause level %s, expected one of: all, except-validators", level)
	}
}

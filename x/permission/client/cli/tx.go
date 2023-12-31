package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
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

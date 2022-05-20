package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
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
		GetCmdActivateChain(),
		GetCmdDeactivateChain(),
		GetCmdRegisterAssetFee(),
	)

	return txCmd
}

// GetCmdRegisterChainMaintainer returns the cli command to register a validator as a chain maintainer for the given chains
func GetCmdRegisterChainMaintainer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-chain-maintainer [chain]...",
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
		Use:   "deregister-chain-maintainer [chain]...",
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

// GetCmdActivateChain returns the cli command to activate the given chains
func GetCmdActivateChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-chain [chain]...",
		Short: "activate the given chains",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewActivateChainRequest(cliCtx.GetFromAddress(), args...)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeactivateChain returns the cli command to deactivate the given chains
func GetCmdDeactivateChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-chain [chain]...",
		Short: "deactivate the given chains",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewDeactivateChainRequest(cliCtx.GetFromAddress(), args...)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterAssetFee returns the cli command to register an asset fee
func GetCmdRegisterAssetFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-asset-fee [chain] [asset] [fee-rate] [min-fee] [max-fee]",
		Short: "register fees for an asset on a chain",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			feeRate := sdk.MustNewDecFromStr(args[2])

			minFee, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid value provided for min fee")
			}

			maxFee, ok := sdk.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid value provided for max fee")
			}

			feeInfo := exported.NewFeeInfo(exported.ChainName(args[0]), args[1], feeRate, minFee, maxFee)

			msg := types.NewRegisterAssetFeeRequest(cliCtx.GetFromAddress(), feeInfo)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

const (
	flagIsNativeAsset = "is-native-asset"
	flagNativeAsset   = "native-asset"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	axelarTxCmd := &cobra.Command{
		Use:                        "axelarnet",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	axelarTxCmd.AddCommand(
		GetCmdLink(),
		GetCmdConfirmDeposit(),
		GetCmdExecutePendingTransfersTx(),
		GetCmdRegisterIBCPathTx(),
		GetCmdAddCosmosBasedChain(),
		GetCmdRegisterAsset(),
		GetCmdRouteIBCTransfersTx(),
		GetCmdRegisterFeeCollector(),
	)

	return axelarTxCmd
}

// GetCmdLink links a cross chain address to an Axelar chain address
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [recipient chain] [recipient address] [asset]",
		Short: "Link a cross chain address to an Axelar address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewLinkRequest(clientCtx.GetFromAddress(), args[0], args[1], args[2])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmDeposit returns the cli command to confirm a deposit
func GetCmdConfirmDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-deposit [denom] [burnerAddr]",
		Short: "Confirm a deposit to Axelar chain that sent given the asset denomination and the burner address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			burnerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewConfirmDepositRequest(cliCtx.GetFromAddress(), args[0], burnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdExecutePendingTransfersTx returns the cli command to transfer all pending token transfers to Axelar chain
func GetCmdExecutePendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-pending-transfers",
		Short: "Send all pending transfers to Axelar chain",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewExecutePendingTransfersRequest(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterIBCPathTx returns the cli command to register an IBC tracing path for a cosmos chain
func GetCmdRegisterIBCPathTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-path [chain] [path]",
		Short: "Register an ibc path for a cosmos chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewRegisterIBCPathRequest(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddCosmosBasedChain returns the cli command to register a new cosmos based chain in nexus
func GetCmdAddCosmosBasedChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-cosmos-based-chain [name] [address prefix]",
		Short: "Add a new cosmos based chain",
		Args:  cobra.ExactArgs(2),
	}

	nativeAssets := cmd.Flags().StringSlice(flagNativeAsset, []string{}, "denom and minimum transferable amount, e.g. uaxl:1000000")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		if len(*nativeAssets) == 0 {
			return fmt.Errorf("nativeAssets are required")
		}

		assets := make([]nexus.Asset, len(*nativeAssets))
		for i, asset := range *nativeAssets {
			assetInto := strings.Split(asset, ":")
			if len(assetInto) != 2 {
				return fmt.Errorf("denom and minimum amount have to be separated by \":\"")
			}

			denom := utils.NormalizeString(assetInto[0])
			minAmount, ok := sdk.NewIntFromString(assetInto[1])
			if !ok {
				return fmt.Errorf("could not convert string to integer")
			}

			assets[i] = nexus.NewAsset(denom, minAmount, true)
		}

		name := args[0]
		addrPrefix := args[1]

		msg := types.NewAddCosmosBasedChainRequest(cliCtx.GetFromAddress(), name, addrPrefix, assets)
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterAsset returns the cli command to register an asset to a cosmos based chain
func GetCmdRegisterAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-asset [chain] [denom] [min amount]",
		Short: "Register a new asset to a cosmos based chain",
		Args:  cobra.ExactArgs(3),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		chain := args[0]
		denom := utils.NormalizeString(args[1])

		minAmount, ok := sdk.NewIntFromString(args[2])
		if !ok {
			return fmt.Errorf("could not convert string to integer")
		}

		isNativeAsset, err := cmd.Flags().GetBool(flagIsNativeAsset)
		if err != nil {
			return err
		}

		msg := types.NewRegisterAssetRequest(cliCtx.GetFromAddress(), chain, nexus.NewAsset(denom, minAmount, isNativeAsset))
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
	}

	cmd.Flags().Bool(flagIsNativeAsset, false, "is it a native asset from cosmos chain")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRouteIBCTransfersTx returns the cli command to route all pending token transfers to cosmos chains
func GetCmdRouteIBCTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route-ibc-transfers",
		Short: "Routes pending transfers to cosmos chains",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewRouteIBCTransfersRequest(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterFeeCollector returns the cli command to register axelarnet fee collector account
func GetCmdRegisterFeeCollector() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-fee-collector [fee collector]",
		Short: "Register axelarnet fee collector account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			feeCollector, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewRegisterFeeCollectorRequest(cliCtx.GetFromAddress(), feeCollector)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

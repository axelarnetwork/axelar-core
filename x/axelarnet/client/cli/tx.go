package cli

import (
	"fmt"
	"strconv"

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
		GetCmdAddCosmosBasedChain(),
		GetCmdRegisterAsset(),
		GetCmdRouteIBCTransfersTx(),
		GetCmdRegisterFeeCollector(),
		getRetryIBCTransfer(),
		getGeneralMessage(),
	)

	return axelarTxCmd
}

// GetCmdLink links a cross chain address to an Axelar chain address
// Deprecated: link-deposit protocol is being disabled
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "link [recipient chain] [recipient address] [asset]",
		Short:      "Link a cross chain address to an Axelar address",
		Deprecated: "link-deposit protocol is being disabled",
		Args:       cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewLinkRequest(clientCtx.GetFromAddress(), args[0], args[1], args[2])

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmDeposit returns the cli command to confirm a deposit
// Deprecated: link-deposit protocol is being disabled
func GetCmdConfirmDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "confirm-deposit [denom] [burnerAddr]",
		Short:      "Confirm a deposit to Axelar chain that sent given the asset denomination and the burner address",
		Deprecated: "link-deposit protocol is being disabled",
		Args:       cobra.ExactArgs(2),
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

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddCosmosBasedChain returns the cli command to register a new cosmos based chain in nexus
func GetCmdAddCosmosBasedChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-cosmos-based-chain [name] [address prefix] [ibc path] [native asset]...",
		Short: "Add a new cosmos based chain",
		Args:  cobra.MinimumNArgs(3),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		// native assets are optional
		assets := make([]nexus.Asset, len(args[3:]))
		for i, asset := range args[3:] {
			assets[i] = nexus.NewAsset(asset, true)
		}

		name := args[0]
		addrPrefix := args[1]
		path := args[2]

		msg := types.NewAddCosmosBasedChainRequest(cliCtx.GetFromAddress(), name, addrPrefix, assets, path)

		return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRegisterAsset returns the cli command to register an asset to a cosmos based chain
func GetCmdRegisterAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-asset [chain] [denom]",
		Short: "Register a new asset to a cosmos based chain",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		chain := args[0]
		denom := args[1]

		isNativeAsset, err := cmd.Flags().GetBool(flagIsNativeAsset)
		if err != nil {
			return err
		}

		msg := types.NewRegisterAssetRequest(cliCtx.GetFromAddress(), chain, nexus.NewAsset(denom, isNativeAsset))

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

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getRetryIBCTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry-ibc-transfer [transfer ID]",
		Short: "Retry a failed IBC transfer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			transferID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewRetryIBCTransferRequest(cliCtx.GetFromAddress(), nexus.TransferID(transferID))

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getGeneralMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route-message [message ID] [payload]",
		Short: "Route an approved general message to the destination chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id := utils.NormalizeString(args[0])
			payload, err := utils.HexDecode(args[1])
			if err != nil {
				return err
			}

			feegranter, err := cmd.Flags().GetString(flags.FlagFeeGranter)
			if err != nil {
				return err
			}

			var feegranterAcc sdk.AccAddress = nil
			if feegranter != "" {
				feegranterAcc, err = sdk.AccAddressFromBech32(feegranter)
				if err != nil {
					return err
				}
			}

			msg := types.NewRouteMessage(cliCtx.GetFromAddress(), feegranterAcc, id, payload)

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

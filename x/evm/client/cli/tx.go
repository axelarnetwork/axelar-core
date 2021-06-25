package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	evmTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	evmTxCmd.AddCommand(
		GetCmdLink(),
		GetCmdSignTx(),
		GetCmdConfirmChain(),
		GetCmdConfirmERC20TokenDeployment(),
		GetCmdConfirmERC20Deposit(),
		GetCmdConfirmTransferOwnership(),
		GetCmdSignPendingTransfersTx(),
		GetCmdSignDeployToken(),
		GetCmdSignBurnTokens(),
		GetCmdSignTransferOwnership(),
		GetCmdAddChain(),
	)

	return evmTxCmd
}

// GetCmdLink links a cross chain address to an EVM chain address created by Axelar
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [chain] [recipient chain] [recipient address] [symbol]",
		Short: "Link a cross chain address to an EVM chain address created by Axelar",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewLinkRequest(cliCtx.GetFromAddress(), args[0], args[1], args[2], args[3])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignTx returns the cli command to sign the given transaction
func GetCmdSignTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [chain] [tx json file path]",
		Short: "sign a raw EVM chain transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			json, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			var ethtx *ethTypes.Transaction
			cliCtx.LegacyAmino.MustUnmarshalJSON(json, &ethtx)

			msg := types.NewSignTxRequest(cliCtx.GetFromAddress(), chain, json)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmChain returns the cli command to confirm a ERC20 token deployment
func GetCmdConfirmChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-chain [chain]",
		Short: "Confirm an EVM chain for a given name and native asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewConfirmChainRequest(cliCtx.GetFromAddress(), args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmERC20TokenDeployment returns the cli command to confirm a ERC20 token deployment
func GetCmdConfirmERC20TokenDeployment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-erc20-token [chain] [txID] [symbol]",
		Short: "Confirm an ERC20 token deployment in an EVM chain transaction for a given symbol of token and gateway address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID := common.HexToHash(args[1])
			symbol := args[2]
			msg := types.NewConfirmTokenRequest(cliCtx.GetFromAddress(), chain, symbol, txID)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmERC20Deposit returns the cli command to confirm an ERC20 deposit
func GetCmdConfirmERC20Deposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-erc20-deposit [chain] [txID] [amount] [burnerAddr]",
		Short: "Confirm an ERC20 deposit in an EVM chain transaction that sent given amount of token to a burner address",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID := common.HexToHash(args[1])
			amount := sdk.NewUintFromString(args[2])
			burnerAddr := common.HexToAddress(args[3])

			msg := types.NewConfirmDepositRequest(cliCtx.GetFromAddress(), chain, txID, amount, burnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmTransferOwnership returns the cli command to confirm a transfer ownership for the gateway contract
func GetCmdConfirmTransferOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-transfer-ownership [chain] [txID] [newOwnerAddr]",
		Short: "Confirm a transfer ownership in an EVM chain transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID := common.HexToHash(args[1])
			newOwnerAddr := common.HexToAddress(args[2])
			msg := types.NewConfirmTransferOwnershipRequest(cliCtx.GetFromAddress(), chain, txID, newOwnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignPendingTransfersTx returns the cli command to sign all pending token transfers to an EVM chain
func GetCmdSignPendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-pending-transfers [chain]",
		Short: "Sign all pending transfers to an EVM chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewSignPendingTransfersRequest(cliCtx.GetFromAddress(), args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignDeployToken returns the cli command to sign deploy-token command data for an EVM chain
func GetCmdSignDeployToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-deploy-token [chain] [name] [symbol] [decimals] [capacity]",
		Short: "Signs the call data to deploy a token with the AxelarGateway contract",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			name := args[1]
			symbol := args[2]
			decs, err := strconv.ParseUint(args[3], 10, 8)
			if err != nil {
				return fmt.Errorf("could not parse decimals")
			}
			capacity, ok := sdk.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("could not parse capacity")
			}

			msg := types.NewSignDeployTokenRequest(cliCtx.GetFromAddress(), chain, name, symbol, uint8(decs), capacity)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignBurnTokens returns the cli command to sign burn command for all confirmed token deposits in an EVM chain
func GetCmdSignBurnTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-burn-tokens [chain]",
		Short: "Sign burn command for all confirmed token deposits in an EVM chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewSignBurnTokensRequest(cliCtx.GetFromAddress(), args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignTransferOwnership returns the cli command to sign transfer-ownership command for an EVM chain contract
func GetCmdSignTransferOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-ownership [chain] [newOwnerAddr]",
		Short: "Sign transfer ownership command for an EVM chain contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			chain := args[0]
			newOwnerAddr := common.HexToAddress(args[1])

			msg := types.NewSignTransferOwnershipRequest(cliCtx.GetFromAddress(), chain, newOwnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddChain returns the cli command to add a new evm chain command
func GetCmdAddChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-chain [name] [native asset] [chain config]",
		Short: "Add a new EVM chain",
		Long:  "Add a new EVM chain. The chain config parameter should be the path to a json file containing the key requirements and the evm module parameters",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			nativeAsset := args[1]
			jsonFile := args[2]

			byteValue, err := ioutil.ReadFile(jsonFile)
			if err != nil {
				return err
			}
			var chainConf struct {
				KeyRequirement tss.KeyRequirement `json:"key_requirement"`
				Params         types.Params       `json:"params"`
			}
			err = json.Unmarshal([]byte(byteValue), &chainConf)
			if err != nil {
				return err
			}

			msg := types.NewAddChainRequest(cliCtx.GetFromAddress(), name, nativeAsset, chainConf.KeyRequirement, chainConf.Params)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

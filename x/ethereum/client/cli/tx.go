package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	ethTxCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	ethTxCmd.AddCommand(
		GetCmdLink(),
		GetCmdSignTx(),
		GetCmdConfirmERC20TokenDeploy(),
		GetCmdConfirmERC20Deposit(),
		GetCmdSignPendingTransfersTx(),
		GetCmdSignDeployToken(),
		GetCmdSignBurnTokens(),
		GetCmdSignTransferOwnership(),
	)

	return ethTxCmd
}

// GetCmdLink links a cross chain address to an ethereum address created by Axelar
func GetCmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [chain] [address] [symbol]",
		Short: "Link a cross chain address to an ethereum address created by Axelar",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			msg := &types.MsgLink{
				Sender:         cliCtx.GetFromAddress(),
				RecipientChain: args[0],
				RecipientAddr:  args[1],
				Symbol:         args[2],
			}
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
		Use:   "sign [tx json file path]",
		Short: "sign a raw Ethereum transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			json, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			var ethtx *ethTypes.Transaction
			cliCtx.LegacyAmino.MustUnmarshalJSON(json, &ethtx)

			msg := types.NewMsgSignTx(cliCtx.GetFromAddress(), json)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdConfirmERC20TokenDeploy returns the cli command to confirm a ERC20 token deployment
func GetCmdConfirmERC20TokenDeploy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-erc20-token [txID] [symbol]",
		Short: "Confirm an ERC20 token deployment in an Ethereum transaction for a given symbol of token and gateway address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			txID := common.HexToHash(args[0])
			msg := types.NewMsgConfirmERC20TokenDeploy(cliCtx.GetFromAddress(), txID, args[1])
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
		Use:   "confirm-erc20-deposit [txID] [amount] [burnerAddr]",
		Short: "Confirm an ERC20 deposit in an Ethereum transaction that sent given amount of token to a burner address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			txID := common.HexToHash(args[0])
			amount := sdk.NewUintFromString(args[1])
			burnerAddr := common.HexToAddress(args[2])

			msg := types.NewMsgConfirmERC20Deposit(cliCtx.GetFromAddress(), txID, amount, burnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignPendingTransfersTx returns the cli command to sign all pending token transfers to Ethereum
func GetCmdSignPendingTransfersTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-pending-transfers",
		Short: "Sign all pending transfers to Ethereum",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			msg := types.NewMsgSignPendingTransfers(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignDeployToken returns the cli command to sign deploy-token command data for Ethereum
func GetCmdSignDeployToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-deploy-token [name] [symbol] [decimals] [capacity]",
		Short: "Signs the call data to deploy a token with the AxelarGateway contract",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			decs, err := strconv.ParseUint(args[2], 10, 8)
			if err != nil {
				return fmt.Errorf("could not parse decimals")
			}

			capacity, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("could not parse capacity")
			}
			msg := types.NewMsgSignDeployToken(cliCtx.GetFromAddress(), args[0], args[1], uint8(decs), capacity)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignBurnTokens returns the cli command to sign burn command for all confirmed Ethereum token deposits
func GetCmdSignBurnTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-burn-tokens",
		Short: "Sign burn command for all confirmed Ethereum token deposits",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)

			msg := types.NewMsgSignBurnTokens(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignTransferOwnership returns the cli command to sign transfer-ownership command for Ethereum contract
func GetCmdSignTransferOwnership() *cobra.Command {
	cmd:=  &cobra.Command{
		Use:   "transfer-ownership [newOwnerAddr]",
		Short: "Sign transfer ownership command for Ethereum contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)
			newOwnerAddr := common.HexToAddress(args[0])

			msg := types.NewMsgSignTransferOwnership(cliCtx.GetFromAddress(), newOwnerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

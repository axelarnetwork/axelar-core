package cli

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/evm/keeper"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	ethQueryCmd := &cobra.Command{
		Use:                        "evm",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ethQueryCmd.AddCommand(
		GetCmdMasterAddress(queryRoute),
		GetCmdAxelarGatewayAddress(queryRoute),
		GetCmdTokenAddress(queryRoute),
		GetCmdCreateDeployTx(queryRoute),
		GetCmdSendTx(queryRoute),
		GetCmdSendCommand(queryRoute),
		GetCmdQueryCommandData(queryRoute),
	)

	return ethQueryCmd

}

// GetCmdMasterAddress returns the query for an EVM chain master address that owns the AxelarGateway contract
func GetCmdMasterAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Query an address by key ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterAddress), nil)
			if err != nil {
				fmt.Printf(types.ErrFMasterKey, err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintObjectLegacy(out.Hex())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdTokenAddress returns the query for an EVM chain master address that owns the AxelarGateway contract
func GetCmdTokenAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-address [symbol]",
		Short: "Query a token address by symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryTokenAddress, args[0]), nil)
			if err != nil {
				fmt.Printf(types.ErrFTokenAddress, err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintObjectLegacy(out.Hex())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdAxelarGatewayAddress returns the query for the AxelarGateway contract address
func GetCmdAxelarGatewayAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway-address",
		Short: "Query the Axelar Gateway contract address",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryAxelarGatewayAddress), nil)
			if err != nil {
				fmt.Printf(types.ErrFGatewayAddress, err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintObjectLegacy(out.Hex())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateDeployTx returns the query for a raw unsigned EVM deploy transaction for the smart contract of a given path
func GetCmdCreateDeployTx(queryRoute string) *cobra.Command {
	var gasPriceStr string
	var gasLimit uint64
	cmd := &cobra.Command{
		Use:   "deploy-gateway",
		Short: "Obtain a raw transaction for the deployment of Axelar Gateway.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			gasPriceBig, ok := big.NewInt(0).SetString(gasPriceStr, 10)
			if !ok {
				return fmt.Errorf("could not parse specified gas price")
			}

			gasPrice := sdk.NewIntFromBigInt(gasPriceBig)

			params := types.DeployParams{
				GasPrice: gasPrice,
				GasLimit: gasLimit,
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.CreateDeployTx), cliCtx.LegacyAmino.MustMarshalJSON(params))
			if err != nil {
				fmt.Printf(types.ErrFDeployTx, err.Error())

				return nil
			}

			var result types.DeployResult
			cliCtx.LegacyAmino.MustUnmarshalJSON(res, &result)
			fmt.Println(string(cliCtx.LegacyAmino.MustMarshalJSON(result.Tx)))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().Uint64Var(&gasLimit, "gas-limit", 3000000,
		"Ethereum gas limit to use in the transaction (default value is 3000000). Set to 0 to estimate gas limit at the node.")
	cmd.Flags().StringVar(&gasPriceStr, "gas-price", "0",
		"Ethereum gas price to use in the transaction. If flag is omitted (or value set to 0), the gas price will be suggested by the node")
	return cmd
}

// GetCmdSendTx sends a transaction to an EVM chain
func GetCmdSendTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendTx [txID]",
		Short: "Send a transaction that spends tx [txID] to chain [chain]",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.SendTx, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFSendTx, args[0])
			}

			return cliCtx.PrintObjectLegacy(fmt.Sprintf("successfully sent transaction %s to Ethereum", common.BytesToHash(res).Hex()))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSendCommand returns the query to send a signed command from an externally controlled address to the specified contract
func GetCmdSendCommand(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendCommand [commandID] [fromAddress]",
		Short: "Send a transaction signed by [fromAddress] that executes the command [commandID] to Axelar Gateway",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			var commandID types.CommandID
			copy(commandID[:], common.Hex2Bytes(args[0]))
			params := types.CommandParams{
				CommandID: commandID,
				Sender:    args[1],
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.SendCommand), cliCtx.LegacyAmino.MustMarshalJSON(params))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send Ethereum transaction executing command %s", commandID)
			}

			return cliCtx.PrintObjectLegacy(fmt.Sprintf("successfully sent transaction %s to Ethereum", common.BytesToHash(res).Hex()))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCommandData returns the query to get the signed command data
func GetCmdQueryCommandData(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command [commandID]",
		Short: "Get the signed command data that can be wrapped in an Ethereum transaction to execute the command [commandID] on Axelar Gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			commandIDHex := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryCommandData, commandIDHex), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not get command %s", commandIDHex)
			}

			return cliCtx.PrintObjectLegacy(common.Bytes2Hex(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

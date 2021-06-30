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
	evmQueryCmd := &cobra.Command{
		Use:                        "evm",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmQueryCmd.AddCommand(
		GetCmdDepositAddress(queryRoute),
		GetCmdMasterAddress(queryRoute),
		GetCmdAxelarGatewayAddress(queryRoute),
		GetCmdTokenAddress(queryRoute),
		GetCmdCreateDeployTx(queryRoute),
		GetCmdBytecodes(queryRoute),
		GetCmdSignedTx(queryRoute),
		GetCmdSendTx(queryRoute),
		GetCmdSendCommand(queryRoute),
		GetCmdQueryCommandData(queryRoute),
	)

	return evmQueryCmd

}

// GetCmdDepositAddress returns the deposit address command
func GetCmdDepositAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-address [evm chain] [recipient chain] [recipient address] [symbol]",
		Short: "Returns an evm chain deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QDepositAddress, args[0])

			res, _, err := cliCtx.QueryWithData(path, types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{Chain: args[1], Address: args[2], Symbol: args[3]}))
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFDepositAddress)
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintObjectLegacy(out.Hex())
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdMasterAddress returns the query for an EVM chain master address that owns the AxelarGateway contract
func GetCmdMasterAddress(queryRoute string) *cobra.Command {
	var IncludeKeyID bool
	cmd := &cobra.Command{
		Use:   "master-address [chain]",
		Short: "Returns the EVM address of the current master key, and optionally the key's ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QMasterAddress, args[0]), nil)
			if err != nil {
				fmt.Printf(types.ErrFMasterKey, err.Error())

				return nil
			}

			var resp types.QueryMasterAddressResponse
			err = resp.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrap(err, types.ErrFMasterKey)
			}

			if IncludeKeyID {
				return cliCtx.PrintObjectLegacy(resp)
			}

			address := common.BytesToAddress(resp.Address)

			return cliCtx.PrintObjectLegacy(address.Hex())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().BoolVar(&IncludeKeyID, "include-key-id", false, "include the current master key ID in the output")
	return cmd
}

// GetCmdTokenAddress returns the query for an EVM chain master address that owns the AxelarGateway contract
func GetCmdTokenAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-address [chain] [symbol]",
		Short: "Query a token address by symbol",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QTokenAddress, args[0], args[1]), nil)
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
		Use:   "gateway-address [chain]",
		Short: "Query the Axelar Gateway contract address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QAxelarGatewayAddress, args[0]), nil)
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
		Use:   "deploy-gateway [chain]",
		Short: "Obtain a raw transaction for the deployment of Axelar Gateway.",
		Args:  cobra.ExactArgs(1),
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
				Chain:    args[0],
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
		"EVM gas limit to use in the transaction (default value is 3000000). Set to 0 to estimate gas limit at the node.")
	cmd.Flags().StringVar(&gasPriceStr, "gas-price", "0",
		"EVM gas price to use in the transaction. If flag is omitted (or value set to 0), the gas price will be suggested by the node")
	return cmd
}

// GetCmdBytecodes fetches the bytecodes of an EVM contract
func GetCmdBytecodes(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bytecode [chain] [contract]",
		Short: "Fetch the bytecodes of an EVM contract [contract] for chain [chain]",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QBytecodes, args[0], args[1]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFBytecodes, args[1])
			}

			fmt.Println(common.Bytes2Hex(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignedTx fetches an EVM transaction that has been signed by the validators
func GetCmdSignedTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signedTx [chain] [txID]",
		Short: "Fetch an EVM transaction [txID] that has been signed by the validators for chain [chain]",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QSignedTx, args[0], args[1]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFSignedTx, args[1])
			}

			fmt.Println(string(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSendTx sends a transaction to an EVM chain
func GetCmdSendTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendTx [chain] [txID]",
		Short: "Send a transaction that spends tx [txID] to chain [chain]",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.SendTx, args[0], args[1]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFSendTx, args[1])
			}

			return cliCtx.PrintObjectLegacy(fmt.Sprintf("successfully sent transaction %s to %s", common.BytesToHash(res).Hex(), args[0]))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSendCommand returns the query to send a signed command from an externally controlled address to the specified contract
func GetCmdSendCommand(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sendCommand [chain] [commandID] [fromAddress]",
		Short: "Send a transaction signed by [fromAddress] that executes the command [commandID] to Axelar Gateway",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			var commandID types.CommandID
			copy(commandID[:], common.Hex2Bytes(args[1]))
			params := types.CommandParams{
				Chain:     args[0],
				CommandID: commandID,
				Sender:    args[2],
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.SendCommand), cliCtx.LegacyAmino.MustMarshalJSON(params))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send %s transaction executing command %s", args[0], commandID)
			}

			return cliCtx.PrintObjectLegacy(fmt.Sprintf("successfully sent transaction %s to %s", common.BytesToHash(res).Hex(), args[0]))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCommandData returns the query to get the signed command data
func GetCmdQueryCommandData(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command [chain] [commandID]",
		Short: "Get the signed command data that can be wrapped in an EVM transaction to execute the command [commandID] on Axelar Gateway",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			commandIDHex := args[1]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QCommandData, chain, commandIDHex), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not get command %s", commandIDHex)
			}

			return cliCtx.PrintObjectLegacy(common.Bytes2Hex(res))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

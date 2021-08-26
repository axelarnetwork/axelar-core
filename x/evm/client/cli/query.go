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
		GetCmdAddress(queryRoute),
		GetCmdAxelarGatewayAddress(queryRoute),
		GetCmdTokenAddress(queryRoute),
		GetCmdDepositState(queryRoute),
		GetCmdCreateDeployTx(queryRoute),
		GetCmdBytecode(queryRoute),
		GetCmdSignedTx(queryRoute),
		GetCmdSendTx(queryRoute),
		GetCmdQueryBatchedCommands(queryRoute),
		GetCmdLatestBatchedCommands(queryRoute),
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

// GetCmdAddress returns the query for an EVM chain address
func GetCmdAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address [chain]",
		Short: "Returns the EVM address",
		Args:  cobra.ExactArgs(1),
	}
	keyRole := cmd.Flags().String("key-role", "", "the role of the key to get the address for")
	keyID := cmd.Flags().String("key-id", "", "the ID of the key to get the address for")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		var query string
		var param string
		switch {
		case *keyRole != "" && *keyID == "":
			query = keeper.QAddressByKeyRole
			param = *keyRole
		case *keyRole == "" && *keyID != "":
			query = keeper.QAddressByKeyID
			param = *keyID
		default:
			return fmt.Errorf("one and only one of the two flags key-role and key-id has to be set")
		}

		bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, query, args[0], param))
		if err != nil {
			return sdkerrors.Wrap(err, types.ErrAddress)
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

		return clientCtx.PrintProto(&res)
	}

	flags.AddQueryFlagsToCmd(cmd)
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

// GetCmdDepositState returns the query for an ERC20 deposit transaction state
func GetCmdDepositState(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-state [chain] [txID] [deposit address]",
		Short: "Query the state of a deposit transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s/%s", queryRoute, keeper.QDepositState, args[0], args[1], args[2]), nil)
			if err != nil {
				fmt.Printf(types.ErrFTokenAddress, err.Error())

				return nil
			}

			var res types.QueryDepositStateResponse
			types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

			return cliCtx.PrintProto(&res)
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

// GetCmdBytecode fetches the bytecodes of an EVM contract
func GetCmdBytecode(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bytecode [chain] [contract]",
		Short: "Fetch the bytecodes of an EVM contract [contract] for chain [chain]",
		Long: fmt.Sprintf("Fetch the bytecodes of an EVM contract [contract] for chain [chain]. "+
			"The value for [contract] can be either '%s', '%s', or '%s'.",
			keeper.BCGateway, keeper.BCToken, keeper.BCBurner),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QBytecode, args[0], args[1]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFBytecode, args[1])
			}

			fmt.Println("0x" + common.Bytes2Hex(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignedTx fetches an EVM transaction that has been signed by the validators
func GetCmdSignedTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signed-tx [chain] [txID]",
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

			fmt.Println("0x" + common.Bytes2Hex(res))
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdSendTx sends a transaction to an EVM chain
func GetCmdSendTx(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-tx [chain] [txID]",
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

// GetCmdQueryBatchedCommands returns the query to get the batched commands
func GetCmdQueryBatchedCommands(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batched-commands [chain] [batchedCommandsID]",
		Short: "Get the signed batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			idHex := args[1]

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QBatchedCommands, chain, idHex))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not get batched commands %s", idHex)
			}

			var res types.QueryBatchedCommandsResponse
			types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdLatestBatchedCommands returns the query to get the latest batched commands
func GetCmdLatestBatchedCommands(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-batched-commands [chain]",
		Short: "Get the latest batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QLatestBatchedCommands, chain))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not get the latest batched commands for chain %s", chain)
			}

			var res types.QueryBatchedCommandsResponse
			types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		GetCmdBytecode(queryRoute),
		GetCmdSignedTx(queryRoute),
		GetCmdQueryBatchedCommands(queryRoute),
		GetCmdLatestBatchedCommands(queryRoute),
	)

	return evmQueryCmd

}

// GetCmdDepositAddress returns the deposit address command
func GetCmdDepositAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-address [evm chain] [recipient chain] [recipient address] [asset]",
		Short: "Returns an evm chain deposit address for a recipient address on another blockchain",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QDepositAddress, args[0])

			res, _, err := cliCtx.QueryWithData(path, types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{Chain: args[1], Address: args[2], Asset: args[3]}))
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
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

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
		Use:   "deposit-state [chain] [txID] [burner address] [amount]",
		Short: "Query the state of a deposit transaction",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID := common.HexToHash(args[1])
			burnerAddress := common.HexToAddress(args[2])
			amount := sdk.NewUintFromString(args[3])

			params := types.QueryDepositStateParams{
				TxID:          types.Hash(txID),
				BurnerAddress: types.Address(burnerAddress),
				Amount:        amount.Uint64(),
			}
			data := types.ModuleCdc.MustMarshalJSON(&params)

			bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QDepositState, chain), data)
			if err != nil {
				fmt.Printf(types.ErrFTokenAddress, err.Error())

				return nil
			}

			var res types.QueryDepositStateResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

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

// GetCmdBytecode fetches the bytecodes of an EVM contract
func GetCmdBytecode(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bytecode [chain] [contract]",
		Short: "Fetch the bytecodes of an EVM contract [contract] for chain [chain]",
		Long: fmt.Sprintf("Fetch the bytecodes of an EVM contract [contract] for chain [chain]. "+
			"The value for [contract] can be either '%s', '%s', '%s', or '%s'.",
			keeper.BCGateway, keeper.BCGatewayDeployment, keeper.BCToken, keeper.BCBurner),
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
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

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
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

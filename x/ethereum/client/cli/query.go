package cli

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ethQueryCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ethQueryCmd.AddCommand(flags.GetCommands(
		GetCmdMasterAddress(queryRoute, cdc),
		GetCmdAxelarGatewayAddress(queryRoute, cdc),
		GetCmdCreateDeployTx(queryRoute, cdc),
		GetCmdSendTx(queryRoute, cdc),
		GetCmdSendCommand(queryRoute, cdc),
	)...)

	return ethQueryCmd

}

// GetCmdMasterAddress returns the query for the ethereum master address that owns the AxelarGateway contract
func GetCmdMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Query an address by key ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterAddress), nil)
			if err != nil {
				fmt.Printf(types.ErrFMasterKey, err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintOutput(out.Hex())
		},
	}

	return cmd
}

// GetCmdAxelarGatewayAddress returns the query for the AxelarGateway contract address
func GetCmdAxelarGatewayAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway-address",
		Short: "Query the Axelar Gateway contract address",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryAxelarGatewayAddress), nil)
			if err != nil {
				fmt.Printf(types.ErrFMasterKey, err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintOutput(out.Hex())
		},
	}

	return cmd
}

// GetCmdCreateDeployTx returns the query for a raw unsigned Ethereum deploy transaction for the smart contract of a given path
func GetCmdCreateDeployTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	var gasPriceStr string
	var gasLimit uint64
	cmd := &cobra.Command{
		Use:   "deploy-gateway",
		Short: "Obtain a raw transaction for the deployment of Axelar Gateway.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			gasPriceBig, ok := big.NewInt(0).SetString(gasPriceStr, 10)
			if !ok {
				return fmt.Errorf("could not parse specified gas price")
			}

			gasPrice := sdk.NewIntFromBigInt(gasPriceBig)

			params := types.DeployParams{
				GasPrice: gasPrice,
				GasLimit: gasLimit,
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.CreateDeployTx), cdc.MustMarshalJSON(params))
			if err != nil {
				fmt.Printf(types.ErrFDeployTx, err.Error())

				return nil
			}

			var result types.DeployResult
			cdc.MustUnmarshalJSON(res, &result)
			fmt.Println(string(cdc.MustMarshalJSON(result.Tx)))
			return nil
		},
	}
	cmd.Flags().Uint64Var(&gasLimit, "gas-limit", 3000000,
		"Ethereum gas limit to use in the transaction (default value is 3000000). Set to 0 to estimate gas limit at the node.")
	cmd.Flags().StringVar(&gasPriceStr, "gas-price", "0",
		"Ethereum gas price to use in the transaction. If flag is omitted (or value set to 0), the gas price will be suggested by the node")
	return cmd
}

// GetCmdSendTx sends a transaction to Ethereum
func GetCmdSendTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sendTx [txID]",
		Short: "Send a transaction that spends tx [txID] to Ethereum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.SendTx, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFSendTx, args[0])
			}

			var result types.SendTxResult
			cdc.MustUnmarshalJSON(res, &result)

			return cliCtx.PrintOutput(fmt.Sprintf("successfully sent transaction %s to Ethereum", result.SignedTx.Hash().String()))
		},
	}
}

// GetCmdSendCommand returns the query to send a signed command from an externally controlled address to the specified contract
func GetCmdSendCommand(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sendCommand [commandID] [fromAddress]",
		Short: "Send a transaction signed by [fromAddress] that executes the command [commandID] to Axelar Gateway",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			var commandID types.CommandID
			copy(commandID[:], common.Hex2Bytes(args[0]))
			params := types.CommandParams{
				CommandID: commandID,
				Sender:    args[1],
			}
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.SendCommand), cdc.MustMarshalJSON(params))
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send Ethereum transaction executing command %s", commandID)
			}

			var txHash string
			cliCtx.Codec.MustUnmarshalJSON(res, &txHash)

			return cliCtx.PrintOutput(fmt.Sprintf("successfully sent transaction %s to Ethereum", txHash))
		},
	}
}

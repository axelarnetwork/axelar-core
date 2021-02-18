package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

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

// GetCmdCreateDeployTx returns the query for a raw unsigned Ethereum deploy transaction for the smart contract of a given path
func GetCmdCreateDeployTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	var gasLimit uint64
	cmd := &cobra.Command{
		Use:   "deploy [smart contract file path]",
		Short: "Receive a raw deploy transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, err := parseByteCode(args[0])
			if err != nil {
				return err
			}

			params := types.DeployParams{
				ByteCode: bz,
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
	cmd.Flags().Uint64Var(&gasLimit, "gas-limit", 3000000, "default Ethereum gas limit")
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
			cdc.MustUnmarshalJSON(res, &res)

			return cliCtx.PrintOutput(fmt.Sprintf("successfully sent transaction %s to Ethereum", result.SignedTx.Hash().String()))
		},
	}
}

// GetCmdSendCommand returns the query to send a signed command from an externally controlled address to the specified contract
func GetCmdSendCommand(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "sendCommand [commandID] [fromAddress] [contractAddress]",
		Short: "Send a transaction signed by [fromAddress] that executes the command [commandID] to Ethereum contract at [contractAddress]",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			var commandID types.CommandID
			copy(commandID[:], common.Hex2Bytes(args[0]))
			params := types.CommandParams{
				CommandID:    commandID,
				Sender:       args[1],
				ContractAddr: args[2],
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

func parseByteCode(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	byteCode := common.FromHex(strings.TrimSuffix(string(content), "\n"))
	return byteCode, nil
}

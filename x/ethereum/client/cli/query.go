package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils/denom"
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
		GetCmdCreateMintTx(queryRoute, cdc),
		GetCmdCreateDeployTx(queryRoute, cdc),
		GetCmdSendTx(queryRoute, cdc),
	)...)

	return ethQueryCmd

}

func GetCmdMasterAddress(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "master-address",
		Short: "Query an address by key ID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryMasterKey), nil)
			if err != nil {
				fmt.Printf("could not resolve master key: %s\n", err.Error())

				return nil
			}

			out := common.BytesToAddress(res)
			return cliCtx.PrintOutput(out.Hex())
		},
	}

	return cmd
}

func GetCmdCreateMintTx(queryRoute string, cdc *codec.Codec) *cobra.Command {
	var gasLimit uint64
	cmd := &cobra.Command{
		Use:   "mint [contractAddr] [recipient] [amount]",
		Short: "Receive a raw mint transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			amount, err := denom.ParseSatoshi(args[2])
			if err != nil {
				return err
			}

			// check if the addresses are valid
			if !validAddress(args[0]) {
				return fmt.Errorf("invalid contract address")
			}

			if !validAddress(args[1]) {
				return fmt.Errorf("invalid recipient address")
			}

			params := types.MintParams{
				Recipient:    args[1],
				Amount:       amount.Amount,
				ContractAddr: args[0],
				GasLimit:     gasLimit,
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.CreateMintTx), cdc.MustMarshalJSON(params))
			if err != nil {
				fmt.Printf("could not resolve master key: %s\n", err.Error())

				return nil
			}

			var tx *ethTypes.Transaction
			cdc.MustUnmarshalJSON(res, &tx)
			fmt.Println(string(cdc.MustMarshalJSON(tx)))
			return nil
		},
	}
	cmd.Flags().Uint64Var(&gasLimit, "gas-limit", 3000000, "default Ethereum gas limit")
	return cmd
}

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
				fmt.Printf("could not resolve master key: %s\n", err.Error())

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
		Use:   "send [txID]",
		Short: "Send a transaction that spends tx [txID] to Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.SendTx, args[0]), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not send the transaction spending transaction %s", args[0])
			}

			var out string
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
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

func validAddress(address string) bool {

	if bytes.Equal(common.HexToAddress(address).Bytes(), make([]byte, common.AddressLength)) {

		return false

	}

	return true
}

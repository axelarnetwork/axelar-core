package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/common"
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

	ethQueryCmd.AddCommand(flags.GetCommands(GetCmdCreateMintTx(queryRoute, cdc), GetCmdCreateDeployTx(queryRoute, cdc))...)

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
		Use:   "mint [contractID] [recipient] [amount]",
		Short: "Receive a raw mint transaction",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			amount, err := denom.ParseSatoshi(args[3])
			if err != nil {
				return err
			}

			params := types.MintParams{
				Recipient:  common.HexToAddress(args[1]),
				Amount:     amount.Amount,
				ContractID: args[0],
				GasLimit:   gasLimit,
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.CreateMintTx), cdc.MustMarshalJSON(params))
			if err != nil {
				fmt.Printf("could not resolve master key: %s\n", err.Error())

				return nil
			}

			return cliCtx.PrintOutput(res)
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
		Args:  cobra.ExactArgs(0),
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

			return cliCtx.PrintOutput(res)
		},
	}
	cmd.Flags().Uint64Var(&gasLimit, "gas-limit", 3000000, "default Ethereum gas limit")
	return cmd
}

func parseByteCode(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	byteCode := common.FromHex(strings.TrimSuffix(string(content), "\n"))
	return byteCode, nil
}

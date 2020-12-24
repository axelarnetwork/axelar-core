package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	ethTxCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	ethTxCmd.AddCommand(flags.PostCommands(GetCmdInstallSC(cdc))...)

	nets := []types.Network{types.Mainnet, types.Ropsten, types.Kovan, types.Rinkeby, types.Goerli, types.Ganache}
	for _, net := range nets {
		rawTxCmd := makeCommand("raw")
		rawTxCmd.AddCommand(flags.PostCommands(GetCmdDeploy(net, cdc), GetCmdMint(net, cdc))...)

		verifyTxCmd := makeCommand("verify")
		verifyTxCmd.AddCommand(flags.PostCommands(GetCmdVerifyMintTx(net, cdc), GetCmdVerifyDeployTx(net, cdc))...)

		sendCmd := GetCmdSend(cdc)

		netRootCmd := makeCommand(string(net))
		netRootCmd.AddCommand(rawTxCmd, verifyTxCmd, sendCmd)

		ethTxCmd.AddCommand(netRootCmd)
	}

	return ethTxCmd
}

func makeCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:                        name,
		Short:                      fmt.Sprintf("%s transactions subcommands", name),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
}

func GetCmdSend(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [sourceTxId] [sigId]",
		Short: "Submit the specified transaction to ethereum with the specified signature",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgSendTx(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdDeploy(net types.Network, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [contract ID]",
		Short: "deploy a contract controlled by the master key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgRawTxForDeploy(cliCtx.GetFromAddress(), net, args[0])

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdMint(net types.Network, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [contract ID] [amount] [destination]",
		Short: "mint BTC tokens transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			sat, err := denom.ParseSatoshi(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgRawTxForMint(cliCtx.GetFromAddress(), net, args[0], sat.Amount, common.HexToAddress(args[2]))

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdInstallSC(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{

		Use:   "installSC [contract ID] [file path] ",
		Short: "Install an ethereum smart contract in Axelar",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			byteCode, err := parseByteCode(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgInstallSC(cliCtx.GetFromAddress(), args[0], byteCode)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})

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

func GetCmdVerifyMintTx(network types.Network, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mint [txId] [destination] [amount] ",
		Short: "Verify an Ethereum transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			sat, err := denom.ParseSatoshi(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgVerifyMintTx(cliCtx.GetFromAddress(), network, common.HexToHash(args[0]), common.HexToAddress(args[1]), sat.Amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVerifyDeployTx(network types.Network, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy [tx hash] [contract ID] ",
		Short: "Verify an Ethereum transaction",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			hash := common.HexToHash(args[0])

			msg := types.NewMsgVerifyDeployTx(cliCtx.GetFromAddress(), network, hash, args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

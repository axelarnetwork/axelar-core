package cli

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	cliUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	eth      = "eth"
	ethereum = "ethereum"
	wei      = "wei"
	gwei     = "gwei"
	btc      = "btc"
	bitcoin  = "bitcoin"

	//TODO: not sure how many decimals a BTC token is supposed to have in ethereum
	btcDecs = 18

	masterKeyFlag = "master-key"
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

	ethTxCmd.AddCommand(flags.PostCommands(
		GetCmdInstallSC(cdc),
	)...)

	mainnetCmd := makeCommand(types.Mainnet)
	mainnetEtherCmd := makeCommand("rawTx")
	mainnetEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdDeploy(types.Chain(types.Mainnet), cdc),
			GetCmdEther(types.Chain(types.Mainnet), cdc),
			GetCmdMint(types.Chain(types.Mainnet), cdc))...)
	mainnetCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Mainnet), cdc))...)

	ropstenCmd := makeCommand(types.Ropsten)
	ropstenEtherCmd := makeCommand("rawTx")
	ropstenEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdDeploy(types.Chain(types.Ropsten), cdc),
			GetCmdEther(types.Chain(types.Ropsten), cdc),
			GetCmdMint(types.Chain(types.Ropsten), cdc))...)
	ropstenCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Ropsten), cdc))...)

	kovanCmd := makeCommand(types.Kovan)
	kovanEtherCmd := makeCommand("rawTx")
	kovanEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdDeploy(types.Chain(types.Kovan), cdc),
			GetCmdEther(types.Chain(types.Kovan), cdc),
			GetCmdMint(types.Chain(types.Kovan), cdc))...)
	kovanCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Kovan), cdc))...)

	rinkebyCmd := makeCommand(types.Rinkeby)
	rinkebyEtherCmd := makeCommand("rawTx")
	rinkebyEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdDeploy(types.Chain(types.Rinkeby), cdc),
			GetCmdEther(types.Chain(types.Rinkeby), cdc),
			GetCmdMint(types.Chain(types.Rinkeby), cdc))...)
	rinkebyCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Rinkeby), cdc))...)

	goerliCmd := makeCommand(types.Goerli)
	goerliEtherCmd := makeCommand("rawTx")
	goerliEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdDeploy(types.Chain(types.Goerli), cdc),
			GetCmdEther(types.Chain(types.Goerli), cdc),
			GetCmdMint(types.Chain(types.Goerli), cdc))...)
	goerliCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Goerli), cdc))...)

	ethTxCmd.AddCommand(mainnetCmd, ropstenCmd, kovanCmd, rinkebyCmd, goerliCmd)

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

func GetCmdDeploy(chain types.Chain, cdc *codec.Codec) *cobra.Command {

	var useMasterKey bool
	var deployer string

	cmd := &cobra.Command{
		Use:   "deploy [contract ID] [-d <deployer address> | -m]",
		Short: "deploy smart contract transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			if (deployer == "" && !useMasterKey) || (deployer != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the destination or to use the master key, not both\"")
			}

			var msg sdk.Msg
			if useMasterKey {
				msg = types.NewMsgRawTxForNextMasterKey(cliCtx.GetFromAddress(), nil, *big.NewInt(0), []byte(args[0]), types.TypeSCdeploy)
			} else {
				addr, err := types.ParseEthAddress(deployer, chain)
				if err != nil {
					return err
				}

				msg = types.NewMsgRawTx(cliCtx.GetFromAddress(), nil, *big.NewInt(0), []byte(args[0]), addr, types.TypeSCdeploy)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	addDestinationFlag(cmd, &deployer)
	addMasterKeyFlag(cmd, &useMasterKey)
	return cmd
}

func GetCmdMint(chain types.Chain, cdc *codec.Codec) *cobra.Command {

	var useMasterKey bool
	var destination string

	cmd := &cobra.Command{
		Use:   "mint [sourceTxId] [amount] [contract address] [-d <destination> | -m]",
		Short: "mint BTC tokens transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			hash := common.HexToHash(args[0])

			eth, txType, err := parseValue(args[1])
			if err != nil {
				return err
			}

			if txType != types.TypeERC20mint {

				return fmt.Errorf("amount given must be a unit of bitcoin")
			}

			if (destination == "" && !useMasterKey) || (destination != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the destination or to use the master key, not both\"")
			}

			var msg sdk.Msg

			data := common.HexToAddress(args[2]).Bytes()

			if useMasterKey {
				msg = types.NewMsgRawTxForNextMasterKey(cliCtx.GetFromAddress(), &hash, eth, data, txType)
			} else {
				addr, err := types.ParseEthAddress(destination, chain)
				if err != nil {
					return err
				}

				msg = types.NewMsgRawTx(cliCtx.GetFromAddress(), &hash, eth, data, addr, txType)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	addDestinationFlag(cmd, &destination)
	addMasterKeyFlag(cmd, &useMasterKey)
	return cmd
}

func GetCmdEther(chain types.Chain, cdc *codec.Codec) *cobra.Command {

	var useMasterKey bool
	var destination string

	cmd := &cobra.Command{
		Use:   "ether [sourceTxId] [amount] [-d <destination> | -m]",
		Short: "ether transfer transaction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			hash := common.HexToHash(args[0])

			eth, txType, err := parseValue(args[1])
			if err != nil {
				return err
			}

			if txType != types.TypeETH {

				return fmt.Errorf("amount given must be a unit of ether")
			}

			if (destination == "" && !useMasterKey) || (destination != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the destination or to use the master key, not both\"")
			}

			var msg sdk.Msg
			if useMasterKey {
				msg = types.NewMsgRawTxForNextMasterKey(cliCtx.GetFromAddress(), &hash, eth, make([]byte, 0), txType)
			} else {
				addr, err := types.ParseEthAddress(destination, chain)
				if err != nil {
					return err
				}

				msg = types.NewMsgRawTx(cliCtx.GetFromAddress(), &hash, eth, make([]byte, 0), addr, txType)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	addDestinationFlag(cmd, &destination)
	addMasterKeyFlag(cmd, &useMasterKey)
	return cmd
}

func GetCmdInstallSC(cdc *codec.Codec) *cobra.Command {

	return &cobra.Command{

		Use:   "installSC [contract ID] [file path] ",
		Short: "Install an ethereum smart contract in Axelar",

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			content, err := ioutil.ReadFile(args[1])

			if err != nil {
				return err
			}

			byteCode := common.FromHex(strings.TrimSuffix(string(content), "\n"))

			// if this conversion fails, it may mean that it is already a binary file and not an hex string
			if err != nil {

				byteCode = content
			}

			msg := types.NewMsgInstallSC(cliCtx.GetFromAddress(), args[0], byteCode)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})

		},
	}
}

func GetCmdVerifyTx(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [txId] [destination] [amount] ",
		Short: "Verify an Ethereum transaction",
		Long: fmt.Sprintf(
			"Verify that a transaction happened on the Ethereum chain so it can be processed on axelar. "+
				"Choose %s, %s, %s, %s or %s for the chain. Accepted denominations (case-insensitive): %s/%s, %s, %s. "+
				"Example: verifyTx f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16 "+
				"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq 0.13eth",
			types.Mainnet, types.Ropsten, types.Rinkeby, types.Kovan, types.Goerli,
			eth, ethereum, wei, gwei),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			hash := common.HexToHash(args[0])

			addr, err := types.ParseEthAddress(args[1], chain)
			if err != nil {
				return err
			}

			amount, txType, err := parseValue(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), &hash, addr, amount, txType)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func addMasterKeyFlag(cmd *cobra.Command, useMasterKey *bool) {
	cmd.Flags().BoolVarP(useMasterKey, masterKeyFlag, "m", false, "Use the current master key instead of a specific key")
}

func addDestinationFlag(cmd *cobra.Command, destination *string) {
	cmd.Flags().StringVarP(destination, "destination", "d", "", "Set the destination address for the transaction")
}

func prepare(reader io.Reader, cdc *codec.Codec) (context.CLIContext, authTypes.TxBuilder) {
	cliCtx := context.NewCLIContext().WithCodec(cdc)
	inBuf := bufio.NewReader(reader)
	txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
	return cliCtx, txBldr
}

func parseValue(rawValue string) (value big.Int, txType types.TXType, err error) {

	value = *big.NewInt(0)
	err = nil
	txType = types.TypeERC20mint

	// if the function is given a basic integer value without units,
	// it will assume is raw representation of a ERC20 token
	if v, ok := big.NewInt(0).SetString(rawValue, 10); ok {

		return *v, txType, nil

	}

	var coin sdk.DecCoin
	coin, err = sdk.ParseDecCoin(rawValue)
	if err != nil {
		return value, txType, fmt.Errorf("could not parse coin string")
	}

	val := big.NewInt(coin.Amount.Int64())

	switch coin.Denom {
	case wei:

		if !coin.Amount.IsInteger() {
			err = fmt.Errorf("wei must be an integer value")
			break
		}

		value = *val
		txType = types.TypeETH

	case eth, ethereum:

		value = *(new(big.Int).Mul(val, big.NewInt(params.Ether)))
		txType = types.TypeETH

	case gwei:
		value = *(new(big.Int).Mul(val, big.NewInt(params.GWei)))
		txType = types.TypeETH

	case btc, bitcoin:
		var i, e = big.NewInt(10), big.NewInt(btcDecs)
		i.Exp(i, e, nil)
		value = *(new(big.Int).Mul(val, i))

	default:
		err = fmt.Errorf("choose a correct denomination: %s (%s), %s, %s", eth, ethereum, wei, gwei)
	}

	return
}

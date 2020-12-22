package cli

import (
	"bufio"
	"fmt"
	"io"
	"math/big"

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

	mainnetCmd := makeCommand(types.Mainnet)
	mainnetEtherCmd := makeCommand("rawTx")
	mainnetEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdRawTx(types.Chain(types.Mainnet), types.TypeETH, cdc),
			GetCmdRawTx(types.Chain(types.Mainnet), types.TypeERC20mint, cdc))...)
	mainnetCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Mainnet), cdc), mainnetEtherCmd)...)

	ropstenCmd := makeCommand(types.Ropsten)
	ropstenEtherCmd := makeCommand("rawTx")
	ropstenEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdRawTx(types.Chain(types.Ropsten), types.TypeETH, cdc),
			GetCmdRawTx(types.Chain(types.Ropsten), types.TypeERC20mint, cdc))...)
	ropstenCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Ropsten), cdc), ropstenCmd)...)

	kovanCmd := makeCommand(types.Kovan)
	kovanEtherCmd := makeCommand("rawTx")
	kovanEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdRawTx(types.Chain(types.Kovan), types.TypeETH, cdc),
			GetCmdRawTx(types.Chain(types.Kovan), types.TypeERC20mint, cdc))...)
	kovanCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Kovan), cdc), kovanEtherCmd)...)

	rinkebyCmd := makeCommand(types.Rinkeby)
	rinkebyEtherCmd := makeCommand("rawTx")
	rinkebyEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdRawTx(types.Chain(types.Rinkeby), types.TypeETH, cdc),
			GetCmdRawTx(types.Chain(types.Rinkeby), types.TypeERC20mint, cdc))...)
	rinkebyCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Rinkeby), cdc), rinkebyEtherCmd)...)

	goerliCmd := makeCommand(types.Goerli)
	goerliEtherCmd := makeCommand("rawTx")
	goerliEtherCmd.AddCommand(
		flags.PostCommands(
			GetCmdRawTx(types.Chain(types.Goerli), types.TypeETH, cdc),
			GetCmdRawTx(types.Chain(types.Goerli), types.TypeERC20mint, cdc))...)
	goerliCmd.AddCommand(
		flags.PostCommands(
			GetCmdVerifyTx(types.Chain(types.Goerli), cdc), goerliEtherCmd)...)

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

func GetCmdRawTx(chain types.Chain, subCmd types.TXType, cdc *codec.Codec) *cobra.Command {

	return &cobra.Command{
		Use:   fmt.Sprintf("%s [sourceTxId] [amount] [destination]", subCmd),
		Short: "Generate raw transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			hash := common.HexToHash(args[0])

			eth, txType, err := parseValue(args[1])
			if err != nil {
				return err
			}

			if txType != subCmd {

				return fmt.Errorf("amount must be a unit of %s", subCmd)
			}

			//TODO: Add parameters to specify a key other than the master key
			addr, err := types.ParseEthAddress(args[2], chain)
			if err != nil {
				return err
			}

			msg := types.NewMsgRawTx(cliCtx.GetFromAddress(), &hash, eth, addr, txType)

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

package cli

import (
	"bufio"
	"fmt"
	"io"
	"math/big"

	"github.com/axelarnetwork/axelar-core/x/eth_bridge/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

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
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "ethereum",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	return btcTxCmd
}

func GetCmdVerifyTx(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [txId] [voutIdx] [destination] [amount] ",
		Short: "Verify a Bitcoin transaction",
		Long: fmt.Sprintf(
			"Verify that a transaction happened on the Ethereum chain so it can be processed on axelar. "+
				"Choose %s, %s, %s, %s or %s for the chain. Accepted denominations (case-insensitive): %s/%s, %s, %s. "+
				"Example: verifyTx f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16 "+
				"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq 0.13btc",
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

			amount, err := parseEth(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), &hash, addr, amount)
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

func parseEth(rawCoin string) (value big.Int, err error) {

	value = *big.NewInt(0)
	err = nil

	var coin sdk.DecCoin
	coin, err = sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return value, fmt.Errorf("could not parse coin string")
	}

	val := big.NewInt(coin.Amount.Int64())

	switch coin.Denom {
	case wei:

		if !coin.Amount.IsInteger() {
			err = fmt.Errorf("wei must be an integer value")
			break
		}

		value = *val

	case eth, ethereum:

		value = *(new(big.Int).Mul(val, big.NewInt(params.Ether)))

	case gwei:
		value = *(new(big.Int).Mul(val, big.NewInt(params.Ether)))

	default:
		err = fmt.Errorf("choose a correct denomination: %s (%s), %s, %s", eth, ethereum, wei, gwei)
	}

	return
}

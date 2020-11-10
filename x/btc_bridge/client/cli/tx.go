package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	btcTxCmd.AddCommand(flags.PostCommands(
		GetCmdTrackAddress(cdc),
		GetCmdTrackAddressFromPubKey(cdc),
		GetCmdVerifyTx(cdc),
		GetCmdWithdraw(cdc),
		GetCmdGenerateRawTx(cdc),
	)...)

	return btcTxCmd
}

func GetCmdTrackAddress(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trackAddress [address]",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			msg := types.NewMsgTrackAddress(cliCtx.GetFromAddress(), args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdTrackAddressFromPubKey(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trackAddressFromPubKey [chain] [keyId]",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Long: fmt.Sprintf("Make the axelar network aware of a specific address on Bitcoin. Choose \"%s\" or \"%s\" for the chain.",
			chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			msg := types.NewMsgTrackAddressFromPubKey(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVerifyTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [chain] [txId] [destination] [amount] [opt. voutIdx]",
		Short: "Verify a Bitcoin transaction",
		Long: "Verify that a transaction happened on the Bitcoin chain so it can be processed on axelar." +
			"The parameter [voutIdx] is optional. Accepted denominations (case-insensitive): satoshi (sat), bitcoin (btc)",
		Args: cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[1])
			if err != nil {
				return err
			}

			_, err = parseAddress(args[0], args[2])
			if err != nil {
				return err
			}

			amount, err := types.ParseBtc(args[3])
			if err != nil {
				return err
			}

			var voutIdx uint32 = 0
			if len(args) == 5 {
				n, err := strconv.ParseUint(args[4], 10, 32)
				if err != nil {
					return sdkerrors.Wrap(err, "could not parse voutIdx")
				}
				voutIdx = uint32(n)
			}

			msg := types.NewMsgVerifyTx(cliCtx.GetFromAddress(), args[0], hash, voutIdx, args[2], amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [sourceTxId] [sigId] [keyId]",
		Short: "Withdraw funds from an Axelar address",
		Long: `Withdraw funds from an Axelar address according to a previously signed raw transaction. 
Ensure the axelar address is being tracked and the transaction signed first`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			msg := types.NewMsgWithdraw(cliCtx.GetFromAddress(), args[0], args[1], args[2])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdGenerateRawTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawTx [chain] [sourceTxId] [amount] [destination]",
		Short: "Generate raw transaction",
		Long:  `Generate raw transaction that can be used to spend the [amount] from the source transaction to the [destination]`,
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := prepare(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[1])
			if err != nil {
				return err
			}

			btc, err := types.ParseBtc(args[2])
			if err != nil {
				return err
			}

			_, err = parseAddress(args[0], args[3])

			msg := types.NewMsgRawTx(cliCtx.GetFromAddress(), args[0], hash, btc, args[3])
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

func parseHash(txId string) (*chainhash.Hash, error) {
	hash, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}
	return hash, nil
}

func parseAddress(chain, address string) (addr btcutil.Address, err error) {
	switch chain {
	case chaincfg.MainNetParams.Name:
		if addr, err = btcutil.DecodeAddress(address, &chaincfg.MainNetParams); err != nil {
			return nil, sdkerrors.Wrap(err, "could not decode destination address")
		}
	case chaincfg.TestNet3Params.Name:
		if addr, err = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params); err != nil {
			return nil, sdkerrors.Wrap(err, "could not decode destination address")
		}
	default:
		return nil, fmt.Errorf(
			"missing chain name, choose %s or %s",
			chaincfg.MainNetParams.Name,
			chaincfg.TestNet3Params.Name,
		)
	}
	return addr, nil
}

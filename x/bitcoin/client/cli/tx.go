package cli

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	cliUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	sat     = "sat"
	satoshi = "satoshi"
	btc     = "btc"
	bitcoin = "bitcoin"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	btcTxCmd := &cobra.Command{
		Use:                        "bitcoin",
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	networks := []chaincfg.Params{chaincfg.MainNetParams, chaincfg.TestNet3Params, chaincfg.RegressionNetParams}

	for _, network := range networks {

		cmd := &cobra.Command{
			Use:                        network.Name,
			Short:                      fmt.Sprintf("%s transactions subcommands", network.Name),
			SuggestionsMinimumDistance: 2,
			RunE:                       client.ValidateCmd,
		}

		addSubCommands(cmd, types.Chain(network.Name), cdc)

		btcTxCmd.AddCommand(cmd)
	}

	return btcTxCmd
}

func addSubCommands(command *cobra.Command, chain types.Chain, cdc *codec.Codec) {
	cmds := append([]*cobra.Command{GetCmdTrack(chain, cdc)},
		flags.PostCommands(
			GetCmdVerifyTx(chain, cdc),
			GetCmdRawTx(chain, cdc),
			GetCmdSend(cdc))...)
	command.AddCommand(cmds...)
}

func GetCmdTrack(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	trackCmd := &cobra.Command{
		Use:   "track",
		Short: "Bitcoin address or public key tracking subcommand",
		RunE:  client.ValidateCmd,
	}

	trackCmd.AddCommand(flags.PostCommands(getCmdTrackAddress(chain, cdc), getCmdTrackPubKey(chain, cdc))...)
	return trackCmd
}

func getCmdTrackAddress(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	var rescan bool
	addrCmd := &cobra.Command{
		Use:   "address [address]",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Long:  "Make the axelar network aware of a specific address on Bitcoin. Use --rescan to rescan the entire Bitcoin history for past transactions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			addr, err := types.ParseBtcAddress(args[0], chain)
			if err != nil {
				return nil
			}

			msg := types.NewMsgTrackAddress(cliCtx.GetFromAddress(), addr, rescan)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	addRescanFlag(addrCmd, &rescan)
	return addrCmd
}

func getCmdTrackPubKey(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	var rescan bool
	var useMasterKey bool
	var keyID string
	pubKeyCmd := &cobra.Command{
		Use:   "pubKey [--key-id <key ID> | -m]",
		Short: "Make the axelar network aware of a specific address on Bitcoin derived from a public key",
		Long: "Make the axelar network aware of a specific address on Bitcoin derived from a public key. " +
			"Either specify the key ID associated with a previously completed keygen round or use the current master key. " +
			"Use --rescan|-r to rescan the entire Bitcoin history for past transactions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			var msg sdk.Msg
			if (useMasterKey && keyID != "") || (!useMasterKey && keyID == "") {
				return fmt.Errorf("either set the flag to use a key ID or to use the master key, not both")
			}
			if useMasterKey {
				msg = types.NewMsgTrackPubKeyWithMasterKey(cliCtx.GetFromAddress(), chain, rescan)
			} else {
				msg = types.NewMsgTrackPubKey(cliCtx.GetFromAddress(), chain, keyID, rescan)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addKeyIdFlag(pubKeyCmd, &keyID)
	addMasterKeyFlag(pubKeyCmd, &useMasterKey)
	addRescanFlag(pubKeyCmd, &rescan)
	return pubKeyCmd
}

func GetCmdVerifyTx(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	var useCurrentMasterKey bool
	var useNextMasterKey bool
	var destination string
	verifyCmd := &cobra.Command{
		Use:   "verifyTx [txId] [voutIdx] [amount] [-d <destination> | --curr-mk | --next-mk]",
		Short: "Verify a Bitcoin transaction",
		Long: fmt.Sprintf(
			"Verify that a transaction happened on the Bitcoin chain so it can be processed on axelar. "+
				"Choose %s or %s for the chain. Accepted denominations (case-insensitive): %s/%s, %s/%s. "+
				"Select the index of the transaction output as voutIdx.\n"+
				"Example: verifyTx f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16 1 "+
				"0.13btc -d bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name, satoshi, sat, bitcoin, btc),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[0])
			if err != nil {
				return err
			}

			amount, err := parseBtc(args[2])
			if err != nil {
				return err
			}

			voutIdx, err := parseVoutIdx(err, args[1])
			if err != nil {
				return err
			}

			var msg sdk.Msg
			if useCurrentMasterKey {
				msg = types.NewMsgVerifyTxForMasterKey(cliCtx.GetFromAddress(), hash, voutIdx, amount, types.ModeCurrentMasterKey, chain)
			} else if useNextMasterKey {
				msg = types.NewMsgVerifyTxForMasterKey(cliCtx.GetFromAddress(), hash, voutIdx, amount, types.ModeNextMasterKey, chain)
			} else {
				addr, err := types.ParseBtcAddress(destination, chain)
				if err != nil {
					return err
				}

				msg = types.NewMsgVerifyTx(cliCtx.GetFromAddress(), hash, voutIdx, addr, amount)
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addDestinationFlag(verifyCmd, &destination)
	verifyCmd.Flags().BoolVar(&useCurrentMasterKey, "curr-mk", false, "Use the current master key instead of a specific key")
	verifyCmd.Flags().BoolVar(&useNextMasterKey, "next-mk", false, "Use the current master key instead of a specific key")
	return verifyCmd
}

func GetCmdRawTx(chain types.Chain, cdc *codec.Codec) *cobra.Command {
	var useMasterKey bool
	var destination string
	rawTxCmd := &cobra.Command{
		Use:   "rawTx [sourceTxId] [amount] [-d <destination> | -m]",
		Short: "Generate raw transaction",
		Long: "Generate raw transaction that can be used to spend the [amount] from the source transaction to the destination (specific address or next master key). " +
			"The difference between the source transaction output amount and the given [amount] becomes the transaction fee",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[0])
			if err != nil {
				return err
			}

			btc, err := parseBtc(args[1])
			if err != nil {
				return err
			}

			if (destination == "" && !useMasterKey) || (destination != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the destination or to use the master key, not both\"")
			}

			var msg sdk.Msg
			if useMasterKey {
				msg = types.NewMsgRawTxForNextMasterKey(cliCtx.GetFromAddress(), chain, hash, btc)
			} else {
				addr, err := types.ParseBtcAddress(destination, chain)
				if err != nil {
					return err
				}

				msg = types.NewMsgRawTx(cliCtx.GetFromAddress(), hash, btc, addr)
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addDestinationFlag(rawTxCmd, &destination)
	addMasterKeyFlag(rawTxCmd, &useMasterKey)
	return rawTxCmd
}

func GetCmdSend(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "send [sourceTxId] [sigId]",
		Short: "Withdraw funds from an axelar address",
		Long: "Withdraw funds from an axelar address according to a previously signed raw transaction. " +
			"Ensure the axelar address is being tracked and the transaction signed first.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := cliUtils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgSendTx(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func parseHash(txId string) (*chainhash.Hash, error) {
	hash, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}
	return hash, nil
}

func parseVoutIdx(err error, voutIdx string) (uint32, error) {
	n, err := strconv.ParseUint(voutIdx, 10, 32)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not parse voutIdx")
	}
	return uint32(n), nil
}

func parseBtc(rawCoin string) (btcutil.Amount, error) {
	var coin sdk.DecCoin
	coin, err := sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return 0, fmt.Errorf("could not parse coin string")
	}

	switch coin.Denom {
	case sat, satoshi:
		if !coin.Amount.IsInteger() {
			return 0, fmt.Errorf("satoshi must be an integer value")
		}
		return btcutil.Amount(coin.Amount.Int64()), nil
	case btc, bitcoin:
		// sdk.Coin does not reduce precision, even if all decimal places are 0,
		// so need to call RoundInt64 to return the correct value
		return btcutil.Amount(coin.Amount.MulInt64(btcutil.SatoshiPerBitcoin).RoundInt64()), nil
	default:
		return 0, fmt.Errorf("choose a correct denomination: %s (%s), %s (%s)", satoshi, sat, bitcoin, btc)
	}
}

func addMasterKeyFlag(cmd *cobra.Command, useMasterKey *bool) {
	cmd.Flags().BoolVarP(useMasterKey, "master-key", "m", false, "Use the current master key instead of a specific key")
}

func addDestinationFlag(cmd *cobra.Command, destination *string) {
	cmd.Flags().StringVarP(destination, "destination", "d", "", "Set the destination address for the transaction")
}

func addRescanFlag(cmd *cobra.Command, rescan *bool) {
	cmd.Flags().BoolVarP(rescan, "rescan", "r", false,
		"Rescan the entire Bitcoin blockchain for previous transactions to this address")
}

func addKeyIdFlag(pubKeyCmd *cobra.Command, keyID *string) {
	pubKeyCmd.Flags().StringVarP(keyID, "key-id", "k", "", "Specify the ID of the key to use")
}

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
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
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

		addSubCommands(cmd, types.Network(network.Name), cdc)

		btcTxCmd.AddCommand(cmd)
	}

	return btcTxCmd
}

func addSubCommands(command *cobra.Command, network types.Network, cdc *codec.Codec) {
	cmds := append([]*cobra.Command{GetCmdTrack(network, cdc)},
		flags.PostCommands(
			GetCmdVerifyTx(network, cdc),
			GetCmdRawTx(network, cdc),
			GetCmdSend(cdc))...)
	GetCmdTransfer(network, cdc)
	command.AddCommand(cmds...)
}

func GetCmdTrack(network types.Network, cdc *codec.Codec) *cobra.Command {
	trackCmd := &cobra.Command{
		Use:   "track",
		Short: "Bitcoin address or public key tracking subcommand",
		RunE:  client.ValidateCmd,
	}

	trackCmd.AddCommand(flags.PostCommands(getCmdTrackAddress(network, cdc), getCmdTrackPubKey(network, cdc))...)
	return trackCmd
}

func getCmdTrackAddress(network types.Network, cdc *codec.Codec) *cobra.Command {
	var rescan bool
	addrCmd := &cobra.Command{
		Use:   "address [address]",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Long:  "Make the axelar network aware of a specific address on Bitcoin. Use --rescan to rescan the entire Bitcoin history for past transactions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			addr, err := types.ParseBtcAddress(args[0], network)
			if err != nil {
				return nil
			}

			msg := types.NewMsgTrackAddress(cliCtx.GetFromAddress(), addr, rescan)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	addRescanFlag(addrCmd, &rescan)
	return addrCmd
}

func getCmdTrackPubKey(network types.Network, cdc *codec.Codec) *cobra.Command {
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

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			var msg sdk.Msg
			if (useMasterKey && keyID != "") || (!useMasterKey && keyID == "") {
				return fmt.Errorf("either set the flag to use a key ID or to use the master key, not both")
			}
			if useMasterKey {
				msg = types.NewMsgTrackPubKeyWithMasterKey(cliCtx.GetFromAddress(), network, rescan)
			} else {
				msg = types.NewMsgTrackPubKey(cliCtx.GetFromAddress(), network, keyID, rescan)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addKeyIdFlag(pubKeyCmd, &keyID)
	addMasterKeyFlag(pubKeyCmd, &useMasterKey)
	addRescanFlag(pubKeyCmd, &rescan)
	return pubKeyCmd
}

func GetCmdVerifyTx(network types.Network, cdc *codec.Codec) *cobra.Command {
	var toCurrentMasterKey bool
	var toNextMasterKey bool
	var recipient string
	var sender string
	var fromCurrentMasterKey bool
	verifyCmd := &cobra.Command{
		Use:   "verifyTx [txId] [voutIdx] [amount] [-s <sender> | --from-curr-mk ] [-r <recipient> | --to-curr-mk | --to-next-mk ]",
		Short: "Verify a Bitcoin transaction",
		Long: fmt.Sprintf(
			"Verify that a transaction happened on the Bitcoin network so it can be processed on axelar. "+
				"Choose %s or %s for the network. Accepted denominations (case-insensitive): %s/%s, %s/%s. "+
				"Select the index of the transaction output as voutIdx.\n"+
				"Example: verifyTx 3PtPE3yZAnGoqKsN23gWVpLMYQ4b7a4PxK 1 "+
				"0.13btc -d bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name, denom.Satoshi, denom.Sat, denom.Bitcoin, denom.Btc),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[0])
			if err != nil {
				return err
			}

			sat, err := denom.ParseSatoshi(args[3])
			if err != nil {
				return err
			}
			amount := btcutil.Amount(sat.Amount.Int64())

			voutIdx, err := parseVoutIdx(err, args[2])
			if err != nil {
				return err
			}

			var msg sdk.Msg
			if toCurrentMasterKey {
				sender, err := types.ParseBtcAddress(sender, network)
				if err != nil {
					return err
				}
				msg = types.NewMsgVerifyTxToCurrentMasterKey(cliCtx.GetFromAddress(), hash, voutIdx, sender, amount, network)
			} else if fromCurrentMasterKey && toNextMasterKey {
				msg = types.NewMsgVerifyTxForNextMasterKey(cliCtx.GetFromAddress(), hash, voutIdx, amount, network)
			} else {
				sender, err := types.ParseBtcAddress(sender, network)
				recipient, err := types.ParseBtcAddress(recipient, network)
				if err != nil {
					return err
				}

				msg = types.NewMsgVerifyTx(cliCtx.GetFromAddress(), hash, voutIdx, sender, recipient, amount)
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addRecipientFlag(verifyCmd, &recipient)
	verifyCmd.Flags().StringVarP(&sender, "sender", "s", "", "Address of the sender")
	verifyCmd.Flags().BoolVar(&fromCurrentMasterKey, "from-curr-mk", false, "Send to current master key instead of a specific key")
	verifyCmd.Flags().BoolVar(&toCurrentMasterKey, "to-curr-mk", false, "Send to current master key instead of a specific key")
	verifyCmd.Flags().BoolVar(&toNextMasterKey, "to-next-mk", false, "Send to next master key instead of a specific key")
	return verifyCmd
}

func GetCmdRawTx(network types.Network, cdc *codec.Codec) *cobra.Command {
	var useMasterKey bool
	var recipient string
	rawTxCmd := &cobra.Command{
		Use:   "rawTx [sourceTxId] [amount] [-r <recipient> | -m]",
		Short: "Generate raw transaction",
		Long: "Generate raw transaction that can be used to spend the [amount] from the source transaction to the recipient (specific address or next master key). " +
			"The difference between the source transaction output amount and the given [amount] becomes the transaction fee",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			hash, err := parseHash(args[0])
			if err != nil {
				return err
			}

			sat, err := denom.ParseSatoshi(args[2])
			if err != nil {
				return err
			}
			amount := btcutil.Amount(sat.Amount.Int64())

			if (recipient == "" && !useMasterKey) || (recipient != "" && useMasterKey) {
				return fmt.Errorf("either set the flag to set the recipient or to use the master key, not both\"")
			}

			var msg sdk.Msg
			if useMasterKey {
				msg = types.NewMsgRawTxForNextMasterKey(cliCtx.GetFromAddress(), network, hash, amount)
			} else {
				addr, err := types.ParseBtcAddress(recipient, network)
				if err != nil {
					return err
				}

				msg = types.NewMsgRawTx(cliCtx.GetFromAddress(), hash, amount, addr)
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	addRecipientFlag(rawTxCmd, &recipient)
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

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			msg := types.NewMsgSendTx(cliCtx.GetFromAddress(), args[0], args[1])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdTransfer(network types.Network, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [btcAddress] [recipient chain] [recipient address]",
		Short: "Connect a Bitcoin address to a recipient address on a recipient chain for a future transfer.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			btcAddr, err := types.ParseBtcAddress(args[0], network)
			if err != nil {
				return err
			}

			msg := types.NewMsgTransfer(cliCtx.GetFromAddress(), btcAddr, exported.CrossChainAddress{
				Chain:   exported.ChainFromString(args[1]),
				Address: args[1],
			})
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
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

func addMasterKeyFlag(cmd *cobra.Command, useMasterKey *bool) {
	cmd.Flags().BoolVarP(useMasterKey, "master-key", "m", false, "Use the current master key instead of a specific key")
}

func addRecipientFlag(cmd *cobra.Command, recipient *string) {
	cmd.Flags().StringVarP(recipient, "recipient", "r", "", "Set the recipient address for the transaction")
}

func addRescanFlag(cmd *cobra.Command, rescan *bool) {
	cmd.Flags().BoolVarP(rescan, "rescan", "r", false,
		"Rescan the entire Bitcoin blockchain for previous transactions to this address")
}

func addKeyIdFlag(pubKeyCmd *cobra.Command, keyID *string) {
	pubKeyCmd.Flags().StringVarP(keyID, "key-id", "k", "", "Specify the ID of the key to use")
}

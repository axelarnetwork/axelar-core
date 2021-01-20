package cli

import (
	"fmt"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authUtils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
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

	addTxSubCommands(btcTxCmd, cdc)

	return btcTxCmd
}

func addTxSubCommands(command *cobra.Command, cdc *codec.Codec) {
	cmds := append([]*cobra.Command{GetCmdTrack(cdc)},
		flags.PostCommands(
			GetCmdVerifyTx(cdc),
			GetCmdSignRawTx(cdc),
			GetCmdLink(cdc),
		)...)

	command.AddCommand(cmds...)
}

// GetCmdTrack returns the address tracking command
func GetCmdTrack(cdc *codec.Codec) *cobra.Command {
	trackCmd := &cobra.Command{
		Use:   "track",
		Short: "Bitcoin address or public key tracking subcommand",
		RunE:  client.ValidateCmd,
	}

	trackCmd.AddCommand(flags.PostCommands(getCmdTrackAddress(cdc), getCmdTrackPubKey(cdc))...)
	return trackCmd
}

func getCmdTrackAddress(cdc *codec.Codec) *cobra.Command {
	var rescan bool
	addrCmd := &cobra.Command{
		Use:   "address [address]",
		Short: "Make the axelar network aware of a specific address on Bitcoin",
		Long:  "Make the axelar network aware of a specific address on Bitcoin. Use --rescan to rescan the entire Bitcoin history for past transactions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			addr, err := btcutil.DecodeAddress(args[0], &chaincfg.MainNetParams)
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

func getCmdTrackPubKey(cdc *codec.Codec) *cobra.Command {
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
				msg = types.NewMsgTrackPubKeyWithMasterKey(cliCtx.GetFromAddress(), rescan)
			} else {
				msg = types.NewMsgTrackPubKey(cliCtx.GetFromAddress(), keyID, rescan)
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

// GetCmdVerifyTx returns the transaction verification command
func GetCmdVerifyTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verifyTx [txInfo json]",
		Short: "Verify a Bitcoin transaction",
		Long: fmt.Sprintf(
			"Verify that a transaction happened on the Bitcoin network so it can be processed on axelar. "+
				"Get the json string by using the %s query", keeper.QueryOutInfo),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			var out types.OutPointInfo
			cliCtx.Codec.MustUnmarshalJSON([]byte(args[0]), &out)

			msg := types.MsgVerifyTx{Sender: cliCtx.GetFromAddress(), OutPointInfo: out}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSignRawTx returns the command to sign a raw Bitcoin transaction
func GetCmdSignRawTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "signTx [txID] [tx json]",
		Short: "Sign raw spending transaction with utxo of [txID]",
		Long:  fmt.Sprintf("Sign raw transaction. Get raw transaction by querying %s", keeper.QueryRawTx),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)
			var tx *wire.MsgTx
			types.ModuleCdc.MustUnmarshalJSON([]byte(args[1]), &tx)

			msg := types.NewMsgSignTx(cliCtx.GetFromAddress(), args[0], tx)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdLink links a cross chain address to a bitcoin address created by Axelar
func GetCmdLink(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "link [chain] [address]",
		Short: "Link a cross chain address to a bitcoin address created by Axelar",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx, txBldr := utils.PrepareCli(cmd.InOrStdin(), cdc)

			chain := balance.ChainFromString(args[0])
			address := balance.CrossChainAddress{Chain: chain, Address: args[1]}

			if err := address.Validate(); err != nil {
				return err
			}

			msg := types.MsgLink{Sender: cliCtx.GetFromAddress(), Recipient: address}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authUtils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
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

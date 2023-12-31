package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		getCmdStartKeygen(),
		getCmdRotateKey(),
		getCmdKeygen(),
	)

	return txCmd
}

func getCmdStartKeygen() *cobra.Command {
	cmd := getCmdStart()
	cmd.Use = "start-keygen"
	cmd.Deprecated = "use \"keygen start\" instead"

	return cmd
}

func getCmdRotateKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate [chain] [keyID]",
		Short: "Rotate the given chain to the given key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chain := nexus.ChainName(args[0])
			keyID := exported.KeyID(args[1])

			msg := types.NewRotateKeyRequest(clientCtx.FromAddress, chain, keyID)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getCmdKeygen() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "keygen",
		Short:                      "sub-commands for keygen",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		TraverseChildren:           true,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		getCmdStart(),
		getCmdOptOut(),
		getCmdOptIn(),
	)

	return cmd
}

func getCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Initiate key generation protocol",
		Args:  cobra.NoArgs,
	}

	keyID := cmd.Flags().String("id", "", "unique ID for new key (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic("id flag not set")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		msg := types.NewStartKeygenRequest(cliCtx.FromAddress, exported.KeyID(*keyID))

		return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getCmdOptOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opt-out",
		Short: "Opt the sender out of future keygens. Sender should be a proxy address for a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.KeygenOptOutRequest{Sender: clientCtx.FromAddress}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getCmdOptIn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opt-in",
		Short: "Opt the sender into future keygens. Sender should be a proxy address for a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.KeygenOptInRequest{Sender: clientCtx.FromAddress}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

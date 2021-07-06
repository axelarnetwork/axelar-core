package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	tssTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	tssTxCmd.AddCommand(
		getCmdKeygenStart(),
		getCmdRotateKey(),
	)

	return tssTxCmd
}

func getCmdKeygenStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-keygen",
		Short: "Initiate threshold key generation protocol",
		Args:  cobra.NoArgs,
	}

	newKeyID := cmd.Flags().String("id", "", "unique ID for new key (required)")
	if cmd.MarkFlagRequired("id") != nil {
		panic("flag not set")
	}

	subsetSize := cmd.Flags().Int64("subset-size", 0, "number of top validators to participate in the key generation")
	keyShareDistributionPolicy := cmd.Flags().String(
		"key-share-distribution-policy",
		exported.WeightedByStake.SimpleString(),
		fmt.Sprintf("policy for distributing key shares; available options: %s, %s", exported.WeightedByStake.SimpleString(), exported.OnePerValidator.SimpleString()),
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		keyShareDistributionPolicy, err := exported.KeyShareDistributionPolicyFromSimpleStr(*keyShareDistributionPolicy)
		if err != nil {
			return err
		}

		msg := types.NewStartKeygenRequest(clientCtx.FromAddress, *newKeyID, *subsetSize, keyShareDistributionPolicy)
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getCmdRotateKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate [chain] [role] [keyID]",
		Short: "Rotate the given chain from the old key to the given key",
		Args:  cobra.ExactArgs(3),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		chain := args[0]
		keyRole, err := exported.KeyRoleFromSimpleStr(args[1])
		if err != nil {
			return err
		}

		msg := types.NewRotateKeyRequest(clientCtx.FromAddress, chain, keyRole, args[2])
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

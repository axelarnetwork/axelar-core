package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

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
		GetCmdRegisterExternalKeys(),
	)

	return tssTxCmd
}

func getCmdKeygenStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-keygen",
		Short: "Initiate threshold key generation protocol",
		Args:  cobra.NoArgs,
	}

	keyID := cmd.Flags().String("id", "", "unique ID for new key (required)")
	if cmd.MarkFlagRequired("id") != nil {
		panic("flag not set")
	}

	keyRoleStr := cmd.Flags().String("key-role", exported.MasterKey.SimpleString(), "role of the key to be generated")

	keyTypeStr := cmd.Flags().String("key-type", "", "type of the key to be generated")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		keyRole, err := exported.KeyRoleFromSimpleStr(*keyRoleStr)
		if err != nil {
			return err
		}

		keyType, err := exported.KeyTypeFromSimpleStr(*keyTypeStr)
		if err != nil {
			return err
		}

		msg := types.NewStartKeygenRequest(clientCtx.FromAddress, *keyID, keyRole, keyType)
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

// GetCmdRegisterExternalKeys returns the cli command to register an external key
func GetCmdRegisterExternalKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-external-keys [chain]",
		Short: "Register the external keys for the given chain",
		Args:  cobra.ExactArgs(1),
	}
	keys := cmd.Flags().StringSlice("key", []string{}, "key ID and public key in the hex format, e.g. [keyID:keyHex]")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		if len(*keys) == 0 {
			return fmt.Errorf("keys are required")
		}

		chain := args[0]
		externalKeys := make([]types.RegisterExternalKeysRequest_ExternalKey, len(*keys))
		for i, key := range *keys {
			keyInfos := strings.Split(key, ":")
			if len(keyInfos) != 2 {
				return fmt.Errorf("key ID and public key hex have to be separated by \":\"")
			}

			keyID := keyInfos[0]
			pubKeyHex := keyInfos[1]

			pubKeyBytes, err := hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			externalKeys[i] = types.RegisterExternalKeysRequest_ExternalKey{ID: exported.KeyID(keyID), PubKey: pubKeyBytes}
		}

		msg := types.NewRegisterExternalKeysRequest(clientCtx.GetFromAddress(), chain, externalKeys...)
		if err := msg.ValidateBasic(); err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

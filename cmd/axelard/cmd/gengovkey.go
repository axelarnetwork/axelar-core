package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissionTypes "github.com/axelarnetwork/axelar-core/x/permission/types"
)

// SetMultisigGovernanceCmd returns set-governance-key cobra Command.
func SetMultisigGovernanceCmd(defaultNodeHome string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-governance-key [threshold] [[pubKey]...]",
		Short: "Set the genesis multisig governance key for the axelar network",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			threshold, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			var pubKeys []crypto.PubKey
			for i := 1; i < len(args); i++ {
				var pk crypto.PubKey
				err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[i]), &pk)
				if err != nil {
					return err
				}

				pubKeys = append(pubKeys, pk)
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			genesisPermission := permissionTypes.GetGenesisStateFromAppState(cdc, appState)

			multisigPubkey := multisig.NewLegacyAminoPubKey(threshold, pubKeys)
			genesisPermission.GovernanceKey = multisigPubkey
			genesisPermission.GovAccounts = []permissionTypes.GovAccount{
				permissionTypes.NewGovAccount(multisigPubkey.Address().Bytes(), exported.ROLE_ACCESS_CONTROL),
			}

			genesisPermissionBz, err := cdc.MarshalJSON(&genesisPermission)
			if err != nil {
				return fmt.Errorf("failed to marshal permission genesis state: %w", err)
			}

			appState[permissionTypes.ModuleName] = genesisPermissionBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	return cmd
}

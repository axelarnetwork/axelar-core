package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
)

const flagLockingPeriod = "locking-period"

// SetGenesisSnapshotCmd returns set-genesis-chain-params cobra Command.
func SetGenesisSnapshotCmd(defaultNodeHome string) *cobra.Command {
	var lockingPeriod time.Duration

	cmd := &cobra.Command{
		Use:   "set-genesis-snapshot",
		Short: "Set the genesis parameters for the snapshot module",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc.(codec.Codec)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			genesisSnapshot := snapshotTypes.GetGenesisStateFromAppState(cdc, appState)

			genesisSnapshot.Params.LockingPeriod = lockingPeriod

			genesisSnapshotBz, err := cdc.MarshalJSON(&genesisSnapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot genesis state: %w", err)
			}

			appState[snapshotTypes.ModuleName] = genesisSnapshotBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().DurationVar(&lockingPeriod, flagLockingPeriod, snapshotTypes.DefaultParams().LockingPeriod, "Locking period for the snapshot module (e.g., \"6h\").")

	return cmd
}

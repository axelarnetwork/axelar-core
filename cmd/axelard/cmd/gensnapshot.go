package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

const flagMinProxyBalance = "min-proxy-balance"

// SetGenesisSnapshotCmd returns set-genesis-chain-params cobra Command.
func SetGenesisSnapshotCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-genesis-snapshot",
		Short: "Set the genesis parameters for the snapshot module",
		Args:  cobra.ExactArgs(0),
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	minProxyBalance := cmd.Flags().Int64(flagMinProxyBalance, snapshotTypes.DefaultParams().MinProxyBalance, "minimum balance required for a proxy address to be registered")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd)
		cdc := clientCtx.Codec

		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		genFile := config.GenesisFile()
		appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
		if err != nil {
			return fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		genesisSnapshot := snapshotTypes.GetGenesisStateFromAppState(cdc, appState)
		genesisSnapshot.Params.MinProxyBalance = *minProxyBalance

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
	}

	return cmd
}

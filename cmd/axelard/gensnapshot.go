package main

import (
	"fmt"
	"time"

	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// SetGenesisSnapshotCmd returns set-genesis-chain-params cobra Command.
func SetGenesisSnapshotCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	var period string

	cmd := &cobra.Command{
		Use:   "set-genesis-snapshot",
		Short: "Set the genesis parameters for the snapshot module",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			genesisSnapshot := snapshotTypes.GetGenesisStateFromAppState(cdc, appState)

			if period != "" {
				period, err := time.ParseDuration(period)
				if err != nil {
					return err
				}
				genesisSnapshot.Params.LockingPeriod = period
			}

			genesisSnapshotBz, err := cdc.MarshalJSON(genesisSnapshot)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot genesis state: %w", err)
			}

			appState[snapshotTypes.ModuleName] = genesisSnapshotBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().StringVar(&period, "locking-period", "", "Locking period for the snapshot module.")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/spf13/cobra"
)

const (
	flagSignedBlocksWindow      = "signed-blocks-window"
	flagMinSignedPerWindow      = "min-signed-per-window"
	flagDowntimeJailDuration    = "downtime-jail-duration"
	flagSlashFractionDoubleSign = "slash-fraction-double-sign"
	flagSlashFractionDowntime   = "slash-fraction-downtime"
)

// SetGenesisSlashingCmd returns set-genesis-chain-params cobra Command.
func SetGenesisSlashingCmd(defaultNodeHome string) *cobra.Command {
	var (
		signedBlocksWindow      uint64
		minSignedPerWindow      string
		downtimeJailDuration    string
		slashFractionDoubleSign string
		slashFractionDowntime   string
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-slashing",
		Short: "Set the genesis parameters for the slashing module",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var genesisSlashing slashingTypes.GenesisState
			if appState[slashingTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[slashingTypes.ModuleName], &genesisSlashing)
			}

			if signedBlocksWindow > 0 {
				genesisSlashing.Params.SignedBlocksWindow = int64(signedBlocksWindow)
			}

			if minSignedPerWindow != "" {
				min, err := sdk.NewDecFromStr(minSignedPerWindow)
				if err != nil {
					return err
				}

				genesisSlashing.Params.MinSignedPerWindow = min
			}

			if downtimeJailDuration != "" {
				duration, err := time.ParseDuration(downtimeJailDuration)
				if err != nil {
					return err
				}
				genesisSlashing.Params.DowntimeJailDuration = duration
			}

			if slashFractionDoubleSign != "" {
				fraction, err := sdk.NewDecFromStr(slashFractionDoubleSign)
				if err != nil {
					return err
				}

				genesisSlashing.Params.SlashFractionDoubleSign = fraction
			}

			if slashFractionDowntime != "" {
				fraction, err := sdk.NewDecFromStr(slashFractionDowntime)
				if err != nil {
					return err
				}

				genesisSlashing.Params.SlashFractionDowntime = fraction
			}

			genesisSlashingBz, err := cdc.MarshalJSON(&genesisSlashing)
			if err != nil {
				return fmt.Errorf("failed to marshal slashing genesis state: %w", err)
			}
			appState[slashingTypes.ModuleName] = genesisSlashingBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().Uint64Var(&signedBlocksWindow, flagSignedBlocksWindow, 0, "Block height window to measure liveness of each validator (e.g., 10000).")
	cmd.Flags().StringVar(&minSignedPerWindow, flagMinSignedPerWindow, "", "Minimum amount of signed blocks per window (e.g., \"0.50\").")
	cmd.Flags().StringVar(&downtimeJailDuration, flagDowntimeJailDuration, "", "Jail duration due to downtime (e.g., \"600s\").")
	cmd.Flags().StringVar(&slashFractionDoubleSign, flagSlashFractionDoubleSign, "", "Slashing fraction due to double signing (e.g., \"0.01\").")
	cmd.Flags().StringVar(&slashFractionDowntime, flagSlashFractionDowntime, "", "Slashing fraction due to downtime (e.g., \"0.0001\").")

	return cmd
}

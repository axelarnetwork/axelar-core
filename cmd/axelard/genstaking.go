package main

import (
	"fmt"
	"time"

	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// SetGenesisStakingCmd returns set-genesis-chain-params cobra Command.
func SetGenesisStakingCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	var period string
	var max uint16

	cmd := &cobra.Command{
		Use:   "set-genesis-staking",
		Short: "Set the genesis parameters for the staking module",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var genesisStaking stakingTypes.GenesisState
			if appState[stakingTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[stakingTypes.ModuleName], &genesisStaking)
			}

			if period != "" {
				period, err := time.ParseDuration(period)
				if err != nil {
					return err
				}
				genesisStaking.Params.UnbondingTime = period
			}

			if max > 0 {
				genesisStaking.Params.MaxValidators = max
			}

			genesisSnapshotBz, err := cdc.MarshalJSON(genesisStaking)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot genesis state: %w", err)
			}

			appState[stakingTypes.ModuleName] = genesisSnapshotBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().StringVar(&period, "locking-period", "", "Time duration of unbonding (e.g., \"6h\").")
	cmd.Flags().Uint16Var(&max, "max-validators", 0, "A positive integer representing the maximum number of validators (max uint16 = 65535)")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

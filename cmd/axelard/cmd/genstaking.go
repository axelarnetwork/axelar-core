package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

const (
	flagUnbondingPeriod = "unbonding-period"
	flagMaxValidators   = "max-validators"
	flagBondDenom       = "bond-denom"
)

// SetGenesisStakingCmd returns set-genesis-chain-params cobra Command.
func SetGenesisStakingCmd(defaultNodeHome string) *cobra.Command {
	var (
		unbond    string
		max       uint32
		bondDenom string
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-staking",
		Short: "Set the genesis parameters for the staking module",
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

			var genesisStaking stakingTypes.GenesisState
			if appState[stakingTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[stakingTypes.ModuleName], &genesisStaking)
			}

			if unbond != "" {
				period, err := time.ParseDuration(unbond)
				if err != nil {
					return err
				}
				genesisStaking.Params.UnbondingTime = period
			}

			if max > 0 {
				genesisStaking.Params.MaxValidators = max
			}

			if bondDenom != "" {
				genesisStaking.Params.BondDenom = bondDenom

				var genesisMint mintTypes.GenesisState
				if appState[mintTypes.ModuleName] != nil {
					cdc.MustUnmarshalJSON(appState[mintTypes.ModuleName], &genesisMint)
				}
				genesisMint.Params.MintDenom = bondDenom
				genesisSnapshotBz, err := cdc.MarshalJSON(&genesisMint)
				if err != nil {
					return fmt.Errorf("failed to marshal snapshot genesis state: %w", err)
				}
				appState[mintTypes.ModuleName] = genesisSnapshotBz
			}

			genesisSnapshotBz, err := cdc.MarshalJSON(&genesisStaking)
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot genesis state: %w", err)
			}
			appState[stakingTypes.ModuleName] = genesisSnapshotBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().StringVar(&unbond, flagUnbondingPeriod, "", "Time duration of unbonding (e.g., \"6h\").")
	cmd.Flags().Uint32Var(&max, flagMaxValidators, 0, "A positive integer representing the maximum number of validators (max uint16 = 65535)")
	cmd.Flags().StringVar(&bondDenom, flagBondDenom, "", "A string representing bondable coin denomination")

	return cmd
}

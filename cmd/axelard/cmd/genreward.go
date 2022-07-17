package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	rewardTypes "github.com/axelarnetwork/axelar-core/x/reward/types"
)

const (
	flagExternalChainVotingInflationRate = "external-chain-voting-inflation-rate"
	flagKeyMgmtRelativeInflationRate     = "key-mgmt-relative-inflation-rate"
)

// SetGenesisRewardCmd returns set-genesis-chain-params cobra Command.
func SetGenesisRewardCmd(defaultNodeHome string) *cobra.Command {
	var (
		externalChainVotingInflationRate string
		keyMgmtRelativeInflationRate     string
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-reward",
		Short: "Set the genesis parameters for the reward module",
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
			genesisReward := rewardTypes.GetGenesisStateFromAppState(cdc, appState)

			if externalChainVotingInflationRate != "" {
				rate, err := sdk.NewDecFromStr(externalChainVotingInflationRate)
				if err != nil {
					return err
				}

				genesisReward.Params.ExternalChainVotingInflationRate = rate
			}

			if keyMgmtRelativeInflationRate != "" {
				rate, err := sdk.NewDecFromStr(keyMgmtRelativeInflationRate)
				if err != nil {
					return err
				}

				genesisReward.Params.KeyMgmtRelativeInflationRate = rate
			}

			genesisRewardBz, err := cdc.MarshalJSON(&genesisReward)
			if err != nil {
				return fmt.Errorf("failed to marshal reward genesis state: %w", err)
			}
			appState[rewardTypes.ModuleName] = genesisRewardBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().StringVar(&externalChainVotingInflationRate, flagExternalChainVotingInflationRate, "", "The fraction of total stake per year that's distributed among external chain voters (e.g., \"0.02\").")
	cmd.Flags().StringVar(&keyMgmtRelativeInflationRate, flagKeyMgmtRelativeInflationRate, "", "The fraction of current inflation rate that's rewarded for participating in TSS (e.g., \"1.00\").")

	return cmd
}

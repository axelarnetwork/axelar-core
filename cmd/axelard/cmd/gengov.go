package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"
)

const (
	flagMinDeposit       = "minimum-deposit"
	flagMaxDepositPeriod = "max-deposit-period"
	flagVotingPeriod     = "voting-period"
)

// SetGenesisGovCmd returns set-genesis-gov cobra Command.
func SetGenesisGovCmd(defaultNodeHome string) *cobra.Command {
	var (
		minDeposit       string
		maxDepositPeriod uint64
		votingPeriod     uint64
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-gov",
		Short: "Set the genesis parameters for the governance module",
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

			var genesisGov govTypes.GenesisState
			if appState[govTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[govTypes.ModuleName], &genesisGov)
			}

			if minDeposit != "" {
				coin, err := sdk.ParseCoinNormalized(minDeposit)
				if err != nil {
					return err
				}
				genesisGov.DepositParams.MinDeposit = sdk.NewCoins(coin)
			}

			if maxDepositPeriod > 0 {
				genesisGov.DepositParams.MaxDepositPeriod = time.Duration(maxDepositPeriod) * time.Nanosecond
			}

			if votingPeriod > 0 {
				genesisGov.VotingParams.VotingPeriod = time.Duration(votingPeriod) * time.Nanosecond
			}

			genesisGovBz, err := cdc.MarshalJSON(&genesisGov)
			if err != nil {
				return fmt.Errorf("failed to marshal gov genesis state: %w", err)
			}
			appState[govTypes.ModuleName] = genesisGovBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().StringVar(&minDeposit, flagMinDeposit, "", "Minimum deposit for a proposal to enter voting period")
	cmd.Flags().Uint64Var(&maxDepositPeriod, flagMaxDepositPeriod, 0, "Maximum period for AXL holders to deposit on a proposal (time ns)")
	cmd.Flags().Uint64Var(&votingPeriod, flagVotingPeriod, 0, "Length of the voting period (time ns)")

	return cmd
}

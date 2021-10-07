package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/axelarnetwork/axelar-core/x/vote"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
)

const (
	flagThreshold = "threshold"
)

// SetGenesisVoteCmd returns set-genesis-chain-params cobra Command.
func SetGenesisVoteCmd(defaultNodeHome string) *cobra.Command {
	var (
		threshold string
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-vote",
		Short: "Set the genesis parameters for the vote module",
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
			genesisVote := vote.GetGenesisStateFromAppState(cdc, appState)

			if threshold != "" {
				threshold, err := parseThreshold(threshold)
				if err != nil {
					return err
				}
				genesisVote.VotingThreshold = threshold
			}

			genesisVoteBz, err := cdc.MarshalJSON(&genesisVote)
			if err != nil {
				return fmt.Errorf("failed to marshal vote genesis state: %w", err)
			}
			appState[voteTypes.ModuleName] = genesisVoteBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().StringVar(&threshold, flagThreshold, "", "The % of stake that is required for a voting poll to conclude (e.g., \"2/3\").")

	return cmd
}

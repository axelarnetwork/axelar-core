package cmd

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/vote"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// SetGenesisVoteCmd returns set-genesis-chain-params cobra Command.
func SetGenesisVoteCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	var threshold string
	var interval int64

	cmd := &cobra.Command{
		Use:   "set-genesis-vote",
		Short: "Set the genesis parameters for the vote module",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
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

			if interval > 0 {
				genesisVote.VotingInterval = interval
			}

			genesisVoteBz, err := cdc.MarshalJSON(genesisVote)
			if err != nil {
				return fmt.Errorf("failed to marshal vote genesis state: %w", err)
			}
			appState[voteTypes.ModuleName] = genesisVoteBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().StringVar(&threshold, "threshold", "", "The % of stake that is required for a voting poll to conclude (e.g., \"2/3\").")
	cmd.Flags().Int64Var(&interval, "interval", 0, "A positive integer representing the number of blocks between tallying votes.")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(cliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

package main

import (
	"fmt"

	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// SetGenesisTSSCmd returns set-genesis-chain-params cobra Command.
func SetGenesisTSSCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	var period int64
	var keygen string
	var corruption string

	cmd := &cobra.Command{
		Use:   "set-genesis-tss",
		Short: "Set the genesis parameters for the tss module",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			genesisTSS := tssTypes.GetGenesisStateFromAppState(cdc, appState)

			if period > 0 {
				genesisTSS.Params.LockingPeriod = period
			}

			if keygen != "" {
				threshold, err := parseThreshold(keygen)
				if err != nil {
					return err
				}
				genesisTSS.Params.MinKeygenThreshold = threshold
			}

			if corruption != "" {
				threshold, err := parseThreshold(corruption)
				if err != nil {
					return err
				}
				genesisTSS.Params.CorruptionThreshold = threshold
			}

			genesisTSSBz, err := cdc.MarshalJSON(genesisTSS)
			if err != nil {
				return fmt.Errorf("failed to marshal tss genesis state: %w", err)
			}

			appState[tssTypes.ModuleName] = genesisTSSBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().Int64Var(&period, "locking-period", 0, "Locking period for the TSS module.")
	cmd.Flags().StringVar(&keygen, "keygen", "", "The minimum % of stake that must be online to authorize generation of a new key in the system.")
	cmd.Flags().StringVar(&corruption, "corruption", "", "The corruption threshold with which Axelar Core will run the keygen protocol.")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

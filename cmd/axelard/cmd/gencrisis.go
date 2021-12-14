package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crisisTypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

const (
	flagConstantFee = "constant-fee"
)

// SetGenesisCrisisCmd returns set-genesis-chain-params cobra Command.
func SetGenesisCrisisCmd(defaultNodeHome string) *cobra.Command {
	var (
		constantFee string
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-crisis",
		Short: "Set the genesis parameters for the crisis module",
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

			var genesisCrisis crisisTypes.GenesisState
			if appState[crisisTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[crisisTypes.ModuleName], &genesisCrisis)
			}

			if constantFee != "" {
				fee, err := sdk.ParseCoinNormalized(constantFee)
				if err != nil {
					return err
				}

				genesisCrisis.ConstantFee = fee
			}

			genesisCrisisBz, err := cdc.MarshalJSON(&genesisCrisis)
			if err != nil {
				return fmt.Errorf("failed to marshal crisis genesis state: %w", err)
			}
			appState[crisisTypes.ModuleName] = genesisCrisisBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().StringVar(&constantFee, flagConstantFee, "", "Transaction fee to initiate a broken invariant check.")

	return cmd
}

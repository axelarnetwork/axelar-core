package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

const (
	flagTxSigLimit = "tx-sig-limit"
)

// SetGenesisAuthCmd returns set-genesis-chain-params cobra Command.
func SetGenesisAuthCmd(defaultNodeHome string) *cobra.Command {
	var (
		txSigLimit uint64
	)

	cmd := &cobra.Command{
		Use:   "set-genesis-auth",
		Short: "Set the genesis parameters for the auth module",
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

			var genesisAuth authTypes.GenesisState
			if appState[authTypes.ModuleName] != nil {
				cdc.MustUnmarshalJSON(appState[authTypes.ModuleName], &genesisAuth)
			}

			if txSigLimit != 0 {
				genesisAuth.Params.TxSigLimit = txSigLimit
			}

			genesisAuthBz, err := cdc.MarshalJSON(&genesisAuth)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}
			appState[authTypes.ModuleName] = genesisAuthBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().Uint64Var(&txSigLimit, flagTxSigLimit, 0, "Max number of signatures allowed in a transaction.")

	return cmd
}

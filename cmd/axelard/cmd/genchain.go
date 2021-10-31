package cmd

import (
	"encoding/json"
	"fmt"

	nexusExported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

// AddGenesisEVMChainCmd returns set-genesis-chain cobra Command.
func AddGenesisEVMChainCmd(defaultNodeHome string) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "add-genesis-evm-chain [name] [native asset]",
		Short: "Adds an EVM chain in genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc.(codec.Codec)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			name := args[0]
			nativeAsset := args[1]

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			chain := nexusExported.Chain{
				Name:                  name,
				NativeAsset:           nativeAsset,
				SupportsForeignAssets: true,
				KeyType:               tss.Multisig,
			}

			if err := chain.Validate(); err != nil {
				return err
			}

			genesisState := nexusTypes.GetGenesisStateFromAppState(cdc, appState)
			chains := genesisState.Params.Chains
			chains = append(chains, chain)
			genesisState.Params.Chains = chains

			genesisStateBz, err := cdc.MarshalJSON(&genesisState)
			if err != nil {
				return fmt.Errorf("failed to marshal nexus genesis state: %w", err)
			}

			appState[nexusTypes.ModuleName] = genesisStateBz
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		}}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	return cmd
}

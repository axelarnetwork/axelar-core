package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	nexusExported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

// AddGenesisChainSpecCmd returns set-genesis-chain cobra Command.
func AddGenesisChainSpecCmd(defaultNodeHome string) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "add-genesis-chain-spec [name] [platform] [native asset] [supports foreign assets]",
		Short: "Add a chain spec in genesis.json",
		Long:  "Add a chain spec in genesis.json. If the chain is already set in the genesis file, it will be updated.",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			name := args[0]
			platform := strings.ToLower(args[1])
			nativeAsset := args[2]
			supportsForeign, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("last parameter must be a boolean value")
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			chain := nexusExported.Chain{
				Name:                  name,
				Platform:              platform,
				NativeAsset:           nativeAsset,
				SupportsForeignAssets: supportsForeign,
			}

			if err := chain.Validate(); err != nil {
				return err
			}

			genesisState := nexusTypes.GetGenesisStateFromAppState(cdc, appState)

			chains := genesisState.Params.Chains
			for i, chain := range chains {
				if strings.ToLower(chain.Name) == strings.ToLower(name) {
					chains = append(chains[:i], chains[i+1:]...)
				}
			}
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

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
)

const (
	flagGateway  = "gateway"
	flagToken    = "token"
	flagBurnable = "burnable"
)

// SetGenesisEVMContractsCmd returns set-genesis-chain-params cobra Command.
func SetGenesisEVMContractsCmd(defaultNodeHome string) *cobra.Command {
	var gatewayFile, tokenFile, burnableFile string

	cmd := &cobra.Command{
		Use:   "set-genesis-evm-contracts",
		Short: "Set the EVM's contract parameters in genesis.json",
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
			genesisState := evmTypes.GetGenesisStateFromAppState(cdc, appState)

			if gatewayFile != "" {
				gateway, err := getByteCodes(gatewayFile)
				if err != nil {
					return err
				}
				//TODO:  Currently assuming a single element in the Params slice. We need to generalize for more EVM chains.
				genesisState.Params[0].Gateway = gateway
			}

			if tokenFile != "" {
				token, err := getByteCodes(tokenFile)
				if err != nil {
					return err
				}
				//TODO:  Currently assuming a single element in the Params slice. We need to generalize for more EVM chains.
				genesisState.Params[0].Token = token
			}

			if burnableFile != "" {
				burnable, err := getByteCodes(burnableFile)
				if err != nil {
					return err
				}
				//TODO:  Currently assuming a single element in the Params slice. We need to generalize for more EVM chains.
				genesisState.Params[0].Burnable = burnable
			}

			genesisStateBz, err := cdc.MarshalJSON(&genesisState)
			if err != nil {
				return fmt.Errorf("failed to marshal genesis state: %w", err)
			}
			appState[evmTypes.ModuleName] = genesisStateBz
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")

	cmd.Flags().StringVar(&gatewayFile, flagGateway, "", "Path to the Axelar Gateway contract ABI.")
	cmd.Flags().StringVar(&tokenFile, flagToken, "", "Path to the tokens contract ABI.")
	cmd.Flags().StringVar(&burnableFile, flagBurnable, "", "Path to the burner contract ABI.")

	return cmd
}

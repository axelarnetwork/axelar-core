package main

import (
	"fmt"

	ethereumTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// SetGenesisEthContractsCmd returns set-genesis-chain-params cobra Command.
func SetGenesisEthContractsCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {

	var gatewayFile string
	var tokenFile string
	var burnableFile string
	var tokenDeploySig string

	cmd := &cobra.Command{
		Use:   "set-genesis-ethereum-contracts",
		Short: "Set the ethereum's contract parameters in genesis.json",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			genesisState := ethereumTypes.GetGenesisStateFromAppState(cdc, appState)

			if gatewayFile != "" {
				gateway, err := getByteCodes(gatewayFile)
				if err != nil {
					return err
				}
				genesisState.Params.Gateway = gateway
			}

			if tokenFile != "" {
				token, err := getByteCodes(tokenFile)
				if err != nil {
					return err
				}
				genesisState.Params.Token = token
			}

			if burnableFile != "" {
				burnable, err := getByteCodes(burnableFile)
				if err != nil {
					return err
				}
				genesisState.Params.Burnable = burnable
			}

			if tokenDeploySig != "" {
				genesisState.Params.TokenDeploySig = crypto.Keccak256Hash([]byte(tokenDeploySig)).Bytes()
			}

			genesisStateBz, err := cdc.MarshalJSON(genesisState)
			if err != nil {
				return fmt.Errorf("failed to marshal ethereum genesis state: %w", err)
			}
			appState[ethereumTypes.ModuleName] = genesisStateBz
			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().StringVar(&gatewayFile, "gateway", "", "Path to the Axelar Gateway contract ABI.")
	cmd.Flags().StringVar(&tokenFile, "token", "", "Path to the tokens contract ABI.")
	cmd.Flags().StringVar(&burnableFile, "burnable", "", "Path to the burner contract ABI.")
	cmd.Flags().StringVar(&tokenDeploySig, "token-deploy-sig", "", "The signature of Axelar Gateway token deployment method (e.g.,\"TokenDeployed(string,address)\").")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

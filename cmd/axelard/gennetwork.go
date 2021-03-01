package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	bitcoinTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	ethereumTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// SetGenesisChainParamsCmd returns set-genesis-chain-params cobra Command.
func SetGenesisChainParamsCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {
	var networkStr string
	var confirmationHeight uint64

	cmd := &cobra.Command{
		Use:   "set-genesis-chain-params [chain]",
		Short: "Set the chain's parameters in genesis.json",
		Long:  "Set the chain's parameters in genesis.json. The provided chain must be one of those axelar supports.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainStr := args[0]

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var genesisStateBz []byte
			var moduleName string

			switch strings.ToLower(chainStr) {
			case strings.ToLower(btc.Bitcoin.Name):
				genesisState := bitcoinTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = bitcoinTypes.ModuleName

				if networkStr != "" {
					network, err := bitcoinTypes.NetworkFromStr(networkStr)
					if err != nil {
						return err
					}

					genesisState.Params.Network = network
				}

				if confirmationHeight > 0 {
					genesisState.Params.ConfirmationHeight = confirmationHeight
				}

				genesisStateBz, err = cdc.MarshalJSON(genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal bitcoin genesis state: %w", err)
				}
			case strings.ToLower(eth.Ethereum.Name):
				genesisState := ethereumTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = ethereumTypes.ModuleName

				if networkStr != "" {
					network, err := ethereumTypes.NetworkFromStr(networkStr)
					if err != nil {
						return err
					}

					genesisState.Params.Network = network
				}

				if confirmationHeight > 0 {
					genesisState.Params.ConfirmationHeight = confirmationHeight
				}

				genesisStateBz, err = cdc.MarshalJSON(genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal ethereum genesis state: %w", err)
				}
			default:
				return fmt.Errorf("unknown chain: %s", chainStr)
			}

			appState[moduleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		}}

	cmd.Flags().StringVar(&networkStr, "network", "", "Name of the network to set for the given chain.")
	cmd.Flags().Uint64Var(&confirmationHeight, "confirmation-height", 0, "Confirmation height to set for the given chain.")

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

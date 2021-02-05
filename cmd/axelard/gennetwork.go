package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	bitcoinTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	ethereumTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// SetGenesisNetworkCmd returns set-genesis-network cobra Command.
func SetGenesisNetworkCmd(
	ctx *server.Context, cdc *codec.Codec, defaultNodeHome, defaultClientHome string,
) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "set-genesis-network [chain] [network]",
		Short: "Set the chain's network in genesis.json",
		Long:  "Set the chain's network in genesis.json. The provided chain must be one of those axelar supports, as well as the given network.",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainStr := args[0]
			networkStr := args[1]

			chain := balance.ChainFromString(chainStr)
			if err := chain.Validate(); err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var genesisStateBz []byte
			var moduleName string

			switch chain {
			case balance.Bitcoin:
				network, err := bitcoinTypes.NetworkFromStr(networkStr)
				if err != nil {
					return err
				}

				genesisState := bitcoinTypes.GetGenesisStateFromAppState(cdc, appState)

				genesisState.Params.Network = network
				genesisStateBz, err = cdc.MarshalJSON(genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal bitcoin genesis state: %w", err)
				}

				moduleName = bitcoinTypes.ModuleName
			case balance.Ethereum:
				network, err := ethereumTypes.NetworkFromStr(networkStr)
				if err != nil {
					return err
				}

				genesisState := ethereumTypes.GetGenesisStateFromAppState(cdc, appState)

				genesisState.Params.Network = network
				genesisStateBz, err = cdc.MarshalJSON(genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal ethereum genesis state: %w", err)
				}

				moduleName = ethereumTypes.ModuleName
			default:
				return fmt.Errorf("unknown chain: %s", chain.String())
			}

			appState[moduleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		}}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(CliHomeFlag, defaultClientHome, "client's home directory")

	return cmd
}

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	bitcoinTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagConfHeight = "confirmation-height"
	flagNetwork    = "network"
)

// SetGenesisChainParamsCmd returns set-genesis-chain-params cobra Command.
func SetGenesisChainParamsCmd(defaultNodeHome string) *cobra.Command {
	var (
		networkStr         string
		confirmationHeight uint64
	)
	cmd := &cobra.Command{
		Use:   "set-genesis-chain-params [chain]",
		Short: "Set the chain's parameters in genesis.json",
		Long:  "Set the chain's parameters in genesis.json. The provided chain must be one of those axelar supports.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			chainStr := args[0]

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
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

				genesisStateBz, err = cdc.MarshalJSON(&genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal bitcoin genesis state: %w", err)
				}
			case strings.ToLower(evm.Ethereum.Name):
				genesisState := evmTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = evmTypes.ModuleName

				if networkStr != "" {
					network, err := evmTypes.NetworkFromStr(networkStr)
					if err != nil {
						return err
					}

					genesisState.Params.Network = network
				}

				if confirmationHeight > 0 {
					genesisState.Params.ConfirmationHeight = confirmationHeight
				}

				genesisStateBz, err = cdc.MarshalJSON(&genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal ethereum genesis state: %w", err)
				}
			default:
				return fmt.Errorf("unknown chain: %s", chainStr)
			}

			appState[moduleName] = genesisStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		}}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().StringVar(&networkStr, flagNetwork, "", "Name of the network to set for the given chain.")
	cmd.Flags().Uint64Var(&confirmationHeight, flagConfHeight, 0, "Confirmation height to set for the given chain.")

	return cmd
}

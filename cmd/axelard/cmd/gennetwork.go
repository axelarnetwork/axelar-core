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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagConfHeight          = "confirmation-height"
	flagNetwork             = "network"
	flagrevoteLockingPEriod = "revote-locking-period"

	//EVM only
	flagEVMChainName   = "evm-chain-name"
	flagEVMNetworkName = "evm-network-name"
	flagEVMChainID     = "evm-chain-id"
)

// SetGenesisChainParamsCmd returns set-genesis-chain-params cobra Command.
func SetGenesisChainParamsCmd(defaultNodeHome string) *cobra.Command {
	var (
		expectedNetwork     string
		confirmationHeight  uint64
		revoteLockingPeriod int64

		// EVM only
		evmChainName   string
		evmNetworkName string
		evmChainID     string
	)
	cmd := &cobra.Command{
		Use:   "set-genesis-chain-params [bitcoin | evm]",
		Short: "Set chain parameters in genesis.json",
		Long:  "Set chain parameters in genesis.json. The provided platform must be one of those axelar supports (bitcoin, EVM).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			platformStr := args[0]

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var genesisStateBz []byte
			var moduleName string

			switch strings.ToLower(platformStr) {
			case strings.ToLower(btc.Bitcoin.Name):
				genesisState := bitcoinTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = bitcoinTypes.ModuleName

				if expectedNetwork != "" {
					network, err := bitcoinTypes.NetworkFromStr(expectedNetwork)
					if err != nil {
						return err
					}

					genesisState.Params.Network = network
				}

				if confirmationHeight > 0 {
					genesisState.Params.ConfirmationHeight = confirmationHeight
				}

				if revoteLockingPeriod > 0 {
					genesisState.Params.RevoteLockingPeriod = revoteLockingPeriod
				}

				genesisStateBz, err = cdc.MarshalJSON(&genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal bitcoin genesis state: %w", err)
				}
			case strings.ToLower(evmTypes.ModuleName):
				if evmChainName == "" {
					return fmt.Errorf("flag %s is required for EVM platform", flagEVMChainName)
				}

				genesisState := evmTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = evmTypes.ModuleName
				var params evmTypes.Params

				_, index := findEVMChain(genesisState.Params, evmChainName)
				if index < 0 {
					defaults := evmTypes.DefaultParams()
					params, _ = findEVMChain(defaults, evm.Ethereum.Name)
					params.Chain = evmChainName
					params.Network = ""
					params.Networks = []evmTypes.NetworkInfo{}
					genesisState.Params = append(genesisState.Params, params)
					index = len(genesisState.Params) - 1
				}

				if evmNetworkName == "" || evmChainID == "" {
					return fmt.Errorf("flags %s and %s must be used together", flagEVMNetworkName, flagEVMChainID)

				}

				id, ok := sdk.NewIntFromString(evmChainID)
				if !ok {
					return fmt.Errorf("chain ID must be an integer")
				}

				i := findEVMNetwork(genesisState.Params[index].Networks, evmNetworkName)
				if i < 0 {
					genesisState.Params[index].Networks =
						append(genesisState.Params[index].Networks,
							evmTypes.NetworkInfo{Name: evmNetworkName, Id: id})
				} else {
					genesisState.Params[index].Networks[i].Id = id
				}

				if expectedNetwork == "" {
					return fmt.Errorf("flags %s must be specified", flagNetwork)

				}
				found := false
				for _, network := range params.Networks {
					if network.Name == expectedNetwork {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("unable to find network %s", expectedNetwork)
				}

				genesisState.Params[index].Network = expectedNetwork

				if confirmationHeight > 0 {
					genesisState.Params[index].ConfirmationHeight = confirmationHeight
				}

				if revoteLockingPeriod > 0 {
					genesisState.Params[index].RevoteLockingPeriod = revoteLockingPeriod
				}

				genesisStateBz, err = cdc.MarshalJSON(&genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal ethereum genesis state: %w", err)
				}
			default:
				return fmt.Errorf("unknown platform: %s", platformStr)
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
	cmd.Flags().StringVar(&expectedNetwork, flagNetwork, "", "Name of the network to set for the given chain.")
	cmd.Flags().Uint64Var(&confirmationHeight, flagConfHeight, 0, "Confirmation height to set for the given chain.")
	cmd.Flags().Int64Var(&revoteLockingPeriod, flagrevoteLockingPEriod, 0, "Revote locking period to set for the given chain.")
	cmd.Flags().StringVar(&evmChainName, flagEVMChainName, "", "Chain name (EVM only, required).")
	cmd.Flags().StringVar(&evmNetworkName, flagEVMNetworkName, "", "Network name (EVM only, required).")
	cmd.Flags().StringVar(&evmChainID, flagEVMChainID, "", "Integer representing the chain ID (EVM only, required).")

	return cmd
}

func findEVMChain(params []evmTypes.Params, chain string) (param evmTypes.Params, index int) {
	for index, param = range params {
		if param.Chain == chain {
			return
		}
	}

	index = -1
	return
}

func findEVMNetwork(networks []evmTypes.NetworkInfo, network string) (index int) {
	var info evmTypes.NetworkInfo
	for index, info = range networks {
		if info.Name == network {
			return
		}
	}

	index = -1
	return
}

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

const (
	flagConfHeight          = "confirmation-height"
	flagNetwork             = "network"
	flagRevoteLockingPeriod = "revote-locking-period"

	// EVM only
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
		evmNetworkName string
		evmChainID     string
	)
	cmd := &cobra.Command{
		Use:   "set-genesis-chain-params evm [chain]",
		Short: "Set chain parameters in genesis.json",
		Long: "Set chain parameters in genesis.json. " +
			"The provided platform must be one of those axelar supports (currently only EVM).",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

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
			case strings.ToLower(evmTypes.ModuleName):
				if len(args) < 2 {
					return fmt.Errorf("chain name is required for EVM platform")
				}
				evmChainName := nexus.ChainName(args[1])
				if err := evmChainName.Validate(); err != nil {
					return err
				}

				// fetch existing EVM chain, or add new one
				genesisState := evmTypes.GetGenesisStateFromAppState(cdc, appState)
				moduleName = evmTypes.ModuleName
				var chain evmTypes.GenesisState_Chain

				_, index := findEVMChain(genesisState.Chains, evmChainName)
				if index < 0 {
					defaults := evmTypes.DefaultChains()
					chain, _ = findEVMChain(defaults, evm.Ethereum.Name)
					chain.Params.Chain = evmChainName
					chain.Params.Network = ""
					chain.Params.Networks = []evmTypes.NetworkInfo{}
					genesisState.Chains = append(genesisState.Chains, chain)
					index = len(genesisState.Chains) - 1
				}

				// update confirmation height
				if confirmationHeight > 0 {
					genesisState.Chains[index].Params.ConfirmationHeight = confirmationHeight
				}

				// update revote locking period
				if revoteLockingPeriod > 0 {
					genesisState.Chains[index].Params.RevoteLockingPeriod = revoteLockingPeriod
				}

				// if we are editing the list of known networks, both evm-network-name
				// and evm-chain-id need to be used
				if (evmNetworkName != "" && evmChainID == "") || (evmNetworkName == "" && evmChainID != "") {
					return fmt.Errorf("flags '-%s' and '-%s' must be used together", flagEVMNetworkName, flagEVMChainID)

				}

				// add new, or update existing network
				if evmNetworkName != "" && evmChainID != "" {
					id, ok := sdk.NewIntFromString(evmChainID)
					if !ok {
						return fmt.Errorf("chain ID must be an integer")
					}

					i := findEVMNetwork(genesisState.Chains[index].Params.Networks, evmNetworkName)
					if i < 0 {
						genesisState.Chains[index].Params.Networks =
							append(genesisState.Chains[index].Params.Networks,
								evmTypes.NetworkInfo{Name: evmNetworkName, Id: id})
					} else {
						genesisState.Chains[index].Params.Networks[i].Id = id
					}

				}

				// update expected network
				if expectedNetwork != "" {
					i := findEVMNetwork(genesisState.Chains[index].Params.Networks, expectedNetwork)
					if i < 0 {
						return fmt.Errorf("unable to find network %s", expectedNetwork)
					}

					genesisState.Chains[index].Params.Network = genesisState.Chains[index].Params.Networks[i].Name
				}

				genesisStateBz, err = cdc.MarshalJSON(&genesisState)
				if err != nil {
					return fmt.Errorf("failed to marshal genesis state: %w", err)
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
	cmd.Flags().Int64Var(&revoteLockingPeriod, flagRevoteLockingPeriod, 0, "Revote locking period to set for the given chain.")
	cmd.Flags().StringVar(&evmNetworkName, flagEVMNetworkName, "", "Network name (EVM only).")
	cmd.Flags().StringVar(&evmChainID, flagEVMChainID, "", "Integer representing the chain ID (EVM only).")

	return cmd
}

func findEVMChain(chains []evmTypes.GenesisState_Chain, chainName nexus.ChainName) (chain evmTypes.GenesisState_Chain, index int) {
	for index, chain = range chains {
		if chainName.Equals(chain.Params.Chain) {
			return
		}
	}

	index = -1
	return
}

func findEVMNetwork(networks []evmTypes.NetworkInfo, network string) (index int) {
	var info evmTypes.NetworkInfo
	for index, info = range networks {
		if strings.EqualFold(info.Name, network) {
			return
		}
	}

	index = -1
	return
}

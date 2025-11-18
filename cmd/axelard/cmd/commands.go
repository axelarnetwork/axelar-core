package cmd

import (
	"errors"
	"io"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	rosettaCmd "github.com/cosmos/rosetta/cmd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/vald"
)

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	basicManager := app.GetModuleBasics()

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
		genesisCommand(encodingConfig.TxConfig, basicManager),
		tmcli.NewCompletionCmd(rootCmd, true),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		SetGenesisRewardCmd(app.DefaultNodeHome),
		SetGenesisStakingCmd(app.DefaultNodeHome),
		SetGenesisSlashingCmd(app.DefaultNodeHome),
		SetGenesisVoteCmd(app.DefaultNodeHome),
		SetGenesisSnapshotCmd(app.DefaultNodeHome),
		SetGenesisEVMContractsCmd(app.DefaultNodeHome),
		SetGenesisChainParamsCmd(app.DefaultNodeHome),
		SetGenesisGovCmd(app.DefaultNodeHome),
		AddGenesisEVMChainCmd(app.DefaultNodeHome),
		SetGenesisMintCmd(app.DefaultNodeHome),
		SetMultisigGovernanceCmd(app.DefaultNodeHome),
		SetGenesisCrisisCmd(app.DefaultNodeHome),
		SetGenesisAuthCmd(app.DefaultNodeHome),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, export(encodingConfig), addModuleInitFlags)
	wasmcli.ExtendUnsafeResetAllCmd(rootCmd)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)

	// Add rosetta command
	rootCmd.AddCommand(rosettaCmd.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Codec))

	// Only set default, not actual value, so it can be overwritten by env variable
	utils.OverwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagBroadcastMode:    flags.BroadcastSync,
		flags.FlagChainID:          app.Name,
		flags.FlagGasPrices:        minGasPrice,
		flags.FlagKeyringBackend:   "file",
		flags.FlagSkipConfirmation: "true",
	}, false)

	rootCmd.PersistentFlags().String(tmcli.OutputFlag, "text", "Output format (text|json)")

	// add vald after the overwrite so it can set its own defaults
	rootCmd.AddCommand(vald.GetValdCommand(), vald.GetHealthCheckCommand(), vald.GetSignCommand())
}

// genesisCommand builds genesis-related `simd genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.QueryEventForTxCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockCmd(),
		server.QueryBlocksCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
	startCmd.Flags().String(app.WasmDirFlag, "", "path to the wasm directory, by default set to 'wasm' directory inside the '--db_dir' directory")
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	// this allows for faster block times, because nodes can start optimistic execution of blocks while they are being voted on
	baseappOptions = append(baseappOptions, baseapp.SetOptimisticExecution())

	var wasmOpts []wasm.Option
	if app.IsWasmEnabled() && cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	return app.NewAxelarApp(
		logger, db, traceStore, true,
		app.MakeEncodingConfig(),
		appOpts,
		wasmOpts,
		baseappOptions...,
	)
}

func export(encCfg params.EncodingConfig) servertypes.AppExporter {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
		appOpts servertypes.AppOptions, modulesToExport []string) (servertypes.ExportedApp, error) {

		homePath := cast.ToString(appOpts.Get(flags.FlagHome))
		if homePath == "" {
			return servertypes.ExportedApp{}, errors.New("application home not set")
		}

		aApp := app.NewAxelarApp(logger, db, traceStore, height == -1,
			encCfg, appOpts, []wasm.Option{})
		if height != -1 {
			if err := aApp.LoadHeight(height); err != nil {
				return servertypes.ExportedApp{}, err
			}
		}

		return aApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
	}
}

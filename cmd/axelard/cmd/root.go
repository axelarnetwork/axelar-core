package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingcli "github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/config"
	"github.com/axelarnetwork/axelar-core/vald"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
)

var (
	minGasPrice = "0.007" + axelarnet.NativeAsset
	wasmDirFlag = "wasm-dir"
)

// NewRootCmd creates a new root command for axelard. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	app.SetConfig()

	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "Axelar App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			axelarAppTemplate, axelarAppConfig := initAppConfig()
			err := server.InterceptConfigsPreRunHandler(cmd, axelarAppTemplate, axelarAppConfig)
			if err != nil {
				return err
			}

			// InterceptConfigsPreRunHandler initializes a console logger with an improper time format with no way of changing the config,
			// so we need to overwrite the logger
			err = overwriteLogger(cmd)
			if err != nil {
				return err
			}

			// we don't have direct access to the definition of the start command, so this is the only place we can add additional seeds
			if cmd.Use == "start" {
				if err := extendSeeds(cmd); err != nil {
					return err
				}
			}

			return nil
		},
	}

	initRootCmd(rootCmd, encodingConfig)
	return rootCmd, encodingConfig
}

func extendSeeds(cmd *cobra.Command) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	seeds, err := config.ReadSeeds(serverCtx.Viper)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		serverCtx.Logger.Info("file seeds.toml not found")
		return nil
	}
	if err != nil {
		return err
	}

	serverCtx.Config = config.MergeSeeds(serverCtx.Config, seeds)
	if err := server.SetCmdServerContext(cmd, serverCtx); err != nil {
		return err
	}

	serverCtx.Logger.Info(fmt.Sprintf("adding %d seeds from seeds.toml", len(seeds)))

	return nil
}

func overwriteLogger(cmd *cobra.Command) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	var logWriter io.Writer
	if strings.ToLower(serverCtx.Viper.GetString(flags.FlagLogFormat)) == tmcfg.LogFormatPlain {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	} else {
		logWriter = os.Stderr
	}

	logLvlStr := serverCtx.Viper.GetString(flags.FlagLogLevel)
	logLvl, err := zerolog.ParseLevel(logLvlStr)
	if err != nil {
		return fmt.Errorf("failed to parse log level (%s): %w", logLvlStr, err)
	}

	serverCtx.Logger = server.ZeroLogWrapper{Logger: zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()}
	return nil
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	type AxelarAppConfig struct {
		serverconfig.Config
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the default min gas price.
	srvCfg.MinGasPrices = minGasPrice

	axelarAppConfig := AxelarAppConfig{
		Config: *srvCfg,
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate

	return customAppTemplate, axelarAppConfig
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(app.GetModuleBasics(), app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(app.GetModuleBasics(), encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.GetModuleBasics()),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
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

	starterFlags := func(startCmd *cobra.Command) {
		crisis.AddModuleInitFlags(startCmd)
		startCmd.Flags().String(wasmDirFlag, "", "path to the wasm directory, by default set to 'wasm' directory inside the '--db_dir' directory")
	}

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, export(encodingConfig), starterFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(app.DefaultNodeHome),
	)

	// Add rosetta command
	rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Codec))

	// Only set default, not actual value, so it can be overwritten by env variable
	utils.OverwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagBroadcastMode:    flags.BroadcastBlock,
		flags.FlagChainID:          app.Name,
		flags.FlagGasPrices:        minGasPrice,
		flags.FlagKeyringBackend:   "file",
		flags.FlagSkipConfirmation: "true",
	}, false)

	rootCmd.PersistentFlags().String(tmcli.OutputFlag, "text", "Output format (text|json)")

	// add vald after the overwrite so it can set its own defaults
	rootCmd.AddCommand(vald.GetValdCommand(), vald.GetHealthCheckCommand(), vald.GetSignCommand())
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	var wasmOpts []wasm.Option
	if app.IsWasmEnabled() && cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	return app.NewAxelarApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToString(appOpts.Get(wasmDirFlag)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		app.MakeEncodingConfig(),
		appOpts,
		wasmOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

func export(encCfg params.EncodingConfig) servertypes.AppExporter {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
		appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {

		homePath := cast.ToString(appOpts.Get(flags.FlagHome))
		if homePath == "" {
			return servertypes.ExportedApp{}, errors.New("application home not set")
		}

		wasmdir := cast.ToString(appOpts.Get(wasmDirFlag))
		aApp := app.NewAxelarApp(logger, db, traceStore, height == -1, map[int64]bool{}, homePath, wasmdir,
			cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)), encCfg, appOpts, []wasm.Option{})
		if height != -1 {
			if err := aApp.LoadHeight(height); err != nil {
				return servertypes.ExportedApp{}, err
			}
		}

		return aApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
	}
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
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.GetModuleBasics().AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

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
		flags.LineBreak,
		vestingcli.GetTxCmd(),
	)

	app.GetModuleBasics().AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

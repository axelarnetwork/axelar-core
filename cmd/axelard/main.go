package main

import (
	"encoding/json"
	"io"
	"path"

	"github.com/cosmos/cosmos-sdk/store/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/axelarnetwork/axelar-core/app"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	flagInvCheckPeriod = "inv-check-period"
	CliHomeFlag        = "clihome"
)

var invCheckPeriod uint

func main() {
	cdc := app.MakeCodec()
	app.SetConfig()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "axelard",
		Short:             "Axelar Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.MigrateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
			auth.GenesisAccountIterator{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
	)
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics))
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisStakingCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisVoteCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisTSSCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisSnapshotCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisEthContractsCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(SetGenesisChainParamsCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(flags.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(debug.Cmd(cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "AX", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")

	addAdditionalFlags(rootCmd)

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func addAdditionalFlags(rootCmd *cobra.Command) {
	// These flags are intended for submitting transactions.
	// Since the daemon can broadcast messages now we need the flags here as well (mostly for their default values)
	flags.PostCommands(rootCmd)

	// This flag has a defined default in PostCommands, but is not bound to viper
	_ = viper.BindPFlag(flags.FlagGasAdjustment, rootCmd.Flags().Lookup(flags.FlagGasAdjustment))

	// The daemon acts as a broadcaster as well, so we need access to the cli configuration (in particular the keybase)
	rootCmd.PersistentFlags().String(CliHomeFlag, app.DefaultCLIHome, "directory for cli config and data")
	_ = viper.BindPFlag(CliHomeFlag, rootCmd.Flags().Lookup(CliHomeFlag))

	rootCmd.PersistentFlags().String("tofnd-host", "", "host name for tss daemon")
	_ = viper.BindPFlag("tofnd_host", rootCmd.PersistentFlags().Lookup("tofnd-host"))

	rootCmd.PersistentFlags().String("tofnd-port", "50051", "port for tss daemon")
	_ = viper.BindPFlag("tofnd_port", rootCmd.PersistentFlags().Lookup("tofnd-port"))
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	return app.NewInitApp(
		logger, db, traceStore, true, invCheckPeriod, loadConfig(),
		baseapp.SetPruning(types.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
		baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
		baseapp.SetInterBlockCache(cache),
	)
}

func loadConfig() app.Config {
	// need to merge in cli config because axelard now has its own broadcasting client
	conf := app.DefaultConfig()
	homeDir := viper.GetString(cli.HomeFlag)
	cliHomeDir := viper.GetString(CliHomeFlag)
	cliCfgFile := path.Join(cliHomeDir, "config", "config.toml")
	viper.SetConfigFile(cliCfgFile)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
	cfgFile := path.Join(homeDir, "config", "config.toml")
	viper.SetConfigFile(cfgFile)

	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}
	return conf
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	conf := loadConfig()

	if height != -1 {
		aApp := app.NewInitApp(logger, db, traceStore, false, uint(1), conf)
		err := aApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	aApp := app.NewInitApp(logger, db, traceStore, true, uint(1), conf)

	return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

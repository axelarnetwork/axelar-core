package main

import (
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/axelarnetwork/axelar-core/app"
)

const cliHomeFlag = "clihome"

func main() {
	configurate()

	rootCmd := &cobra.Command{
		Use:   "vald",
		Short: "Validator Daemon ",
	}

	startCommand := getStartCommand(log.NewTMLogger(os.Stdout).With("external", "main"))
	rootCmd.AddCommand(flags.PostCommands(startCommand)...)

	setPersistentFlags(rootCmd)

	executor := cli.PrepareMainCmd(rootCmd, "AX", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		tmos.Exit(err.Error())
	}
}

func configurate() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	config.Seal()
}

func setPersistentFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().String(cliHomeFlag, app.DefaultCLIHome, "directory for cli config and data")
	_ = viper.BindPFlag(cliHomeFlag, rootCmd.PersistentFlags().Lookup(cliHomeFlag))

	rootCmd.PersistentFlags().String("tofnd-host", "", "host name for tss daemon")
	_ = viper.BindPFlag("tofnd_host", rootCmd.PersistentFlags().Lookup("tofnd-host"))

	rootCmd.PersistentFlags().String("tofnd-port", "50051", "port for tss daemon")
	_ = viper.BindPFlag("tofnd_port", rootCmd.PersistentFlags().Lookup("tofnd-port"))

	rootCmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator")
	_ = viper.BindPFlag("validator-addr", rootCmd.PersistentFlags().Lookup("validator-addr"))
}

func loadConfig() (app.Config, string) {
	// need to merge in cli config because axelard now has its own broadcasting client
	conf := app.DefaultConfig()
	homeDir := viper.GetString(cli.HomeFlag)
	cliHomeDir := viper.GetString(cliHomeFlag)
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
	// for some reason gas is not being filled
	conf.Gas = viper.GetUint64("gas")

	return conf, viper.GetString("validator-addr")
}

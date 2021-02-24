package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra/doc"

	"github.com/axelarnetwork/axelar-core/app"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
)

//go:generate ./genDocs.sh ./docs

func main() {
	docs := flag.String("docs", "", "only generate documentation for the CLI commands into the specified folder")
	flag.Parse()
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()
	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	config.Seal()

	// If run with the docs flag, generate documentation for all CLI commands
	if *docs != "" {
		cmd := CreateRootCmd(cdc, "$HOME/.axelarcli")
		// The AutoGen tag includes a date, so when the time zone of the local machine is different from the time zone
		// of the github host the date could be different and the PR check fail. Therefore we disable it
		cmd.DisableAutoGenTag = true
		deleteLineBreakCmds(cmd)
		if err := doc.GenMarkdownTree(cmd, *docs); err != nil {
			fmt.Printf("Failed generating CLI command documentation: %s, exiting...\n", err)
			os.Exit(1)
		}

		if err := genTOC(cmd, *docs); err != nil {
			fmt.Printf("Failed generating CLI command table of contents: %s, exiting...\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	rootCmd := CreateRootCmd(cdc, app.DefaultCLIHome)

	// Add flags and prefix all env exposed with AX
	executor := cli.PrepareMainCmd(rootCmd, "AX", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func CreateRootCmd(cdc *codec.Codec, homeDir string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "axelarcli",
		Short: "Axelar Client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Network ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(homeDir),
		queryCmd(cdc),
		txCmd(cdc),
		flags.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)
	return rootCmd
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		flags.LineBreak,
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		flags.LineBreak,
		authcmd.GetEncodeCommand(cdc),
		authcmd.GetDecodeCommand(cdc),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == auth.ModuleName || cmd.Use == bank.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)

	return txCmd
}

func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
	authRest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}

func genTOC(cmd *cobra.Command, dir string) error {
	toc := make([]string, 0)
	toc = append(toc, genTOCEntry(cmd, dir)...)
	filename := filepath.Join(dir, "toc.md")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, "# CLI command overview\n")
	if err != nil {
		return err
	}
	for _, s := range toc {
		_, err = io.WriteString(f, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func genTOCEntry(cmd *cobra.Command, dir string) []string {
	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	label := cmd.Use
	label = strings.ReplaceAll(label, "<", "\\<")
	label = strings.ReplaceAll(label, ">", "\\>")
	toc := []string{fmt.Sprintf("- [%s](%s)\t - %s\n", label, basename, cmd.Short)}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		for _, entry := range genTOCEntry(c, dir) {
			toc = append(toc, "\t"+entry)
		}
	}
	return toc
}

func deleteLineBreakCmds(cmd *cobra.Command) {
	cmd.RemoveCommand(flags.LineBreak)
	for _, c := range cmd.Commands() {
		deleteLineBreakCmds(c)
	}
}

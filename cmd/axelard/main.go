package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
)

//go:generate ./genDocs.sh ../../docs/cli

func main() {
	docs := flag.String("docs", "", "only generate documentation for the CLI commands into the specified folder")
	flag.Parse()

	rootCmd, _ := cmd.NewRootCmd()
	// If run with the docs flag, generate documentation for all CLI commands
	if *docs != "" {
		// add flags from svrcmd.Execute()
		rootCmd.PersistentFlags().String(flags.FlagLogLevel, zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
		rootCmd.PersistentFlags().String(flags.FlagLogFormat, tmcfg.LogFormatPlain, "The logging format (json|plain)")
		home := filepath.Join("$HOME", "."+app.Name)
		executor := tmcli.PrepareBaseCmd(rootCmd, "", home)
		rootCmd = executor.Root()

		// set static values for dynamic (system-dependent) flag defaults
		values := map[string]string{
			flags.FlagHome:  home,
			cli.FlagIP:      "127.0.0.1",
			cli.FlagMoniker: "node",
		}
		utils.OverwriteFlagDefaults(rootCmd, values, true)

		// The AutoGen tag includes a date, so when the time zone of the local machine is different from the time zone
		// of the github host the date could be different and the PR check fail. Therefore we disable it
		rootCmd.DisableAutoGenTag = true
		deleteLineBreakCmds(rootCmd)
		if err := doc.GenMarkdownTree(rootCmd, *docs); err != nil {
			fmt.Printf("Failed generating CLI command documentation: %s, exiting...\n", err)
			os.Exit(1)
		}

		if err := genTOC(rootCmd, *docs); err != nil {
			fmt.Printf("Failed generating CLI command table of contents: %s, exiting...\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	setupMetrics()

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
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

func setupMetrics() {
	telemetry.New(telemetry.Config{
		Enabled:        true,
		EnableHostname: false,
		ServiceName:    "axelar",
		// 1<<62, https://play.golang.org/p/szrQPRHxE0O
		// A hacky way to essentially prevent prometheus metrics from ever expiring
		PrometheusRetentionTime: 4611686018427387904,
		EnableHostnameLabel:     false,
	})
}

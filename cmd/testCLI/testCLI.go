package main

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/sdk-tutorials/scavenge/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "testCLI",
		Short: "Scavenge throughput test client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		tpCmd(cdc),
		rpc.StatusCommand(),
		client.ConfigCmd(app.DefaultCLIHome),
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// Add flags and prefix all env exposed with AA
	executor := cli.PrepareMainCmd(rootCmd, "AA", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func tpCmd(cdc *amino.Codec) *cobra.Command {
	tpCmd := &cobra.Command{
		Use:   "tp [txCount] [goroutines] [from_key_or_address] [to_address] [minAmount]",
		Short: "Throughput test",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txCount, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("txCount must be an integer")
			}
			goroutines, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Printf("goroutines must be an integer")
			}
			txPerGR := txCount / goroutines

			to, err := sdk.AccAddressFromBech32(args[3])
			if err != nil {
				return err
			}

			// parse coins trying to be sent
			coins, err := sdk.ParseCoins(args[4])
			if err != nil {
				return err
			}

			wg := &sync.WaitGroup{}
			errChan := make(chan error, goroutines)
			seq := viper.GetUint64(flags.FlagSequence)
			for i := 0; i < goroutines; i += 1 {
				wg.Add(1)
				go func(wg *sync.WaitGroup, errChan chan<- error) {
					defer wg.Done()
					for j := 0; j < txPerGR; j += 1 {
						inBuf := bufio.NewReader(cmd.InOrStdin())
						s := atomic.AddUint64(&seq, 1)
						txBldr := newTxBldr(s-1, inBuf, cdc)

						cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[2]).WithCodec(cdc)

						//coins[0].Amount = coins[0].Amount.AddRaw(int64(j +txCount*i))
						// build and sign the transaction, then broadcast to Tendermint
						msg := bank.NewMsgSend(cliCtx.GetFromAddress(), to, coins)
						if err := utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}); err != nil {
							errChan <- err
							break
						}
					}
				}(wg, errChan)
			}

			select {
			case err := <-errChan:
				return err
			default:
				wg.Wait()
				errChan <- nil
			}
			return nil
		},
	}

	tpCmd = flags.PostCommands(tpCmd)[0]

	return tpCmd
}

func newTxBldr(seq uint64, inBuf *bufio.Reader, cdc *amino.Codec) types.TxBuilder {
	txBldr := auth.NewTxBuilder(
		nil,
		uint64(viper.GetInt64(flags.FlagAccountNumber)),
		seq,
		flags.GasFlagVar.Gas,
		viper.GetFloat64(flags.FlagGasAdjustment),
		flags.GasFlagVar.Simulate,
		viper.GetString(flags.FlagChainID),
		viper.GetString(flags.FlagMemo),
		nil, nil)

	kb, err := crkeys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), inBuf)
	if err != nil {
		panic(err)
	}

	txBldr = txBldr.WithKeybase(kb)
	txBldr = txBldr.WithTxEncoder(utils.GetTxEncoder(cdc))
	txBldr = txBldr.WithFees(viper.GetString(flags.FlagFees))
	txBldr = txBldr.WithGasPrices(viper.GetString(flags.FlagGasPrices))
	return txBldr
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

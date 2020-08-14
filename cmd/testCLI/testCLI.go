package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cheggaaa/pb/v3"
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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/sdk-tutorials/scavenge/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
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
		tpCmd(cdc, keys.AddKeyCommand()),
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

type tx struct {
	msg sdk.Msg
	ctx context.CLIContext
	seq uint64
}

func tpCmd(cdc *amino.Codec, addKeyCommand *cobra.Command) *cobra.Command {
	tpCmd := &cobra.Command{
		Use:   "tp [txCount] [goroutines] [account_with_funds] [amount]",
		Short: "Throughput test",
		Args:  cobra.ExactArgs(4),
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

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := crkeys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), inBuf)
			if err != nil {
				return err
			}

			for i := 0; i < goroutines; i++ {
				name := "test" + strconv.Itoa(i)
				if _, err = kb.Get(name); err == nil {
					continue
				}
				fmt.Printf("Creating account %v\n", name)
				err = addKeyCommand.RunE(cmd, []string{name})
				if err != nil {
					return err
				}
			}

			// parse coins trying to be sent
			coins, err := sdk.ParseCoins(args[3])
			if err != nil {
				return err
			}

			prepareAccountsBar := pb.StartNew(goroutines)
			fmt.Println("Preparing accounts for testing:")

			prepareCoins := sdk.Coins{}

			for _, coin := range coins {
				c := coin
				for i := 0; i < txPerGR-1; i++ {
					c.Add(coin)
				}
				prepareCoins = prepareCoins.Add(c)
			}

			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			prepCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[2]).WithCodec(cdc)
			prepCtx.Output = ioutil.Discard
			prepCtx.SkipConfirm = true
			prepCtx.BroadcastMode = flags.BroadcastSync
			_, prepSeq, err := authtypes.NewAccountRetriever(prepCtx).GetAccountNumberSequence(prepCtx.FromAddress)
			if err != nil {
				return err
			}

			origStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			for i := 0; i < goroutines; i += 1 {
				if i == goroutines-1 {
					prepCtx.BroadcastMode = flags.BroadcastBlock
				}
				to, _, err := context.GetFromFields(inBuf, "test"+strconv.Itoa(i), false)
				if err != nil {
					return err
				}

				msg := bank.NewMsgSend(prepCtx.FromAddress, to, prepareCoins)

				if err := utils.GenerateOrBroadcastMsgs(prepCtx, txBldr.WithSequence(prepSeq), []sdk.Msg{msg}); err != nil {
					return nil
				}
				prepSeq += 1
				prepareAccountsBar.Increment()
			}
			prepareAccountsBar.Finish()
			os.Stdout = origStdout
			wg := &sync.WaitGroup{}
			errChan := make(chan error, goroutines)
			broadcastChan := make(chan tx, txCount)

			fmt.Println("Creating transactions:")

			createMsgBar := pb.StartNew(txCount)

			for i := 0; i < goroutines; i += 1 {
				wg.Add(1)
				go func(wg *sync.WaitGroup, errChan chan<- error, account string) {
					defer wg.Done()
					cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, account).WithCodec(cdc)
					_, seq, err := authtypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(cliCtx.FromAddress)
					if err != nil {
						errChan <- err
						return
					}
					for j := 0; j < txPerGR; j += 1 {

						to, _, err := context.GetFromFields(inBuf, args[2], false)
						if err != nil {
							errChan <- err
							return
						}

						msg := bank.NewMsgSend(cliCtx.FromAddress, to, coins)
						if cliCtx.SkipConfirm {
							cliCtx.Output = ioutil.Discard
						}

						broadcastChan <- tx{
							msg: msg,
							ctx: cliCtx,
							seq: seq,
						}
						seq += 1
						createMsgBar.Increment()
					}
				}(wg, errChan, "test"+strconv.Itoa(i))
			}

			wg.Wait()
			createMsgBar.Finish()

			select {
			case err := <-errChan:
				return err
			default:
			}

			fmt.Println("Sending transactions:")

			sendMsgBar := pb.StartNew(txCount)

			r, w, err := os.Pipe()
			if err != nil {
				panic(err)
			}
			os.Stdout = w
			wg = &sync.WaitGroup{}
			wg.Add(1)

			buf := bytes.Buffer{}
			go func(reader io.Reader, buffer io.Writer) {
				defer wg.Done()
				if _, err := io.Copy(buffer, reader); err != nil {
					panic(err)
				}
			}(r, &buf)

			for i := 0; i < goroutines; i += 1 {
				wg.Add(1)
				go func(wg *sync.WaitGroup, errChan chan<- error) {
					defer wg.Done()
					for j := 0; j < txPerGR; j += 1 {
						tx := <-broadcastChan

						if err := utils.GenerateOrBroadcastMsgs(tx.ctx, txBldr.WithSequence(tx.seq), []sdk.Msg{tx.msg}); err != nil {
							errChan <- err
							return
						}
						sendMsgBar.Increment()
					}
				}(wg, errChan)
			}

			wg.Wait()
			sendMsgBar.Finish()
			_ = r.Close()
			os.Stdout = origStdout
			fmt.Println(buf.String())
			select {
			case err := <-errChan:
				return err
			default:
				return nil
			}
		},
	}

	tpCmd = flags.PostCommands(tpCmd)[0]
	_ = tpCmd.Flags().MarkHidden(flags.FlagSequence)
	return tpCmd
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

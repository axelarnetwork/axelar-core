package vald

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	keyID = "testkey"

	defaultTimeout time.Duration = 2 * time.Hour
)

// GetTofndPingCommand returns the command to ping tofnd
func GetTofndPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "tofnd-ping",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			logger := server.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}

			loadValdCfg(serverCtx)
			valdCfg := config.DefaultValdConfig()
			if err := serverCtx.Viper.Unmarshal(&valdCfg); err != nil {
				panic(err)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				panic(err)
			}

			timeout, err := time.ParseDuration(serverCtx.Viper.GetString("context-timeout"))
			if err != nil {
				return err
			}

			gg20client, err := tss.CreateTOFNDClient(valdCfg.TssConfig.Host, valdCfg.TssConfig.Port, valdCfg.TssConfig.DialTimeout, logger)
			if err != nil {
				return fmt.Errorf("failed to reach tofnd: %s", err.Error())
			}

			grpcCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			request := &tofnd.KeyPresenceRequest{
				// we do not need to look for a key ID that exists to obtain a successful healthcheck,
				// all we need to do is obtain err == nil && response != FAIL
				KeyUid: keyID,
			}

			response, err := gg20client.KeyPresence(grpcCtx, request)

			if err != nil {
				return fmt.Errorf("failed to invoke tofnd grpc: %s", err.Error())
			}

			if response.Response == tofnd.RESPONSE_FAIL ||
				response.Response == tofnd.RESPONSE_UNSPECIFIED {
				return fmt.Errorf("obtained FAIL response, tofnd not properly configured")
			}

			clientCtx.PrintString("Pong!\n")
			return nil
		},
	}

	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String("tofnd-host", defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String("tofnd-port", defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String("context-timeout", defaultTimeout.String(), "context timeout for the grpc")
	return cmd
}

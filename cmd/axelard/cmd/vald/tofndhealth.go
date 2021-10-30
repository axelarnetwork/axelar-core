package vald

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	keyID      = "testkey"
	tokenDenom = "uaxl"
	minBalance = 5000000

	flagSkipTofnd       = "skip-tofnd"
	flagSkipBroadcaster = "skip-broadcaster"
	flagSkipOperator    = "skip-operator"
	flagBroadcasterAddr = "broadcaster-addr"
	flagContextTimeout  = "context-timeout"
	flagTofndHost       = "tofnd-host"
	flagTofndPort       = "tofnd-port"

	defaultTimeout time.Duration = 2 * time.Hour
)

var allGood bool = true

// GetHealthCheckCommand returns the command to execute a node health check
func GetHealthCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "health-check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				panic(err)
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			execCmd(nil, clientCtx, serverCtx, "tofnd", checkTofnd)
			execCmd(cmd.Context(), clientCtx, serverCtx, "broadcaster", checkBroadcaster)
			execCmd(cmd.Context(), clientCtx, serverCtx, "operator", checkOperator)

			// enforce a non-zero exiting code in case health checks fail
			// without printing cobra output
			if !allGood {
				os.Exit(1)
			}

			return nil
		},
	}

	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String(flagTofndHost, defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String(flagTofndPort, defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String(flagContextTimeout, defaultTimeout.String(), "context timeout for the grpc")
	cmd.PersistentFlags().String(flagBroadcasterAddr, "", "broadcaster address")
	cmd.PersistentFlags().Bool(flagSkipTofnd, false, "skip tofnd check")
	cmd.PersistentFlags().Bool(flagSkipBroadcaster, false, "skip broadcaster check")
	cmd.PersistentFlags().Bool(flagSkipOperator, false, "skip operator check")

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

type checkCmd func(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error

func execCmd(ctx context.Context, clientCtx client.Context, serverCtx *server.Context, name string, cmd checkCmd) {
	fmt.Printf("%s check: ", name)
	if !serverCtx.Viper.GetBool(flagSkipTofnd) {
		err := cmd(ctx, clientCtx, serverCtx)
		if err != nil {
			fmt.Printf("failed (%s)\n", err.Error())
			allGood = false
		} else {
			fmt.Println("passed")
		}
	} else {
		fmt.Println("skipped")
	}
}

func checkTofnd(_ context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	valdCfg := config.DefaultValdConfig()
	if err := serverCtx.Viper.Unmarshal(&valdCfg); err != nil {
		panic(err)
	}

	timeout, err := time.ParseDuration(serverCtx.Viper.GetString(flagContextTimeout))
	if err != nil {
		return err
	}

	nopeLogger := server.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}
	gg20client, err := tss.CreateTOFNDClient(valdCfg.TssConfig.Host, valdCfg.TssConfig.Port, valdCfg.TssConfig.DialTimeout, nopeLogger)
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

	return nil
}

func checkBroadcaster(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	str := serverCtx.Viper.GetString(flagBroadcasterAddr)
	if str == "" {
		return fmt.Errorf("no broadcaster address specified")
	}
	addr, err := sdk.AccAddressFromBech32(str)
	if err != nil {
		return err
	}

	queryClient := bankTypes.NewQueryClient(clientCtx)
	params := bankTypes.NewQueryBalanceRequest(addr, tokenDenom)
	res, err := queryClient.Balance(ctx, params)
	if err != nil {
		return err
	}

	if res.Balance.Amount.LTE(sdk.NewInt(minBalance)) {
		return fmt.Errorf("broadcaster does not have enough funds (minimum balance is %d%s)", minBalance, tokenDenom)
	}

	return nil
}

func checkOperator(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	addr := serverCtx.Viper.GetString(flagBroadcasterAddr)
	if addr == "" {
		return fmt.Errorf("no broadcaster address specified")
	}

	bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", snapshotTypes.QuerierRoute, keeper.QValidators))
	if err != nil {
		return err
	}

	var resValidators types.QueryValidatorsResponse
	types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &resValidators)

	bz, _, err = clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", snapshotTypes.QuerierRoute, keeper.QOperator, addr))
	if err != nil {
		return err
	}
	addr = string(bz)

	for _, v := range resValidators.Validators {
		if v.OperatorAddress == addr {
			if v.TssIllegibilityInfo.Jailed ||
				v.TssIllegibilityInfo.MissedTooManyBlocks ||
				v.TssIllegibilityInfo.NoProxyRegistered ||
				v.TssIllegibilityInfo.Tombstoned ||
				v.TssIllegibilityInfo.TssSuspended {
				return fmt.Errorf("health check to operator %s failed due to the following issues: %v",
					addr, string(snapshotTypes.ModuleCdc.MustMarshalJSON(&v.TssIllegibilityInfo)))
			}
			return nil
		}
	}

	return fmt.Errorf("operator address %s not found amongst current set of validators", addr)
}

package keeper

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	axTypes "github.com/axelarnetwork/axelar-core/x/axelar/types"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

var (
	_ axTypes.BridgeKeeper = Keeper{}
)

const (
	ErrRpcInWarmup = btcjson.RPCErrorCode(-28)
)

type Keeper struct {
	client *rpcclient.Client
}

const (
	sleep = 1 * time.Second
)

func NewBtcKeeper(cfg types.BtcConfig, logger log.Logger) (Keeper, error) {
	client, err := newRPCClient(cfg, logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)))
	if err != nil {
		return Keeper{}, err
	}
	return Keeper{client: client}, nil
}

func newRPCClient(cfg types.BtcConfig, logger log.Logger) (*rpcclient.Client, error) {
	if err := waitForAuthCookie(cfg.CookiePath, cfg.StartUpTimeout, logger); err != nil {
		return nil, err
	}

	rpcCfg := &rpcclient.ConnConfig{
		Host:                 cfg.RpcAddr,
		CookiePath:           cfg.CookiePath,
		DisableTLS:           true, // Bitcoin core does not provide TLS by default
		DisableAutoReconnect: false,
		HTTPPostMode:         true, // Bitcoin core only supports HTTP POST mode
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrConnFailed, "could not start the bitcoin rpc client")
	}

	if err = waitForBtcWarmup(client, cfg.RPCTimeout, logger); err != nil {
		return nil, err
	}
	return client, nil
}

func waitForAuthCookie(cookiePath string, timeout time.Duration, logger log.Logger) error {
	matches, _ := filepath.Glob(cookiePath)
	for _, match := range matches {
		fmt.Println(match)
	}

	_, err := os.Stat(cookiePath)
	var t time.Duration
	for os.IsNotExist(err) && t < timeout {
		time.Sleep(sleep)
		t = t + sleep
		logger.Debug("waiting for bitcoin node to create rpc auth cookie")
		_, err = os.Stat(cookiePath)
	}
	if t < timeout {
		return nil
	} else {
		return sdkerrors.Wrap(types.ErrInvalidConfig, fmt.Sprintf("bitcoin auth cookie could not be found at %s", cookiePath))
	}
}

func waitForBtcWarmup(client *rpcclient.Client, timeout time.Duration, logger log.Logger) error {
	conn := connection{
		client:  client,
		retries: 0,
	}

	maxRetries := int(timeout / sleep)
	for !conn.isAvailable() && conn.retries < maxRetries {
		switch conn.error.(type) {
		case *btcjson.RPCError:
			if conn.error.(*btcjson.RPCError).Code == ErrRpcInWarmup {
				logger.Debug("waiting for bitcoin rpc server to start")
				time.Sleep(sleep)
			} else {
				return unexpectedError(conn.error)
			}
		default:
			return unexpectedError(conn.error)
		}
	}

	if conn.retries < maxRetries {
		logger.Info("btc bridge client successfully connected to bitcoin node")
		return nil
	} else {
		return sdkerrors.Wrap(types.ErrTimeOut, "could not establish a connection to the bitcoin node")
	}
}

func unexpectedError(err error) error {
	return sdkerrors.Wrap(types.ErrConnFailed, fmt.Sprintf("unexpected error when waiting for bitcoin node warmup: %s", err.Error()))
}

type connection struct {
	client  *rpcclient.Client
	retries int
	error   error
}

func (c *connection) isAvailable() bool {
	_, c.error = c.client.GetBlockChainInfo()
	return c.error == nil
}

func (c connection) Error() error {
	return c.error
}

func (k Keeper) Close() {
	k.client.Shutdown()
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) TrackAddress(ctx sdk.Context, address string) error {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", address))

	if err := k.client.ImportAddressRescan(address, "axelar", true); err != nil {
		return err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", address))

	return nil
}

package types

import (
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	sleep          = 1 * time.Second
	ErrRpcInWarmup = btcjson.RPCErrorCode(-28)
)

//go:generate moq -pkg mock -out ./mock/rpcClient.go . RPCClient

type RPCClient interface {
	ImportAddress(address string) error
	ImportAddressRescan(address string, account string, rescan bool) error
	GetRawTransactionVerbose(hash *chainhash.Hash) (*btcjson.TxRawResult, error)
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
}

func NewRPCClient(cfg BtcConfig, logger log.Logger) (*rpcclient.Client, error) {
	logger = logger.With("module", fmt.Sprintf("x/%s", ModuleName))

	// Make sure there are authentication parameters
	if cfg.CookiePath != "" {

		if err := waitForAuthCookie(cfg.CookiePath, cfg.StartUpTimeout, logger); err != nil {
			return nil, err
		}

	} else if cfg.RPCUser == "" || cfg.RPCPass == "" {

		return nil, sdkerrors.Wrap(ErrConnFailed, "Authentication method must be specified (either username/password or cookie)")

	}

	rpcCfg := &rpcclient.ConnConfig{
		Host:                 cfg.RPCAddr,
		CookiePath:           cfg.CookiePath,
		User:                 cfg.RPCUser,
		Pass:                 cfg.RPCPass,
		DisableTLS:           true, // Bitcoin core does not provide TLS by default
		DisableAutoReconnect: false,
		HTTPPostMode:         true, // Bitcoin core only supports HTTP POST mode
	}
	// Note the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, sdkerrors.Wrap(ErrConnFailed, "could not start the bitcoin rpc client")
	}

	if err = waitForBtcWarmup(client, cfg.RPCTimeout, logger); err != nil {
		return nil, err
	}
	return client, nil
}

func waitForAuthCookie(cookiePath string, timeout time.Duration, logger log.Logger) error {
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
		return sdkerrors.Wrap(ErrInvalidConfig, fmt.Sprintf("bitcoin auth cookie could not be found at %s", cookiePath))
	}
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
		return sdkerrors.Wrap(ErrTimeOut, "could not establish a connection to the bitcoin node")
	}
}

func unexpectedError(err error) error {
	return sdkerrors.Wrap(ErrConnFailed, fmt.Sprintf("unexpected error when waiting for bitcoin node warmup: %s", err.Error()))
}

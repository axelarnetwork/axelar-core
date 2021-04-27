package rpc

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	sleep          = 1 * time.Second
	errRPCInWarmup = btcjson.RPCErrorCode(-28)
)

//go:generate moq -pkg mock -out ./mock/rpcClient.go . Client

// Client defines the interface of an rpc client communication with the Bitcoin network
type Client interface {
	GetTxOut(txHash *chainhash.Hash, voutIdx uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	Network() types.Network
}

// ClientImpl implements the Client interface
type ClientImpl struct {
	*rpcclient.Client
	Timeout time.Duration
	network types.Network
}

// Network returns the Bitcoin network the client is connected to
func (r *ClientImpl) Network() types.Network {
	return r.network
}

// NewRPCClient creates a new instance of ClientImpl
func NewRPCClient(cfg types.BtcConfig, logger log.Logger) (*ClientImpl, error) {
	logger = logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))

	// Make sure there are authentication parameters
	if cfg.CookiePath != "" {
		if err := waitForAuthCookie(cfg.CookiePath, cfg.StartUpTimeout, logger); err != nil {
			return nil, err
		}
	}
	parsedURL, err := url.Parse(cfg.RPCAddr)
	if err != nil {
		return nil, err
	}
	user := cfg.RPCUser
	pw := cfg.RPCPass
	if parsedURL.User != nil {
		user = parsedURL.User.Username()
		parsedPW, isSet := parsedURL.User.Password()
		if isSet {
			pw = parsedPW
		}
	}

	rpcCfg := &rpcclient.ConnConfig{
		Host:                 parsedURL.Host,
		CookiePath:           cfg.CookiePath,
		User:                 user,
		Pass:                 pw,
		DisableTLS:           true, // Bitcoin core does not provide TLS by default
		DisableAutoReconnect: false,
		HTTPPostMode:         true, // Bitcoin core only supports HTTP POST mode
	}
	// Note the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrConnFailed, "could not start the bitcoin rpc client")
	}
	r := &ClientImpl{Client: client, Timeout: cfg.RPCTimeout}
	if err = r.setNetwork(logger); err != nil {
		return nil, err
	}

	return r, nil
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
	}

	return sdkerrors.Wrap(types.ErrInvalidConfig, fmt.Sprintf("bitcoin auth cookie could not be found at %s", cookiePath))
}

func (r *ClientImpl) setNetwork(logger log.Logger) error {
	maxRetries := int(r.Timeout / sleep)

	var info *btcjson.GetBlockChainInfoResult
	var retries int

	// Ensure the loop is run at least once
	var err error = &btcjson.RPCError{Code: errRPCInWarmup}
	for retries = 0; err != nil && retries < maxRetries; retries++ {
		switch err := err.(type) {
		case *btcjson.RPCError:
			if err.Code == errRPCInWarmup {
				logger.Debug("waiting for bitcoin rpc server to start")
				time.Sleep(sleep)
			} else {
				return unexpectedError(err)
			}
		default:
			return unexpectedError(err)
		}
		info, err = r.GetBlockChainInfo()
	}

	if retries < maxRetries {
		logger.Info("btc bridge client successfully connected to bitcoin node")
		if info == nil {
			return fmt.Errorf("bitcoin blockchain info is nil")
		}
		r.network, err = types.NetworkFromStr(info.Chain)
		return err
	}
	return sdkerrors.Wrap(types.ErrTimeOut, "could not establish a connection to the bitcoin node")
}

func unexpectedError(err error) error {
	return sdkerrors.Wrap(types.ErrConnFailed, fmt.Sprintf("unexpected error when waiting for bitcoin node warmup: %s", err.Error()))
}

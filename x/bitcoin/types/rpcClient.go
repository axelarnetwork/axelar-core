package types

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
)

const (
	sleep          = 1 * time.Second
	errRpcInWarmup = btcjson.RPCErrorCode(-28)
)

//go:generate moq -pkg mock -out ./mock/rpcClient.go . RPCClient

// RPCClient defines the interface of an rpc client communication with the Bitcoin network
type RPCClient interface {
	GetTxOut(txHash *chainhash.Hash, voutIdx uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	Network() Network
}

// RPCClientImpl implements the RPCClient interface
type RPCClientImpl struct {
	*rpcclient.Client
	Timeout time.Duration
	network Network
}

// Network returns the Bitcoin network the client is connected to
func (r *RPCClientImpl) Network() Network {
	return r.network
}

// NewRPCClient creates a new instance of RPCClientImpl
func NewRPCClient(cfg BtcConfig, logger log.Logger) (*RPCClientImpl, error) {
	logger = logger.With("module", fmt.Sprintf("x/%s", ModuleName))

	// Make sure there are authentication parameters
	if cfg.CookiePath != "" {
		if err := waitForAuthCookie(cfg.CookiePath, cfg.StartUpTimeout, logger); err != nil {
			return nil, err
		}
	}
	parsedUrl, err := url.Parse(cfg.RPCAddr)
	if err != nil {
		return nil, err
	}
	user := cfg.RPCUser
	pw := cfg.RPCPass
	if parsedUrl.User != nil {
		user = parsedUrl.User.Username()
		parsedPW, isSet := parsedUrl.User.Password()
		if isSet {
			pw = parsedPW
		}
	}

	rpcCfg := &rpcclient.ConnConfig{
		Host:                 parsedUrl.Host,
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
		return nil, sdkerrors.Wrap(ErrConnFailed, "could not start the bitcoin rpc client")
	}
	r := &RPCClientImpl{Client: client, Timeout: cfg.RPCTimeout}
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
	} else {
		return sdkerrors.Wrap(ErrInvalidConfig, fmt.Sprintf("bitcoin auth cookie could not be found at %s", cookiePath))
	}
}

func (r *RPCClientImpl) setNetwork(logger log.Logger) error {
	maxRetries := int(r.Timeout / sleep)

	var info *btcjson.GetBlockChainInfoResult
	var retries int

	// Ensure the loop is run at least once
	var err error = &btcjson.RPCError{Code: errRpcInWarmup}
	for retries = 0; err != nil && retries < maxRetries; retries++ {
		switch err := err.(type) {
		case *btcjson.RPCError:
			if err.Code == errRpcInWarmup {
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
		r.network, err = NetworkFromStr(info.Chain)
		return err
	} else {
		return sdkerrors.Wrap(ErrTimeOut, "could not establish a connection to the bitcoin node")
	}
}

func unexpectedError(err error) error {
	return sdkerrors.Wrap(ErrConnFailed, fmt.Sprintf("unexpected error when waiting for bitcoin node warmup: %s", err.Error()))
}

type dummyClient struct{}

// NewDummyRPC returns a placeholder for an rpc client. It does not make any rpc calls
func NewDummyRPC() RPCClient {
	return dummyClient{}
}

// GetTxOut implements RPCClient
func (d dummyClient) GetTxOut(*chainhash.Hash, uint32, bool) (*btcjson.GetTxOutResult, error) {
	return &btcjson.GetTxOutResult{}, fmt.Errorf("no response")
}

// SendRawTransaction implements RPCClient
func (d dummyClient) SendRawTransaction(*wire.MsgTx, bool) (*chainhash.Hash, error) {
	return nil, fmt.Errorf("no response")
}

// Network implements RPCClient
func (d dummyClient) Network() Network {
	return DefaultParams().Network
}

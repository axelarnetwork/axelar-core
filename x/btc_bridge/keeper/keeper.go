package keeper

import (
	"fmt"
	"os"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	axTypes "github.com/axelarnetwork/axelar-core/x/axelar/types"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

var (
	_          axTypes.BridgeKeeper = Keeper{}
	confHeight                      = []byte("confHeight")
)

const (
	ErrRpcInWarmup = btcjson.RPCErrorCode(-28)
)

type Keeper struct {
	storeKey sdk.StoreKey
	client   *rpcclient.Client
	cdc      *codec.Codec
}

const (
	sleep   = 1 * time.Second
	Satoshi = 1
	Bitcoin = 100_000_000 * Satoshi
)

func NewBtcKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, cfg types.BtcConfig, logger log.Logger) (Keeper, error) {
	// logger.Debug("initializing btc keeper")
	// client, err := newRPCClient(cfg, logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)))
	// if err != nil {
	// 	return Keeper{}, err
	// }
	// return Keeper{cdc: cdc, storeKey: storeKey, client: client}, nil
	return Keeper{cdc: cdc, storeKey: storeKey}, nil
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

	future := k.client.ImportAddressAsync(address)

	go func() {
		if err := future.Receive(); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not track address %v", address))
		} else {
			k.Logger(ctx).Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", address))
		}

	}()

	return nil
}

func (k Keeper) VerifyTx(ctx sdk.Context, tx exported.ExternalTx) bool {
	k.Logger(ctx).Debug("verifying bitcoin transaction")
	hash, err := chainhash.NewHashFromStr(tx.TxID)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return false
	}

	btcTxResult, err := k.client.GetTransaction(hash)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return false
	}

	verifiedAmount, err := btcutil.NewAmount(btcTxResult.Amount)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return false
	}

	amountEqual := (tx.Amount.Amount.IsInteger() && k.satoshiEquals(tx.Amount.Amount, verifiedAmount)) ||
		k.btcEquals(tx.Amount.Amount, verifiedAmount)

	isEqual := btcTxResult.TxID == tx.TxID && amountEqual && btcTxResult.Confirmations >= 6

	if !isEqual {
		k.Logger(ctx).Debug(fmt.Sprintf(
			"txID:%s\nbtcTxId:%s\ntx amount:%s\nbtc Amount:%v",
			tx.TxID,
			btcTxResult.TxID,
			tx.Amount.String(),
			verifiedAmount,
		))
	}
	return isEqual
}

func (k Keeper) satoshiEquals(satoshiAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return satoshiAmount.IsInt64() && btcutil.Amount(satoshiAmount.Int64()) == verifiedAmount
}

func (k Keeper) btcEquals(btcAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return btcutil.Amount(btcAmount.MulInt64(Bitcoin).Int64()) == verifiedAmount
}

func (k Keeper) SetConfirmationHeight(ctx sdk.Context, height int64) {
	ctx.KVStore(k.storeKey).Set(confHeight, k.cdc.MustMarshalBinaryLengthPrefixed(height))
}

func (k Keeper) GetConfirmationHeight(ctx sdk.Context) int64 {
	rawHeight := ctx.KVStore(k.storeKey).Get(confHeight)
	if rawHeight == nil {
		return types.DefaultGenesisState().ConfirmationHeight
	} else {
		var height int64
		k.cdc.MustUnmarshalBinaryLengthPrefixed(rawHeight, &height)
		return height
	}
}

package btc_bridge

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.Voter = &TestVoter{}

type TestVoter struct {
	vote *exported.FutureVote
}

func (t *TestVoter) SetFutureVote(ctx sdkTypes.Context, vote exported.FutureVote) {
	t.vote = &vote
}

func (t TestVoter) IsVerified(ctx sdkTypes.Context, tx exported.ExternalTx) bool {
	panic("implement me")
}

var _ types.RPCClient = &TestRPC{}

type TestRPC struct {
	trackedAddress string ``
	cancel         context.CancelFunc
	rawTx          func() (*btcjson.TxRawResult, error)
}

func (t *TestRPC) ImportAddressRescan(address string, account string, rescan bool) error {
	t.trackedAddress = address
	t.cancel()
	return nil
}

func (t *TestRPC) ImportAddress(address string) error {
	t.trackedAddress = address
	t.cancel()
	return nil
}

func (t TestRPC) GetRawTransactionVerbose(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	return t.rawTx()
}

func (t TestRPC) SendRawTransaction(tx *wire.MsgTx, b bool) (*chainhash.Hash, error) {
	panic("implement me")
}

var _ types.Signer = TestSigner{}

type TestSigner struct {
}

func (t TestSigner) StartSign(ctx sdkTypes.Context, info tssTypes.MsgSignStart) error {
	panic("implement me")
}

func (t TestSigner) GetSig(ctx sdkTypes.Context, sigID string) (r *big.Int, s *big.Int, err error) {
	panic("implement me")
}

func (t TestSigner) GetKey(ctx sdkTypes.Context, keyID string) (ecdsa.PublicKey, error) {
	panic("implement me")
}

var _ sdkTypes.MultiStore = TestMultiStore{}

func NewMultiStore() sdkTypes.MultiStore {
	return TestMultiStore{kvstore: map[string]sdkTypes.KVStore{}}
}

type TestMultiStore struct {
	kvstore map[string]sdkTypes.KVStore
}

func (t TestMultiStore) GetStoreType() sdkTypes.StoreType {
	panic("implement me")
}

func (t TestMultiStore) CacheWrap() sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestMultiStore) CacheWrapWithTrace(w io.Writer, tc sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStore() sdkTypes.CacheMultiStore {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStoreWithVersion(version int64) (sdkTypes.CacheMultiStore, error) {
	panic("implement me")
}

func (t TestMultiStore) GetStore(key sdkTypes.StoreKey) sdkTypes.Store {
	panic("implement me")
}

func (t TestMultiStore) GetKVStore(key sdkTypes.StoreKey) sdkTypes.KVStore {
	if store, ok := t.kvstore[key.String()]; ok {
		return store
	} else {
		store := NewTestKVStore()
		t.kvstore[key.String()] = store
		return store
	}
}

func (t TestMultiStore) TracingEnabled() bool {
	panic("implement me")
}

func (t TestMultiStore) SetTracer(w io.Writer) sdkTypes.MultiStore {
	panic("implement me")
}

func (t TestMultiStore) SetTracingContext(context sdkTypes.TraceContext) sdkTypes.MultiStore {
	panic("implement me")
}

func NewTestKVStore() sdkTypes.KVStore {
	return TestKVStore{store: map[string][]byte{}}
}

type TestKVStore struct {
	store map[string][]byte
}

func (t TestKVStore) GetStoreType() sdkTypes.StoreType {
	panic("implement me")
}

func (t TestKVStore) CacheWrap() sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) CacheWrapWithTrace(w io.Writer, tc sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) Get(key []byte) []byte {
	val, ok := t.store[string(key)]

	if ok {
		return val
	} else {
		return nil
	}
}

func (t TestKVStore) Has(key []byte) bool {
	_, ok := t.store[string(key)]
	return ok
}

func (t TestKVStore) Set(key, value []byte) {
	t.store[string(key)] = value
}

func (t TestKVStore) Delete(key []byte) {
	delete(t.store, string(key))
}

func (t TestKVStore) Iterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}

func (t TestKVStore) ReverseIterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}

func TestTrackAddress(t *testing.T) {
	cdc := codec.New()
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rpc := TestRPC{cancel: cancel}
	handler := NewHandler(k, &TestVoter{}, &rpc, TestSigner{})

	ctx := sdkTypes.NewContext(NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	expectedAddress, _ := types.ParseBtcAddress("bitcoinTestAddress", "mainnet")
	_, err := handler(ctx, types.MsgTrackAddress{
		Sender:  sdkTypes.AccAddress("sender"),
		Address: expectedAddress,
	})

	assert.Nil(t, err)
	<-rpcCtx.Done()
	assert.Equal(t, expectedAddress.String(), rpc.trackedAddress)
}

func TestVerifyTx_InvalidHash(t *testing.T) {
	cdc := codec.New()

	types.RegisterCodec(cdc)
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpc := TestRPC{
		rawTx: func() (*btcjson.TxRawResult, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	v := &TestVoter{}
	handler := NewHandler(k, v, &rpc, TestSigner{})
	ctx := sdkTypes.NewContext(NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	addr, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	utxo := types.UTXO{
		Hash:    hash,
		VoutIdx: 0,
		Amount:  10,
		Address: addr,
	}

	assert.Nil(t, utxo.Validate())

	_, err := handler(ctx, types.MsgVerifyTx{
		Sender: sdkTypes.AccAddress("sender"),
		UTXO:   utxo,
	})

	assert.Nil(t, err)
	assert.Equal(t, &exported.FutureVote{
		Tx: exported.ExternalTx{
			Chain: "bitcoin",
			TxID:  hash.String(),
		},
		LocalAccept: false,
	}, v.vote)
}

func TestVerifyTx_ValidUTXO(t *testing.T) {
	cdc := codec.New()

	types.RegisterCodec(cdc)
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")

	addr, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	utxo := types.UTXO{
		Hash:    hash,
		VoutIdx: 0,
		Amount:  10,
		Address: addr,
	}
	rpc := TestRPC{
		rawTx: func() (*btcjson.TxRawResult, error) {
			return &btcjson.TxRawResult{
				Txid: hash.String(),
				Hash: hash.String(),
				Vout: []btcjson.Vout{{
					Value: btcutil.Amount(10).ToBTC(),
					N:     0,
					ScriptPubKey: btcjson.ScriptPubKeyResult{
						Addresses: []string{utxo.Address.String()},
					},
				}},
				Confirmations: 7,
			}, nil
		},
	}
	v := &TestVoter{}
	handler := NewHandler(k, v, &rpc, TestSigner{})
	ctx := sdkTypes.NewContext(NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	assert.Nil(t, utxo.Validate())

	_, err := handler(ctx, types.MsgVerifyTx{
		Sender: sdkTypes.AccAddress("sender"),
		UTXO:   utxo,
	})

	assert.Nil(t, err)
	assert.Equal(t, &exported.FutureVote{
		Tx: exported.ExternalTx{
			Chain: "bitcoin",
			TxID:  hash.String(),
		},
		LocalAccept: true,
	}, v.vote)

	actualUtxo, ok := k.GetUTXO(ctx, hash.String())
	assert.True(t, ok)
	assert.True(t, utxo.Equals(actualUtxo))
}

package btc_bridge

import (
	"crypto/ecdsa"
	"io"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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

var _ types.Voter = TestVoter{}

type TestVoter struct {
}

func (t TestVoter) SetFutureVote(ctx sdkTypes.Context, vote exported.FutureVote) {
	panic("implement me")
}

func (t TestVoter) IsVerified(ctx sdkTypes.Context, tx exported.ExternalTx) bool {
	panic("implement me")
}

var _ types.RPCClient = &TestRPC{}

type TestRPC struct {
	trackedAddress string ``
}

func (t *TestRPC) ImportAddress(address string) error {
	t.trackedAddress = address
	return nil
}

func (t TestRPC) GetRawTransactionVerbose(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	panic("implement me")
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

func (t TestSigner) GetSig(ctx sdkTypes.Context, sigID string) (r *big.Int, s *big.Int) {
	panic("implement me")
}

func (t TestSigner) GetKey(ctx sdkTypes.Context, keyID string) ecdsa.PublicKey {
	panic("implement me")
}

var _ sdkTypes.MultiStore = TestMultiStore{}

func NewMultiStore() sdkTypes.MultiStore {
	return TestMultiStore{kvstore: NewTestKVStore()}
}

type TestMultiStore struct {
	kvstore sdkTypes.KVStore
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
	return t.kvstore
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

var _ sdkTypes.KVStore = TestKVStore{}

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
	panic("implement me")
}

func (t TestKVStore) Iterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}

func (t TestKVStore) ReverseIterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}

func TestTrackAddress(t *testing.T) {
	k := keeper.NewBtcKeeper(codec.New(), sdkTypes.NewKVStoreKey("testKey"))
	rpc := TestRPC{}
	handler := NewHandler(k, TestVoter{}, &rpc, TestSigner{})
	ctx := sdkTypes.NewContext(NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	expectedAddress := "bitcoinTestAddress"
	_, err := handler(ctx, types.MsgTrackAddress{
		Sender:  sdkTypes.AccAddress("sender"),
		Address: expectedAddress,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedAddress, rpc.trackedAddress)
}

// func TestVerifyInvalidTx(t *testing.T){
// 	k := keeper.NewBtcKeeper(codec.New(), sdkTypes.NewKVStoreKey("testKey"))
// 	rpc := TestRPC{}
// 	handler := NewHandler(k, TestVoter{}, &rpc, TestSigner{})
// 	ctx := sdkTypes.NewContext(TestMultiStore{}, abci.Header{}, false, log.TestingLogger())
// }

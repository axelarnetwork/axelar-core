package tests

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/tests/mock"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

func TestTrackAddress(t *testing.T) {
	cdc := codec.New()
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rpc := btcMock.TestRPC{Cancel: cancel}
	handler := bitcoin.NewHandler(k, &btcMock.TestVoter{}, &rpc, nil)

	ctx := sdkTypes.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	expectedAddress, _ := types.ParseBtcAddress("bitcoinTestAddress", "mainnet")
	_, err := handler(ctx, types.MsgTrackAddress{
		Sender:  sdkTypes.AccAddress("sender"),
		Address: expectedAddress,
	})

	assert.Nil(t, err)
	<-rpcCtx.Done()
	assert.Equal(t, expectedAddress.String(), rpc.TrackedAddress)
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := codec.New()

	types.RegisterCodec(cdc)
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpc := btcMock.TestRPC{
		RawTxs: map[string]*btcjson.TxRawResult{},
	}
	v := &btcMock.TestVoter{}
	handler := bitcoin.NewHandler(k, v, &rpc, nil)
	ctx := sdkTypes.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

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
	assert.True(t, v.InitPollCalled)
	assert.True(t, v.VoteCalledCorrectly)
	assert.False(t, v.RecordedVote.Data().(bool))
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
	rpc := btcMock.TestRPC{
		RawTxs: map[string]*btcjson.TxRawResult{
			hash.String(): {
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
			},
		},
	}
	v := &btcMock.TestVoter{}
	handler := bitcoin.NewHandler(k, v, &rpc, nil)
	ctx := sdkTypes.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	assert.Nil(t, utxo.Validate())

	_, err := handler(ctx, types.MsgVerifyTx{
		Sender: sdkTypes.AccAddress("sender"),
		UTXO:   utxo,
	})

	assert.Nil(t, err)
	assert.True(t, v.InitPollCalled)
	assert.True(t, v.VoteCalledCorrectly)
	assert.True(t, v.RecordedVote.Data().(bool))

	actualUtxo, ok := k.GetUTXO(ctx, hash.String())
	assert.True(t, ok)
	assert.True(t, utxo.Equals(actualUtxo))
}

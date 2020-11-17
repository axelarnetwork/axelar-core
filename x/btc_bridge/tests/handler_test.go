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

	"github.com/axelarnetwork/axelar-core/test-utils/mock"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	btcMock "github.com/axelarnetwork/axelar-core/x/btc_bridge/tests/mock"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

func TestTrackAddress(t *testing.T) {
	cdc := codec.New()
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rpc := btcMock.TestRPC{Cancel: cancel}
	handler := btc_bridge.NewHandler(k, &btcMock.TestVoter{}, &rpc, btcMock.TestSigner{})

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

func TestVerifyTx_InvalidHash(t *testing.T) {
	cdc := codec.New()

	types.RegisterCodec(cdc)
	k := keeper.NewBtcKeeper(cdc, sdkTypes.NewKVStoreKey("testKey"))
	rpc := btcMock.TestRPC{
		RawTxs: map[string]*btcjson.TxRawResult{},
	}
	v := &btcMock.TestVoter{}
	handler := btc_bridge.NewHandler(k, v, &rpc, btcMock.TestSigner{})
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
	assert.Equal(t, &exported.FutureVote{
		Tx: exported.ExternalTx{
			Chain: "bitcoin",
			TxID:  hash.String(),
		},
		LocalAccept: false,
	}, v.Vote)
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
	handler := btc_bridge.NewHandler(k, v, &rpc, btcMock.TestSigner{})
	ctx := sdkTypes.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

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
	}, v.Vote)

	actualUtxo, ok := k.GetUTXO(ctx, hash.String())
	assert.True(t, ok)
	assert.True(t, utxo.Equals(actualUtxo))
}

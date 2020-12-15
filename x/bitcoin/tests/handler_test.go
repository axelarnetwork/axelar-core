package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/tests/mock"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const testReps = 100

func TestTrackAddress(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	rpcCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	rpc := btcMock.TestRPC{Cancel: cancel}
	handler := bitcoin.NewHandler(k, &btcMock.TestVoter{}, &rpc, nil)

	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	expectedAddress, _ := types.ParseBtcAddress("bitcoinTestAddress", "mainnet")
	_, err := handler(ctx, types.MsgTrackAddress{
		Sender:  sdk.AccAddress("sender"),
		Address: expectedAddress,
	})

	assert.Nil(t, err)
	<-rpcCtx.Done()
	assert.Equal(t, expectedAddress.String(), rpc.TrackedAddress)
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	rpc := btcMock.TestRPC{
		RawTxs: map[string]*btcjson.TxRawResult{},
	}
	v := &btcMock.TestVoter{}
	handler := bitcoin.NewHandler(k, v, &rpc, nil)
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

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
		Sender: sdk.AccAddress("sender"),
		UTXO:   utxo,
	})
	assert.Nil(t, err)
	assert.True(t, v.InitPollCalled)
	assert.True(t, v.VoteCalledCorrectly)
	assert.False(t, v.RecordedVote.Data().(bool))
}

func TestVerifyTx_ValidUTXO(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))

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
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	assert.Nil(t, utxo.Validate())

	_, err := handler(ctx, types.MsgVerifyTx{
		Sender: sdk.AccAddress("sender"),
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

func TestMasterKey_RawTx_Then_Transfer(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, mock.NewKVStoreKey(types.StoreKey))

	var sk, skNext *ecdsa.PrivateKey
	var txHash []byte
	var txID, sigID string
	signer := btcMock.TestSigner{
		GetCurrentMasterKeyMock: func(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool) {
			return sk.PublicKey, true
		},
		GetNextMasterKeyMock: func(ctx sdk.Context, chain string) (ecdsa.PublicKey, bool) {
			return skNext.PublicKey, true
		},
		GetSigMock: func(ctx sdk.Context, sID string) (tss.Signature, bool) {
			if sID == sigID {
				r, s, err := ecdsa.Sign(rand.Reader, sk, txHash)
				if err != nil {
					panic(err)
				}
				return tss.Signature{R: r, S: s}, true
			}
			return tss.Signature{}, false
		},
	}

	v := &btcMock.TestVoter{ResultMock: func(s sdk.Context, pollMeta exported.PollMeta) exported.VotingData {
		return pollMeta.ID == txID
	}}

	handler := bitcoin.NewHandler(k, v, &btcMock.TestRPC{}, signer)

	for i := 0; i < testReps; i++ {
		ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		sk, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		skNext, _ = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		sigID = testutils.RandString(int(testutils.RandIntBetween(5, 20)))

		rawTx, transfer := prepareMsgTransferToNewMasterKey(ctx, k, sk, sigID)
		txID = transfer.TxID

		res, err := handler(ctx, rawTx)
		assert.NoError(t, err)
		txHash = res.Data

		_, err = handler(ctx, transfer)
		assert.NoError(t, err)
	}
}

func prepareMsgTransferToNewMasterKey(ctx sdk.Context, k keeper.Keeper, sk *ecdsa.PrivateKey, sigID string) (types.MsgRawTxForMasterKey, types.MsgTransferToNewMasterKey) {
	hash, err := chainhash.NewHash([]byte(testutils.RandString(chainhash.HashSize)))
	if err != nil {
		panic(err)
	}

	txId := hash.String()
	btcPk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKey(btcPk.SerializeUncompressed(), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	amount := btcutil.Amount(testutils.RandIntBetween(1, 100000000))
	k.SetUTXO(ctx, txId, types.UTXO{
		Hash:    hash,
		VoutIdx: uint32(testutils.RandIntBetween(0, 10)),
		Amount:  amount,
		Address: types.BtcAddress{Chain: types.Chain(chaincfg.MainNetParams.Name), EncodedString: addr.EncodeAddress()},
	})

	sender := sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 50))))

	rawTx := types.MsgRawTxForMasterKey{
		Sender: sender,
		TxHash: hash,
		Amount: amount,
		Chain:  types.Chain(chaincfg.MainNetParams.Name),
	}

	transfer := types.MsgTransferToNewMasterKey{
		Sender:      sender,
		TxID:        txId,
		SignatureID: sigID,
	}
	return rawTx, transfer
}

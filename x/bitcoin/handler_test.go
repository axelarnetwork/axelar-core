package bitcoin

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	btcMock "github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const testReps = 100

func TestTrackAddress(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	rpc := btcMock.RPCClientMock{ImportAddressRescanFunc: func(address string, _ string, rescan bool) error {
		cancel()
		return nil
	}}

	sk, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	pk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	expectedAddress, err := types.ParseBtcAddress(addr.EncodeAddress(), types.Network(chaincfg.MainNetParams.Name))
	if err != nil {
		panic(err)
	}

	handler := NewHandler(k, &btcMock.VoterMock{}, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgTrackAddress(sdk.AccAddress("sender"), expectedAddress, false))

	<-timeout.Done()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rpc.ImportAddressRescanCalls()))
	assert.False(t, rpc.ImportAddressRescanCalls()[0].Rescan)
	assert.Equal(t, expectedAddress.String(), rpc.ImportAddressRescanCalls()[0].Address)
}

func TestVerifyTx_InvalidHash_VoteDiscard(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	rpc := btcMock.RPCClientMock{
		GetRawTransactionVerboseFunc: func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	var poll exported.PollMeta
	v := &btcMock.VoterMock{
		InitPollFunc:   func(_ sdk.Context, p exported.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote exported.MsgVote) error { return nil },
	}

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	sender, _ := types.ParseBtcAddress("1NtzRAvBRZsUjG5i5MxNt9xwaUByri9Rg2", "mainnet")
	recipient, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	utxo := types.UTXO{
		Hash:      hash,
		VoutIdx:   0,
		Amount:    10,
		Recipient: recipient,
		Sender:    sender,
	}
	if err := utxo.Validate(); err != nil {
		panic(err)
	}

	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	_, err := handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), UTXO: utxo})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, hash.String(), v.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), v.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, v.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(v.RecordVoteCalls()))
	assert.Equal(t, poll, v.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, false, v.RecordVoteCalls()[0].Vote.Data())
}

func TestVerifyTx_ValidUTXO(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey("testKey"))

	hash, _ := chainhash.NewHashFromStr("f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	recipient, _ := types.ParseBtcAddress("bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", "mainnet")
	sender, _ := types.ParseBtcAddress("1NtzRAvBRZsUjG5i5MxNt9xwaUByri9Rg2", "mainnet")
	spentHash, _ := chainhash.NewHashFromStr("74d39e87c810a80faff70dcbd988c661dbe283a27f903cd587ab9c0b221cc602")
	utxo := types.UTXO{
		Hash:      hash,
		VoutIdx:   0,
		Amount:    10,
		Recipient: recipient,
		Sender:    sender,
	}
	if err := utxo.Validate(); err != nil {
		panic(err)
	}

	rpc := btcMock.RPCClientMock{
		GetRawTransactionVerboseFunc: func(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
			if hash.IsEqual(utxo.Hash) {
				return &btcjson.TxRawResult{
					Txid: hash.String(),
					Hash: hash.String(),
					Vin: []btcjson.Vin{{
						Txid: spentHash.String(),
						Vout: 0,
					}},
					Vout: []btcjson.Vout{{
						Value: btcutil.Amount(10).ToBTC(),
						N:     0,
						ScriptPubKey: btcjson.ScriptPubKeyResult{
							Addresses: []string{utxo.Recipient.String()},
						},
					}},
					Confirmations: 7,
				}, nil
			}
			if hash.IsEqual(spentHash) {
				return &btcjson.TxRawResult{
					Txid: hash.String(),
					Hash: hash.String(),
					Vout: []btcjson.Vout{{
						N: 0,
						ScriptPubKey: btcjson.ScriptPubKeyResult{
							Addresses: []string{sender.String()},
						},
					}},
					Confirmations: 20,
				}, nil
			}

			return nil, fmt.Errorf("not found")
		},
	}

	var poll exported.PollMeta
	v := &btcMock.VoterMock{
		InitPollFunc: func(_ sdk.Context, p exported.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote exported.MsgVote) error {
			return nil
		},
	}
	handler := NewHandler(k, v, &rpc, &btcMock.SignerMock{}, &btcMock.BalancerMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	_, err := handler(ctx, types.MsgVerifyTx{Sender: sdk.AccAddress("sender"), UTXO: utxo})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(v.InitPollCalls()))
	assert.Equal(t, hash.String(), v.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), v.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, v.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(v.RecordVoteCalls()))
	assert.Equal(t, poll, v.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, true, v.RecordVoteCalls()[0].Vote.Data())

	actualUtxo, ok := k.GetUTXOForPoll(ctx, hash.String())
	assert.True(t, ok)
	assert.True(t, utxo.Equals(actualUtxo))
}

func TestMasterKey_RawTx_Then_Transfer(t *testing.T) {
	cdc := testutils.Codec()
	k := keeper.NewBtcKeeper(cdc, sdk.NewKVStoreKey(types.StoreKey))

	var sk, skNext *ecdsa.PrivateKey
	var txHash []byte
	var txID, sigID string
	signer := &btcMock.SignerMock{
		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return sk.PublicKey, true
		},
		GetNextMasterKeyFunc: func(ctx sdk.Context, chain balance.Chain) (ecdsa.PublicKey, bool) {
			return skNext.PublicKey, true
		},
		GetSigFunc: func(ctx sdk.Context, sID string) (tss.Signature, bool) {
			if sID == sigID {
				r, s, err := ecdsa.Sign(rand.Reader, sk, txHash)
				if err != nil {
					panic(err)
				}
				return tss.Signature{R: r, S: s}, true
			}
			return tss.Signature{}, false
		},
		GetKeyForSigIDFunc: func(ctx sdk.Context, sigID string) (ecdsa.PublicKey, bool) {
			return sk.PublicKey, true
		},
	}
	v := &btcMock.VoterMock{ResultFunc: func(s sdk.Context, pollMeta exported.PollMeta) exported.VotingData {
		return pollMeta.ID == txID
	}}
	rpc := &btcMock.RPCClientMock{SendRawTransactionFunc: func(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
		hash := tx.TxHash()
		return &hash, nil
	}}
	b := &btcMock.BalancerMock{GetRecipientFunc: func(ctx sdk.Context, sender balance.CrossChainAddress) (balance.CrossChainAddress, bool) {
		return balance.CrossChainAddress{}, false
	}}
	handler := NewHandler(k, v, rpc, signer, b)

	for i := 0; i < testReps; i++ {
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
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

		assert.Equal(t, i+1, len(signer.GetKeyForSigIDCalls()))
		assert.Equal(t, sigID, signer.GetKeyForSigIDCalls()[i].SigID)
	}
}

func prepareMsgTransferToNewMasterKey(ctx sdk.Context, k keeper.Keeper, sk *ecdsa.PrivateKey, sigID string) (types.MsgRawTx, types.MsgSendTx) {
	hash, err := chainhash.NewHash([]byte(testutils.RandString(chainhash.HashSize)))
	if err != nil {
		panic(err)
	}

	txId := hash.String()
	btcPk := btcec.PublicKey(sk.PublicKey)
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(btcPk.SerializeCompressed()), &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	amount := btcutil.Amount(testutils.RandIntBetween(1, 100000000))

	k.SetUTXOForPoll(ctx, txId, types.UTXO{
		Hash:      hash,
		VoutIdx:   uint32(testutils.RandIntBetween(0, 10)),
		Amount:    amount,
		Recipient: types.BtcAddress{Network: types.Network(chaincfg.MainNetParams.Name), EncodedString: addr.EncodeAddress()},
	})
	_ = k.ProcessUTXOPollResult(ctx, txId, true)

	sender := sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 50))))

	rawTx := types.NewMsgRawTxForNextMasterKey(sender, types.Network(chaincfg.MainNetParams.Name), hash, amount)

	transfer := types.NewMsgSendTx(sender, txId, sigID)
	return rawTx, transfer
}
